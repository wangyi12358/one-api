package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/service"
	"strconv"
)

// GetAllChannels
// @Summary 获取所有渠道
// @Description 获取渠道列表，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/channel/ [get]
func GetAllChannels(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	channelService := service.GetChannelService()
	channels, err := channelService.GetAllChannels(p*config.ItemsPerPage, config.ItemsPerPage, "limited")
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, channels)
}

// SearchChannels
// @Summary 搜索渠道
// @Description 根据关键词搜索渠道，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/channel/search [get]
func SearchChannels(c *gin.Context) {
	keyword := c.Query("keyword")
	channelService := service.GetChannelService()
	channels, err := channelService.SearchChannels(keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, channels)
}

// GetChannel
// @Summary 获取指定渠道
// @Description 获取渠道详情，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "渠道 ID"
// @Success 200 {object} model.Channel "成功"
// @Router /api/channel/{id} [get]
func GetChannel(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channelService := service.GetChannelService()
	channel, err := channelService.GetChannel(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, channel)
}

// AddChannel
// @Summary 创建渠道
// @Description 创建新渠道，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body service.CreateChannelRequest true "创建请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/channel/ [post]
func AddChannel(c *gin.Context) {
	var req service.CreateChannelRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channelService := service.GetChannelService()
	err = channelService.CreateChannel(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// DeleteChannel
// @Summary 删除渠道
// @Description 删除指定渠道，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param id path int true "渠道 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/channel/{id} [delete]
func DeleteChannel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	channelService := service.GetChannelService()
	err := channelService.DeleteChannel(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// DeleteDisabledChannel
// @Summary 删除禁用渠道
// @Description 删除所有禁用的渠道，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/channel/disabled [delete]
func DeleteDisabledChannel(c *gin.Context) {
	channelService := service.GetChannelService()
	rows, err := channelService.DeleteDisabledChannels()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, rows)
}

// UpdateChannel
// @Summary 更新渠道
// @Description 更新渠道信息，需要管理员权限
// @Tags 渠道管理
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body model.Channel true "更新请求"
// @Success 200 {object} model.Channel "成功"
// @Router /api/channel/ [put]
func UpdateChannel(c *gin.Context) {
	channel := model.Channel{}
	err := c.ShouldBindJSON(&channel)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	channelService := service.GetChannelService()
	err = channelService.UpdateChannel(&channel)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, channel)
}
