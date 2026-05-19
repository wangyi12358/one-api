package controller

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/i18n"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/service"
)

// GetOptions
// @Summary 获取系统设置
// @Description 获取所有系统设置，需要 Root 权限
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security UserAuth
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/option/ [get]
func GetOptions(c *gin.Context) {
	optionService := service.GetOptionService()
	options := optionService.GetPublicOptions()
	common.ApiSuccess(c, options)
}

// UpdateOption
// @Summary 更新系统设置
// @Description 更新系统设置，需要 Root 权限
// @Tags 系统设置
// @Accept json
// @Produce json
// @Security UserAuth
// @Param request body model.Option true "设置请求"
// @Success 200 {object} map[string]interface{} "成功"
// @Router /api/option/ [put]
func UpdateOption(c *gin.Context) {
	var option model.Option
	err := json.NewDecoder(c.Request.Body).Decode(&option)
	if err != nil {
		common.ApiErrorMsg(c, i18n.Translate(c, "invalid_parameter"))
		return
	}

	optionService := service.GetOptionService()
	err = optionService.UpdateOption(option.Key, option.Value)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, nil)
}
