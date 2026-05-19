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

// GetAllRedemptions
// @Summary 获取所有兑换码
// @Description 获取兑换码列表，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/redemption/ [get]
func GetAllRedemptions(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	redemptionService := service.GetRedemptionService()
	redemptions, err := redemptionService.GetAllRedemptions(p*config.ItemsPerPage, config.ItemsPerPage)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, redemptions)
}

// SearchRedemptions
// @Summary 搜索兑换码
// @Description 根据关键词搜索兑换码，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/redemption/search [get]
func SearchRedemptions(c *gin.Context) {
	keyword := c.Query("keyword")
	redemptionService := service.GetRedemptionService()
	redemptions, err := redemptionService.SearchRedemptions(keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, redemptions)
}

// GetRedemption
// @Summary 获取指定兑换码
// @Description 获取兑换码详情，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "兑换码 ID"
// @Success 200 {object} model.Redemption "成功"
// @Router /api/redemption/{id} [get]
func GetRedemption(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	redemptionService := service.GetRedemptionService()
	redemption, err := redemptionService.GetRedemption(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, redemption)
}

// AddRedemption
// @Summary 创建兑换码
// @Description 创建新兑换码，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.CreateRedemptionRequest true "创建请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/redemption/ [post]
func AddRedemption(c *gin.Context) {
	var req service.CreateRedemptionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	userId := c.GetInt(ctxkey.Id)
	redemptionService := service.GetRedemptionService()
	keys, err := redemptionService.CreateRedemptions(userId, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
			"data":    keys,
		})
		return
	}
	common.ApiSuccess(c, keys)
}

// DeleteRedemption
// @Summary 删除兑换码
// @Description 删除指定兑换码，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "兑换码 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/redemption/{id} [delete]
func DeleteRedemption(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	redemptionService := service.GetRedemptionService()
	err := redemptionService.DeleteRedemption(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// UpdateRedemption
// @Summary 更新兑换码
// @Description 更新兑换码信息，需要管理员权限
// @Tags 兑换码
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "兑换码 ID"
// @Param status_only query bool false "是否只更新状态"
// @Param request body service.UpdateRedemptionRequest true "更新请求"
// @Success 200 {object} model.Redemption "成功"
// @Router /api/redemption/{id} [put]
func UpdateRedemption(c *gin.Context) {
	statusOnly := c.Query("status_only")

	var req service.UpdateRedemptionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	redemptionService := service.GetRedemptionService()
	redemption, err := redemptionService.UpdateRedemption(&req, statusOnly != "")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, redemption)
}
