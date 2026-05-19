package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrRegisterDisabled          = errors.New("管理员关闭了新用户注册")
	ErrPasswordRegisterDisabled  = errors.New("管理员关闭了通过密码进行注册，请使用第三方账户验证的形式进行注册")
	ErrEmailVerificationRequired = errors.New("管理员开启了邮箱验证，请输入邮箱地址和验证码")
	ErrVerificationCodeInvalid   = errors.New("验证码错误或已过期")
	ErrUserNotFound              = errors.New("用户不存在")
	ErrNoPermission              = errors.New("无权执行此操作")
	ErrCannotDisableRootUser     = errors.New("无法禁用超级管理员用户")
	ErrCannotDeleteRootUser      = errors.New("无法删除超级管理员用户")
	ErrCannotDemoteRootUser      = errors.New("无法降级超级管理员用户")
	ErrUserAlreadyAdmin          = errors.New("该用户已经是管理员")
	ErrUserAlreadyCommon         = errors.New("该用户已经是普通用户")
	ErrCannotPromoteByNonRoot    = errors.New("普通管理员用户无法提升其他用户为管理员")
	ErrCannotCreateHigherRole    = errors.New("无法创建权限大于等于自己的用户")
	ErrAccessTokenDuplicate      = errors.New("请重试，系统生成的 UUID 竟然重复了")
)

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	DisplayName      string `json:"display_name"`
	Email            string `json:"email"`
	VerificationCode string `json:"verification_code"`
	AffCode          string `json:"aff_code"`
}

// ManageAction 管理操作类型
type ManageAction string

const (
	ManageActionDisable ManageAction = "disable"
	ManageActionEnable  ManageAction = "enable"
	ManageActionDelete  ManageAction = "delete"
	ManageActionPromote ManageAction = "promote"
	ManageActionDemote  ManageAction = "demote"
)

// ManageUserRequest 管理用户请求参数
type ManageUserRequest struct {
	Username string       `json:"username"`
	Action   ManageAction `json:"action"`
}

// UserUpdateParams 用户更新参数
type UserUpdateParams struct {
	Id          int    `json:"id"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Role        int    `json:"role"`
	Quota       int64  `json:"quota"`
}

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务实例
func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) error {
	// 1. 检查注册功能是否启用
	if !config.RegisterEnabled {
		return ErrRegisterDisabled
	}
	if !config.PasswordRegisterEnabled {
		return ErrPasswordRegisterDisabled
	}

	// 2. 邮箱验证
	if config.EmailVerificationEnabled {
		if req.Email == "" || req.VerificationCode == "" {
			return ErrEmailVerificationRequired
		}
		if !common.VerifyCodeWithKey(req.Email, req.VerificationCode, common.EmailVerificationPurpose) {
			return ErrVerificationCodeInvalid
		}
	}

	// 3. 查询邀请人
	inviterId, _ := model.GetUserIdByAffCode(req.AffCode)

	// 4. 创建用户
	user := &model.User{
		Username:    req.Username,
		Password:    req.Password,
		DisplayName: req.Username,
		InviterId:   inviterId,
	}
	if config.EmailVerificationEnabled {
		user.Email = req.Email
	}

	// 5. 执行注册（包含密码加密、初始额度、邀请奖励、默认token创建）
	return s.createUser(ctx, user, inviterId)
}

// createUser 创建用户（内部方法，包含完整业务逻辑）
func (s *UserService) createUser(ctx context.Context, user *model.User, inviterId int) error {
	// 密码加密
	if user.Password != "" {
		hashedPassword, err := common.Password2Hash(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	// 设置初始值
	user.Quota = config.QuotaForNewUser
	user.AccessToken = random.GetUUID()
	user.AffCode = random.GetRandomString(4)

	// 插入数据库
	if err := user.Insert(ctx, inviterId); err != nil {
		return err
	}

	// 记录日志：新用户赠送额度
	if config.QuotaForNewUser > 0 {
		model.RecordLog(ctx, user.Id, model.LogTypeSystem, fmt.Sprintf("新用户注册赠送 %s", common.LogQuota(config.QuotaForNewUser)))
	}

	// 处理邀请奖励
	if inviterId != 0 {
		s.handleInviteReward(ctx, user.Id, inviterId)
	}

	// 创建默认令牌
	s.createDefaultToken(user.Id)

	return nil
}

// handleInviteReward 处理邀请奖励
func (s *UserService) handleInviteReward(ctx context.Context, userId, inviterId int) {
	if config.QuotaForInvitee > 0 {
		_ = model.IncreaseUserQuota(userId, config.QuotaForInvitee)
		model.RecordLog(ctx, userId, model.LogTypeSystem, fmt.Sprintf("使用邀请码赠送 %s", common.LogQuota(config.QuotaForInvitee)))
	}
	if config.QuotaForInviter > 0 {
		_ = model.IncreaseUserQuota(inviterId, config.QuotaForInviter)
		model.RecordLog(ctx, inviterId, model.LogTypeSystem, fmt.Sprintf("邀请用户赠送 %s", common.LogQuota(config.QuotaForInviter)))
	}
}

// createDefaultToken 创建默认令牌
func (s *UserService) createDefaultToken(userId int) {
	cleanToken := model.Token{
		UserId:         userId,
		Name:           "default",
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    -1,
		RemainQuota:    -1,
		UnlimitedQuota: true,
	}
	if err := cleanToken.Insert(); err != nil {
		logger.SysError(fmt.Sprintf("create default token for user %d failed: %s", userId, err.Error()))
	}
}

// ManageUser 管理用户（禁用/启用/删除/提升/降级）
func (s *UserService) ManageUser(req *ManageUserRequest, operatorRole int) (*model.User, error) {
	// 1. 查询用户
	user := &model.User{Username: req.Username}
	model.DB.Where(user).First(user)
	if user.Id == 0 {
		return nil, ErrUserNotFound
	}

	// 2. 权限校验：不能操作同级或更高级用户
	if operatorRole <= user.Role && operatorRole != model.RoleRootUser {
		return nil, ErrNoPermission
	}

	// 3. 执行操作
	switch req.Action {
	case ManageActionDisable:
		if user.Role == model.RoleRootUser {
			return nil, ErrCannotDisableRootUser
		}
		user.Status = model.UserStatusDisabled

	case ManageActionEnable:
		user.Status = model.UserStatusEnabled

	case ManageActionDelete:
		if user.Role == model.RoleRootUser {
			return nil, ErrCannotDeleteRootUser
		}
		if err := user.Delete(); err != nil {
			return nil, err
		}
		return user, nil

	case ManageActionPromote:
		if operatorRole != model.RoleRootUser {
			return nil, ErrCannotPromoteByNonRoot
		}
		if user.Role >= model.RoleAdminUser {
			return nil, ErrUserAlreadyAdmin
		}
		user.Role = model.RoleAdminUser

	case ManageActionDemote:
		if user.Role == model.RoleRootUser {
			return nil, ErrCannotDemoteRootUser
		}
		if user.Role == model.RoleCommonUser {
			return nil, ErrUserAlreadyCommon
		}
		user.Role = model.RoleCommonUser

	default:
		return nil, fmt.Errorf("未知的操作类型: %s", req.Action)
	}

	// 4. 保存更新
	if err := user.Update(false); err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(params *UserUpdateParams, operatorId int, operatorRole int) error {
	// 1. 查询原始用户
	originUser, err := model.GetUserById(params.Id, false)
	if err != nil {
		return err
	}

	// 2. 权限校验
	if operatorRole <= originUser.Role && operatorRole != model.RoleRootUser {
		return ErrNoPermission
	}
	if operatorRole <= params.Role && operatorRole != model.RoleRootUser {
		return ErrNoPermission
	}

	// 3. 构建更新对象
	updatedUser := &model.User{
		Id:          params.Id,
		Username:    params.Username,
		Password:    params.Password,
		DisplayName: params.DisplayName,
		Role:        params.Role,
		Quota:       params.Quota,
	}

	// 4. 处理密码
	updatePassword := params.Password != ""
	if updatePassword {
		hashedPassword, err := common.Password2Hash(params.Password)
		if err != nil {
			return err
		}
		updatedUser.Password = hashedPassword
	}

	// 5. 执行更新
	if err := updatedUser.Update(updatePassword); err != nil {
		return err
	}

	// 6. 记录额度变更日志
	if originUser.Quota != params.Quota {
		ctx := context.Background()
		model.RecordLog(ctx, originUser.Id, model.LogTypeManage, fmt.Sprintf("管理员将用户额度从 %s修改为 %s", common.LogQuota(originUser.Quota), common.LogQuota(params.Quota)))
	}

	return nil
}

// GenerateAccessToken 生成访问令牌
func (s *UserService) GenerateAccessToken(userId int) (string, error) {
	user, err := model.GetUserById(userId, true)
	if err != nil {
		return "", err
	}

	user.AccessToken = random.GetUUID()

	// 检查是否重复（通过 model 层方法，而非直接访问 DB）
	existingUser := &model.User{}
	if err := model.DB.Where("access_token = ?", user.AccessToken).First(existingUser).Error; err == nil {
		return "", ErrAccessTokenDuplicate
	}

	if err := user.Update(false); err != nil {
		return "", err
	}

	return user.AccessToken, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(userId int, operatorRole int) error {
	user, err := model.GetUserById(userId, false)
	if err != nil {
		return err
	}

	if operatorRole <= user.Role {
		return ErrNoPermission
	}

	return model.DeleteUserById(userId)
}

// DeleteSelf 用户自行删除账户
func (s *UserService) DeleteSelf(userId int) error {
	user, err := model.GetUserById(userId, false)
	if err != nil {
		return err
	}

	if user.Role == model.RoleRootUser {
		return ErrCannotDeleteRootUser
	}

	return model.DeleteUserById(userId)
}

// EmailBind 绑定邮箱
func (s *UserService) EmailBind(userId int, email, code string) error {
	if !common.VerifyCodeWithKey(email, code, common.EmailVerificationPurpose) {
		return ErrVerificationCodeInvalid
	}

	user := &model.User{Id: userId}
	if err := user.FillUserById(); err != nil {
		return err
	}

	user.Email = email
	if err := user.Update(false); err != nil {
		return err
	}

	if user.Role == model.RoleRootUser {
		config.RootUserEmail = email
	}

	return nil
}

// TopUp 用户充值
func (s *UserService) TopUp(ctx context.Context, userId int, key string) (int64, error) {
	return model.Redeem(ctx, key, userId)
}

// AdminTopUp 管理员充值
func (s *UserService) AdminTopUp(ctx context.Context, userId int, quota int, remark string) error {
	if err := model.IncreaseUserQuota(userId, int64(quota)); err != nil {
		return err
	}

	if remark == "" {
		remark = fmt.Sprintf("通过 API 充值 %s", common.LogQuota(int64(quota)))
	}
	model.RecordTopupLog(ctx, userId, remark, quota)

	return nil
}
