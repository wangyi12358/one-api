package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/webhook"
)

// isStripeEnabled 检查 Stripe 是否启用
func isStripeEnabled() bool {
	return config.StripeApiSecret != "" &&
		config.StripeWebhookSecret != "" &&
		config.StripePriceId != ""
}

// StripePayRequest Stripe 支付请求
type StripePayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// RequestStripeAmount 计算 Stripe 支付金额
func RequestStripeAmount(c *gin.Context) {
	var req StripePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.Amount < int64(config.MinTopUp) {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", config.MinTopUp))
		return
	}

	payMoney := float64(req.Amount) * config.StripeUnitPrice
	common.ApiSuccess(c, fmt.Sprintf("%.2f", payMoney))
}

// RequestStripePay 发起 Stripe 支付
func RequestStripePay(c *gin.Context) {
	if !isStripeEnabled() {
		common.ApiErrorMsg(c, "Stripe 支付未启用")
		return
	}

	var req StripePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.PaymentMethod != "stripe" {
		common.ApiErrorMsg(c, "不支持的支付方式")
		return
	}

	if req.Amount < int64(config.MinTopUp) {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", config.MinTopUp))
		return
	}

	if req.Amount > 10000 {
		common.ApiErrorMsg(c, "充值数量不能大于 10000")
		return
	}

	id := c.GetInt(ctxkey.Id)
	user, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 生成订单号
	tradeNo := "ref_" + random.GetUUID()

	// 计算支付金额
	payMoney := float64(req.Amount) * config.StripeUnitPrice

	// 创建 Stripe Checkout Session
	payLink, err := genStripeLink(tradeNo, user.Email, req.Amount)
	if err != nil {
		logger.SysError(fmt.Sprintf("Stripe 创建 Checkout Session 失败: %v", err))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	// 创建充值订单
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          req.Amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodStripe,
		PaymentProvider: model.PaymentProviderStripe,
		CreateTime:      helper.GetTimestamp(),
		Status:          model.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.SysError(fmt.Sprintf("创建充值订单失败: %v", err))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	logger.SysLog(fmt.Sprintf("Stripe 充值订单创建成功 user_id=%d trade_no=%s amount=%d", id, tradeNo, req.Amount))

	common.ApiSuccess(c, map[string]interface{}{
		"pay_link": payLink,
	})
}

// StripeWebhook 处理 Stripe Webhook 回调
func StripeWebhook(c *gin.Context) {
	if config.StripeWebhookSecret == "" {
		c.Status(http.StatusForbidden)
		return
	}

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	signature := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEventWithOptions(payload, signature, config.StripeWebhookSecret, webhook.ConstructEventOptions{
		IgnoreAPIVersionMismatch: true,
	})
	if err != nil {
		logger.SysError(fmt.Sprintf("Stripe webhook 验签失败: %v", err))
		c.Status(http.StatusBadRequest)
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		handleStripeSessionCompleted(c, event)
	case stripe.EventTypeCheckoutSessionExpired:
		handleStripeSessionExpired(c, event)
	default:
		logger.SysLog(fmt.Sprintf("Stripe webhook 忽略事件: %s", event.Type))
	}

	c.Status(http.StatusOK)
}

func handleStripeSessionCompleted(c *gin.Context, event stripe.Event) {
	referenceId := event.GetObjectValue("client_reference_id")
	status := event.GetObjectValue("status")
	paymentStatus := event.GetObjectValue("payment_status")

	if status != "complete" || paymentStatus != "paid" {
		return
	}

	topUp := model.GetTopUpByTradeNo(referenceId)
	if topUp == nil {
		logger.SysError(fmt.Sprintf("Stripe 充值订单不存在: %s", referenceId))
		return
	}

	if topUp.Status != model.TopUpStatusPending {
		return
	}

	// 完成充值
	err := model.Recharge(referenceId, topUp.UserId, topUp.Amount, topUp.Money, model.PaymentMethodStripe, model.PaymentProviderStripe)
	if err != nil {
		logger.SysError(fmt.Sprintf("Stripe 充值失败: %v", err))
		return
	}

	logger.SysLog(fmt.Sprintf("Stripe 充值成功 trade_no=%s", referenceId))
}

func handleStripeSessionExpired(c *gin.Context, event stripe.Event) {
	referenceId := event.GetObjectValue("client_reference_id")
	topUp := model.GetTopUpByTradeNo(referenceId)
	if topUp == nil || topUp.Status != model.TopUpStatusPending {
		return
	}

	topUp.Status = model.TopUpStatusExpired
	topUp.Update()
}

// genStripeLink 生成 Stripe Checkout Session URL
func genStripeLink(tradeNo string, email string, amount int64) (string, error) {
	stripe.Key = config.StripeApiSecret

	successURL := config.StripeSuccessURL
	if successURL == "" {
		successURL = config.ServerAddress + "/topup"
	}
	cancelURL := config.StripeCancelURL
	if cancelURL == "" {
		cancelURL = config.ServerAddress + "/topup"
	}

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(tradeNo),
		SuccessURL:        stripe.String(successURL),
		CancelURL:         stripe.String(cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(config.StripePriceId),
				Quantity: stripe.Int64(amount),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
	}

	if email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	result, err := session.New(params)
	if err != nil {
		return "", err
	}

	return result.URL, nil
}

// getStripeMinTopup 获取 Stripe 最小充值数量
func getStripeMinTopup() int64 {
	return int64(config.MinTopUp)
}

// getStripePayMoney 计算 Stripe 支付金额
func getStripePayMoney(amount float64) float64 {
	return amount * config.StripeUnitPrice
}
