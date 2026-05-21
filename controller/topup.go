package controller

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
)

// GetTopUpInfo 获取充值信息（包含支付方式列表）
func GetTopUpInfo(c *gin.Context) {
	payMethods := []map[string]interface{}{}

	// 检查 Stripe 是否启用
	if isStripeEnabled() {
		payMethods = append(payMethods, map[string]interface{}{
			"name":  "Stripe",
			"type":  "stripe",
			"color": "#635BFF",
		})
	}

	// 检查支付宝是否启用
	if isAlipayEnabled() {
		payMethods = append(payMethods, map[string]interface{}{
			"name":  "支付宝",
			"type":  "alipay",
			"color": "#1677FF",
		})
	}

	common.ApiSuccess(c, map[string]interface{}{
		"pay_methods": payMethods,
		"min_topup":   config.MinTopUp,
	})
}

// GetUserTopUps 获取用户充值记录
func GetUserTopUps(c *gin.Context) {
	id := c.GetInt(ctxkey.Id)
	p, _ := strconv.Atoi(c.DefaultQuery("p", "0"))
	if p < 0 {
		p = 0
	}

	topUps, total, err := model.GetUserTopUps(id, p, config.ItemsPerPage)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, map[string]interface{}{
		"data":  topUps,
		"total": total,
	})
}

// GetAllTopUps 获取所有充值记录（管理员）
func GetAllTopUps(c *gin.Context) {
	p, _ := strconv.Atoi(c.DefaultQuery("p", "0"))
	if p < 0 {
		p = 0
	}

	topUps, total, err := model.GetAllTopUps(p, config.ItemsPerPage)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, map[string]interface{}{
		"data":  topUps,
		"total": total,
	})
}
