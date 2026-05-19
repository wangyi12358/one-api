package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/service"
)

func TestChannel(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}

	modelName := c.Query("model")
	channelTestService := service.GetChannelTestService()
	result, err := channelTestService.TestChannel(ctx, id, modelName)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   result.Success,
		"message":   result.Message,
		"time":      result.Time,
		"modelName": result.ModelName,
	})
}

func TestChannels(c *gin.Context) {
	ctx := c.Request.Context()
	scope := c.Query("scope")
	if scope == "" {
		scope = "all"
	}

	channelTestService := service.GetChannelTestService()
	err := channelTestService.TestAllChannels(ctx, true, scope)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}

func AutomaticallyTestChannels(frequency int) {
	channelTestService := service.GetChannelTestService()
	channelTestService.AutomaticallyTestChannels(frequency)
}
