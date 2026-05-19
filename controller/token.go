package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/service"
	"net/http"
	"strconv"
)

// GetAllTokens
// @Summary 获取所有令牌
// @Description 获取当前用户的所有令牌
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Param order query string false "排序方式" Enums(remain_quota, used_quota)
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/token/ [get]
func GetAllTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}

	order := c.Query("order")
	tokenService := service.GetTokenService()
	tokens, err := tokenService.GetAllTokens(userId, p*config.ItemsPerPage, config.ItemsPerPage, order)

	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tokens)
}

// SearchTokens
// @Summary 搜索令牌
// @Description 根据关键词搜索令牌
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/token/search [get]
func SearchTokens(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	keyword := c.Query("keyword")
	tokenService := service.GetTokenService()
	tokens, err := tokenService.SearchTokens(userId, keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, tokens)
}

// GetToken
// @Summary 获取指定令牌
// @Description 获取令牌详情
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "令牌 ID"
// @Success 200 {object} model.Token "成功"
// @Router /api/token/{id} [get]
func GetToken(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	tokenService := service.GetTokenService()
	token, err := tokenService.GetToken(id, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, token)
}

// GetTokenStatus
// @Summary 获取令牌状态
// @Description 获取令牌的额度信息
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /dashboard/billing/subscription [get]
func GetTokenStatus(c *gin.Context) {
	tokenId := c.GetInt(ctxkey.TokenId)
	userId := c.GetInt(ctxkey.Id)
	tokenService := service.GetTokenService()
	token, err := tokenService.GetToken(tokenId, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	expiredAt := token.ExpiredTime
	if expiredAt == -1 {
		expiredAt = 0
	}
	c.JSON(http.StatusOK, gin.H{
		"object":          "credit_summary",
		"total_granted":   token.RemainQuota,
		"total_used":      0, // not supported currently
		"total_available": token.RemainQuota,
		"expires_at":      expiredAt * 1000,
	})
}

// AddToken
// @Summary 创建令牌
// @Description 创建新令牌
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.CreateTokenRequest true "创建请求"
// @Success 200 {object} model.Token "成功"
// @Router /api/token/ [post]
func AddToken(c *gin.Context) {
	var req service.CreateTokenRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	userId := c.GetInt(ctxkey.Id)
	tokenService := service.GetTokenService()
	token, err := tokenService.CreateToken(userId, &req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, token)
}

// DeleteToken
// @Summary 删除令牌
// @Description 删除指定令牌
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "令牌 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/token/{id} [delete]
func DeleteToken(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userId := c.GetInt(ctxkey.Id)
	tokenService := service.GetTokenService()
	err := tokenService.DeleteToken(id, userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// UpdateToken
// @Summary 更新令牌
// @Description 更新令牌信息
// @Tags 令牌管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.UpdateTokenRequest true "更新请求"
// @Success 200 {object} model.Token "成功"
// @Router /api/token/ [put]
func UpdateToken(c *gin.Context) {
	userId := c.GetInt(ctxkey.Id)
	statusOnly := c.Query("status_only")

	var req service.UpdateTokenRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	tokenService := service.GetTokenService()
	token, err := tokenService.UpdateToken(userId, &req, statusOnly != "")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, token)
}
