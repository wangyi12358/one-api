package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/smartwalle/alipay/v3"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

var alipayClient *alipay.Client

// isAlipayEnabled 检查支付宝是否启用
func isAlipayEnabled() bool {
	return config.AlipayAppId != "" &&
		config.AlipayPrivateKey != "" &&
		config.AlipayPublicKey != ""
}

// initAlipayClient 初始化支付宝客户端
func initAlipayClient() error {
	if config.AlipayAppId == "" || config.AlipayPrivateKey == "" || config.AlipayPublicKey == "" {
		return fmt.Errorf("支付宝配置不完整")
	}

	client, err := alipay.New(config.AlipayAppId, config.AlipayPrivateKey, true)
	if err != nil {
		return fmt.Errorf("初始化支付宝客户端失败: %v", err)
	}

	err = client.LoadAliPayPublicKey(config.AlipayPublicKey)
	if err != nil {
		return fmt.Errorf("加载支付宝公钥失败: %v", err)
	}

	alipayClient = client
	return nil
}

// getAlipayClient 获取支付宝客户端
func getAlipayClient() (*alipay.Client, error) {
	if alipayClient == nil {
		if err := initAlipayClient(); err != nil {
			return nil, err
		}
	}
	return alipayClient, nil
}

// AlipayPayRequest 支付宝支付请求
type AlipayPayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

// RequestAlipayAmount 计算支付宝支付金额
func RequestAlipayAmount(c *gin.Context) {
	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.Amount < int64(config.MinTopUp) {
		common.ApiErrorMsg(c, fmt.Sprintf("充值数量不能小于 %d", config.MinTopUp))
		return
	}

	payMoney := float64(req.Amount) * config.AlipayUnitPrice
	common.ApiSuccess(c, fmt.Sprintf("%.2f", payMoney))
}

// RequestAlipayPay 发起支付宝支付
func RequestAlipayPay(c *gin.Context) {
	if !isAlipayEnabled() {
		common.ApiErrorMsg(c, "支付宝支付未启用")
		return
	}

	var req AlipayPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.PaymentMethod != "alipay" {
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
	_, err := model.GetUserById(id, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 生成订单号
	tradeNo := "alipay_" + random.GetUUID()

	// 计算支付金额
	payMoney := float64(req.Amount) * config.AlipayUnitPrice

	// 创建支付宝支付链接
	payLink, err := genAlipayLink(tradeNo, req.Amount, payMoney)
	if err != nil {
		logger.SysError(fmt.Sprintf("支付宝创建支付订单失败: %v", err))
		common.ApiErrorMsg(c, "拉起支付失败")
		return
	}

	// 创建充值订单
	topUp := &model.TopUp{
		UserId:          id,
		Amount:          req.Amount,
		Money:           payMoney,
		TradeNo:         tradeNo,
		PaymentMethod:   model.PaymentMethodAlipay,
		PaymentProvider: model.PaymentProviderAlipay,
		CreateTime:      helper.GetTimestamp(),
		Status:          model.TopUpStatusPending,
	}
	if err := topUp.Insert(); err != nil {
		logger.SysError(fmt.Sprintf("创建充值订单失败: %v", err))
		common.ApiErrorMsg(c, "创建订单失败")
		return
	}

	logger.SysLog(fmt.Sprintf("支付宝充值订单创建成功 user_id=%d trade_no=%s amount=%d", id, tradeNo, req.Amount))

	common.ApiSuccess(c, map[string]interface{}{
		"pay_link": payLink,
	})
}

// AlipayNotify 处理支付宝异步通知
func AlipayNotify(c *gin.Context) {
	if !isAlipayEnabled() {
		c.String(http.StatusForbidden, "fail")
		return
	}

	client, err := getAlipayClient()
	if err != nil {
		logger.SysError(fmt.Sprintf("支付宝客户端初始化失败: %v", err))
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	notification, err := client.GetTradeNotification(c.Request)
	if err != nil {
		logger.SysError(fmt.Sprintf("支付宝回调解析失败: %v", err))
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if notification == nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	tradeNo := notification.OutTradeNo
	tradeStatus := notification.TradeStatus

	logger.SysLog(fmt.Sprintf("支付宝回调收到通知 trade_no=%s trade_status=%s", tradeNo, tradeStatus))

	// 只处理成功状态
	if tradeStatus != "TRADE_SUCCESS" && tradeStatus != "TRADE_FINISHED" {
		c.String(http.StatusOK, "success")
		return
	}

	topUp := model.GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		logger.SysError(fmt.Sprintf("支付宝充值订单不存在: %s", tradeNo))
		c.String(http.StatusOK, "success")
		return
	}

	if topUp.Status != model.TopUpStatusPending {
		c.String(http.StatusOK, "success")
		return
	}

	// 完成充值
	err = model.Recharge(tradeNo, topUp.UserId, topUp.Amount, topUp.Money, model.PaymentMethodAlipay, model.PaymentProviderAlipay)
	if err != nil {
		logger.SysError(fmt.Sprintf("支付宝充值失败: %v", err))
		c.String(http.StatusInternalServerError, "fail")
		return
	}

	logger.SysLog(fmt.Sprintf("支付宝充值成功 trade_no=%s", tradeNo))
	c.String(http.StatusOK, "success")
}

// AlipayReturn 处理支付宝同步返回
func AlipayReturn(c *gin.Context) {
	tradeNo := c.Query("out_trade_no")
	if tradeNo == "" {
		c.Redirect(http.StatusFound, "/topup")
		return
	}

	c.Redirect(http.StatusFound, "/topup")
}

// genAlipayLink 生成支付宝支付 URL
func genAlipayLink(tradeNo string, amount int64, payMoney float64) (string, error) {
	client, err := getAlipayClient()
	if err != nil {
		return "", err
	}

	returnURL := config.AlipayReturnURL
	if returnURL == "" {
		returnURL = config.ServerAddress + "/topup"
	}

	p := alipay.TradePagePay{}
	p.NotifyURL = config.AlipayNotifyURL
	p.ReturnURL = returnURL
	p.Subject = fmt.Sprintf("充值 %d 额度", amount)
	p.OutTradeNo = tradeNo
	p.TotalAmount = strconv.FormatFloat(payMoney, 'f', 2, 64)
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"

	url, err := client.TradePagePay(p)
	if err != nil {
		return "", err
	}

	return url.String(), nil
}
