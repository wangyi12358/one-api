package controller

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/service"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login
// @Summary 用户登录
// @Description 用户名密码登录，有频率限制
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求"
// @Success 200 {object} model.User "成功"
// @Router /api/user/login [post]
func Login(c *gin.Context) {
	if !config.PasswordLoginEnabled {
		common.ApiErrorMsg(c, i18n.Translate(c, "password_login_disabled"))
		return
	}
	var loginRequest LoginRequest
	err := json.NewDecoder(c.Request.Body).Decode(&loginRequest)
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}
	username := loginRequest.Username
	password := loginRequest.Password
	if username == "" || password == "" {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}
	user := model.User{
		Username: username,
		Password: password,
	}
	err = user.ValidateAndFill()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	SetupLogin(&user, c)
}

// setup session & cookies and then return user info
func SetupLogin(user *model.User, c *gin.Context) {
	session := sessions.Default(c)
	session.Set("id", user.Id)
	session.Set("username", user.Username)
	session.Set("role", user.Role)
	session.Set("status", user.Status)
	err := session.Save()
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "session_save_failed"))
		return
	}
	cleanUser := model.User{
		Id:          user.Id,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Role:        user.Role,
		Status:      user.Status,
	}
	common.ApiSuccess(c, cleanUser)
}

// Logout
// @Summary 用户登出
// @Description 清除用户会话
// @Tags 用户认证
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/logout [get]
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// Register
// @Summary 用户注册
// @Description 新用户注册，有频率限制
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/register [post]
func Register(c *gin.Context) {
	ctx := c.Request.Context()

	var req service.RegisterRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}

	// 参数验证
	if req.Username == "" || req.Password == "" {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}

	userService := service.GetUserService()
	err = userService.Register(ctx, &req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

// GetAllUsers
// @Summary 获取所有用户
// @Description 获取用户列表，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Param order query string false "排序方式" Enums(quota, used_quota, request_count)
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/ [get]
func GetAllUsers(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.DefaultQuery("order", "")
	users, err := model.GetAllUsers(p*config.ItemsPerPage, config.ItemsPerPage, order)

	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, users)
}

// SearchUsers
// @Summary 搜索用户
// @Description 根据关键词搜索用户，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/search [get]
func SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	users, err := model.SearchUsers(keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, users)
}

// GetUser
// @Summary 获取指定用户
// @Description 获取用户详情，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "用户 ID"
// @Success 200 {object} model.User "成功"
// @Router /api/user/{id} [get]
func GetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	user, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	myRole := c.GetInt(ctxkey.Role)
	if myRole <= user.Role && myRole != model.RoleRootUser {
		common.ApiErrorMsg(c, i18n.Translate(c, "no_permission_same_role"))
		return
	}
	common.ApiSuccess(c, user)
}

// GetUserDashboard
// @Summary 获取用户仪表盘
// @Description 获取用户最近7天的使用统计
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/dashboard [get]
func GetUserDashboard(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	now := time.Now()
	startOfDay := now.Truncate(24*time.Hour).AddDate(0, 0, -6).Unix()
	endOfDay := now.Truncate(24 * time.Hour).Add(24*time.Hour - time.Second).Unix()

	dashboards, err := model.SearchLogsByDayAndModel(id, int(startOfDay), int(endOfDay))
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "dashboard_fetch_failed"))
		return
	}
	common.ApiSuccess(c, dashboards)
}

// GenerateAccessToken
// @Summary 生成访问令牌
// @Description 生成用户的访问令牌
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/token [get]
func GenerateAccessToken(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	userService := service.GetUserService()
	accessToken, err := userService.GenerateAccessToken(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, accessToken)
}

// GetAffCode
// @Summary 获取邀请码
// @Description 获取用户的邀请码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/aff [get]
func GetAffCode(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, true)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.AffCode == "" {
		user.AffCode = random.GetRandomString(4)
		if err := user.Update(false); err != nil {
			common.ApiError(c, err)
			return
		}
	}
	common.ApiSuccess(c, user.AffCode)
}

// GetSelf
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} model.User "成功"
// @Router /api/user/self [get]
func GetSelf(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, user)
}

// UpdateUser
// @Summary 更新用户
// @Description 更新用户信息，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.UserUpdateParams true "更新请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/ [put]
func UpdateUser(c *gin.Context) {
	var params service.UserUpdateParams
	err := json.NewDecoder(c.Request.Body).Decode(&params)
	if err != nil || params.Id == 0 {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}

	if params.Password == "" {
		params.Password = "$I_LOVE_U" // make Validator happy :)
	}

	// 验证用户结构
	userForValidation := model.User{
		Id:          params.Id,
		Username:    params.Username,
		Password:    params.Password,
		DisplayName: params.Username,
	}
	if err := common.Validate.Struct(&userForValidation); err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_input"))
		return
	}

	if params.Password == "$I_LOVE_U" {
		params.Password = "" // rollback to what it should be
	}

	myRole := c.GetInt(ctxkey.Role)
	myId := c.GetInt(ctxkey.Id)
	userService := service.GetUserService()
	err = userService.UpdateUser(&params, myId, myRole)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

// UpdateSelf
// @Summary 更新当前用户信息
// @Description 更新当前登录用户的信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body model.User true "更新请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/self [put]
func UpdateSelf(c *gin.Context) {
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}
	if user.Password == "" {
		user.Password = "$I_LOVE_U" // make Validator happy :)
	}
	if err := common.Validate.Struct(&user); err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_input_with_error")+" "+err.Error())
		return
	}

	cleanUser := model.User{
		Id:          c.GetInt(ctxkey.Id),
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if user.Password == "$I_LOVE_U" {
		user.Password = "" // rollback to what it should be
		cleanUser.Password = ""
	}
	updatePassword := user.Password != ""
	if err := cleanUser.Update(updatePassword); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

// DeleteUser
// @Summary 删除用户
// @Description 删除指定用户，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "用户 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/{id} [delete]
func DeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	myRole := c.GetInt("role")
	userService := service.GetUserService()
	err = userService.DeleteUser(id, myRole)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

// DeleteSelf
// @Summary 注销当前用户
// @Description 删除当前登录用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/self [delete]
func DeleteSelf(c *gin.Context) {
	id := c.GetInt("id")
	userService := service.GetUserService()
	err := userService.DeleteSelf(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

// CreateUser
// @Summary 创建用户
// @Description 创建新用户，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body model.User true "创建请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/ [post]
func CreateUser(c *gin.Context) {
	ctx := c.Request.Context()
	var user model.User
	err := json.NewDecoder(c.Request.Body).Decode(&user)
	if err != nil || user.Username == "" || user.Password == "" {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}
	if err := common.Validate.Struct(&user); err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_input"))
		return
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
	myRole := c.GetInt("role")
	if user.Role >= myRole {
		common.ApiErrorMsg(c, i18n.Translate(c, "cannot_create_higher_role"))
		return
	}
	// Even for admin users, we cannot fully trust them!
	cleanUser := model.User{
		Username:    user.Username,
		Password:    user.Password,
		DisplayName: user.DisplayName,
	}
	if err := cleanUser.Insert(ctx, 0); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

type ManageRequest struct {
	Username string `json:"username"`
	Action   string `json:"action"`
}

// ManageUser Only admin user can do this
// ManageUser
// @Summary 管理用户
// @Description 管理员操作用户，如启用/禁用、设置额度等
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.ManageUserRequest true "管理请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/manage [post]
func ManageUser(c *gin.Context) {
	var req service.ManageUserRequest
	err := json.NewDecoder(c.Request.Body).Decode(&req)
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}

	myRole := c.GetInt("role")
	userService := service.GetUserService()
	user, err := userService.ManageUser(&req, myRole)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	clearUser := model.User{
		Role:   user.Role,
		Status: user.Status,
	}
	common.ApiSuccess(c, clearUser)
}

// EmailBind
// @Summary 绑定邮箱
// @Description 绑定用户邮箱
// @Tags OAuth
// @Accept json
// @Produce json
// @Security UserAuth
// @Param email query string true "邮箱地址"
// @Param code query string true "验证码"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/oauth/email/bind [get]
func EmailBind(c *gin.Context) {
	email := c.Query("email")
	code := c.Query("code")

	id := c.GetInt("id")
	userService := service.GetUserService()
	err := userService.EmailBind(id, email, code)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

type topUpRequest struct {
	Key string `json:"key"`
}

// TopUp
// @Summary 用户充值
// @Description 使用兑换码充值
// @Tags 充值
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body topUpRequest true "充值请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/user/topup [post]
func TopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := topUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	id := c.GetInt("id")
	userService := service.GetUserService()
	quota, err := userService.TopUp(ctx, id, req.Key)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, quota)
}

type adminTopUpRequest struct {
	UserId int    `json:"user_id"`
	Quota  int    `json:"quota"`
	Remark string `json:"remark"`
}

// AdminTopUp
// @Summary 管理员充值
// @Description 管理员为用户充值额度
// @Tags 充值
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body adminTopUpRequest true "充值请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/topup [post]
func AdminTopUp(c *gin.Context) {
	ctx := c.Request.Context()
	req := adminTopUpRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	userService := service.GetUserService()
	err = userService.AdminTopUp(ctx, req.UserId, req.Quota, req.Remark)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}
