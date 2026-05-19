package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/service"
)

// GetAllLogs
// @Summary 获取所有日志
// @Description 获取日志列表，需要管理员权限
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Param type query int false "日志类型"
// @Param start_timestamp query int false "开始时间戳"
// @Param end_timestamp query int false "结束时间戳"
// @Param username query string false "用户名"
// @Param token_name query string false "令牌名称"
// @Param model_name query string false "模型名称"
// @Param channel query int false "渠道 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/ [get]
func GetAllLogs(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	username := c.Query("username")
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))

	query := &service.LogQuery{
		Page:           p,
		PageSize:       config.ItemsPerPage,
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
		Channel:        channel,
	}

	logService := service.GetLogService()
	logs, err := logService.GetAllLogs(query)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, logs)
}

// GetUserLogs
// @Summary 获取用户日志
// @Description 获取当前用户的日志
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param p query int false "页码" default(0)
// @Param type query int false "日志类型"
// @Param start_timestamp query int false "开始时间戳"
// @Param end_timestamp query int false "结束时间戳"
// @Param token_name query string false "令牌名称"
// @Param model_name query string false "模型名称"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/self [get]
func GetUserLogs(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	if p < 0 {
		p = 0
	}
	userId := c.GetInt(ctxkey.Id)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")

	query := &service.LogQuery{
		Page:           p,
		PageSize:       config.ItemsPerPage,
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		TokenName:      tokenName,
	}

	logService := service.GetLogService()
	logs, err := logService.GetUserLogs(userId, query)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, logs)
}

// SearchAllLogs
// @Summary 搜索所有日志
// @Description 根据关键词搜索日志，需要管理员权限
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/search [get]
func SearchAllLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	logService := service.GetLogService()
	logs, err := logService.SearchAllLogs(keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, logs)
}

// SearchUserLogs
// @Summary 搜索用户日志
// @Description 根据关键词搜索当前用户的日志
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param keyword query string true "搜索关键词"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/self/search [get]
func SearchUserLogs(c *gin.Context) {
	keyword := c.Query("keyword")
	userId := c.GetInt(ctxkey.Id)
	logService := service.GetLogService()
	logs, err := logService.SearchUserLogs(userId, keyword)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, logs)
}

// GetLogsStat
// @Summary 获取日志统计
// @Description 获取日志统计数据，需要管理员权限
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param type query int false "日志类型"
// @Param start_timestamp query int false "开始时间戳"
// @Param end_timestamp query int false "结束时间戳"
// @Param token_name query string false "令牌名称"
// @Param username query string false "用户名"
// @Param model_name query string false "模型名称"
// @Param channel query int false "渠道 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/stat [get]
func GetLogsStat(c *gin.Context) {
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	username := c.Query("username")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))

	query := &service.LogQuery{
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
		Channel:        channel,
	}

	logService := service.GetLogService()
	stat, err := logService.GetLogsStat(query)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"quota": stat.Quota,
	})
}

// GetLogsSelfStat
// @Summary 获取用户日志统计
// @Description 获取当前用户的日志统计
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param type query int false "日志类型"
// @Param start_timestamp query int false "开始时间戳"
// @Param end_timestamp query int false "结束时间戳"
// @Param token_name query string false "令牌名称"
// @Param model_name query string false "模型名称"
// @Param channel query int false "渠道 ID"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/self/stat [get]
func GetLogsSelfStat(c *gin.Context) {
	username := c.GetString(ctxkey.Username)
	logType, _ := strconv.Atoi(c.Query("type"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	tokenName := c.Query("token_name")
	modelName := c.Query("model_name")
	channel, _ := strconv.Atoi(c.Query("channel"))

	query := &service.LogQuery{
		LogType:        logType,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
		ModelName:      modelName,
		Username:       username,
		TokenName:      tokenName,
		Channel:        channel,
	}

	logService := service.GetLogService()
	stat, err := logService.GetLogsStat(query)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"quota": stat.Quota,
	})
}

// DeleteHistoryLogs
// @Summary 删除历史日志
// @Description 删除指定时间之前的日志，需要管理员权限
// @Tags 日志
// @Accept json
// @Produce json
// @Security UserAuth
// @Param target_timestamp query int true "目标时间戳"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/log/ [delete]
func DeleteHistoryLogs(c *gin.Context) {
	targetTimestamp, _ := strconv.ParseInt(c.Query("target_timestamp"), 10, 64)
	logService := service.GetLogService()
	count, err := logService.DeleteHistoryLogs(targetTimestamp)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, count)
}
