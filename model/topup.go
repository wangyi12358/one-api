package model

import (
	"context"
	"errors"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
)

// TopUp 订单状态
const (
	TopUpStatusPending = "pending"
	TopUpStatusSuccess = "success"
	TopUpStatusFailed  = "failed"
	TopUpStatusExpired = "expired"
)

// 支付方式
const (
	PaymentMethodStripe = "stripe"
	PaymentMethodAlipay = "alipay"
)

// 支付提供商
const (
	PaymentProviderStripe = "stripe"
	PaymentProviderAlipay = "alipay"
)

type TopUp struct {
	Id              int     `json:"id"`
	UserId          int     `json:"user_id" gorm:"index"`
	Amount          int64   `json:"amount"`
	Money           float64 `json:"money"`
	TradeNo         string  `json:"trade_no" gorm:"unique;type:varchar(255);index"`
	PaymentMethod   string  `json:"payment_method" gorm:"type:varchar(50)"`
	PaymentProvider string  `json:"payment_provider" gorm:"type:varchar(50);default:''"`
	CreateTime      int64   `json:"create_time"`
	CompleteTime    int64   `json:"complete_time"`
	Status          string  `json:"status" gorm:"type:varchar(20);default:'pending'"`
}

var (
	ErrTopUpNotFound      = errors.New("topup not found")
	ErrTopUpStatusInvalid = errors.New("topup status invalid")
)

func (topUp *TopUp) Insert() error {
	return DB.Create(topUp).Error
}

func (topUp *TopUp) Update() error {
	return DB.Save(topUp).Error
}

func GetTopUpByTradeNo(tradeNo string) *TopUp {
	var topUp TopUp
	err := DB.Where("trade_no = ?", tradeNo).First(&topUp).Error
	if err != nil {
		return nil
	}
	return &topUp
}

func GetUserTopUps(userId int, page int, pageSize int) (topUps []*TopUp, total int64, err error) {
	offset := (page - 1) * pageSize
	err = DB.Model(&TopUp{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Where("user_id = ?", userId).Order("id desc").Limit(pageSize).Offset(offset).Find(&topUps).Error
	return topUps, total, err
}

func GetAllTopUps(page int, pageSize int) (topUps []*TopUp, total int64, err error) {
	offset := (page - 1) * pageSize
	err = DB.Model(&TopUp{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Order("id desc").Limit(pageSize).Offset(offset).Find(&topUps).Error
	return topUps, total, err
}

// Recharge 完成充值
func Recharge(tradeNo string, userId int, amount int64, money float64, paymentMethod string, paymentProvider string) error {
	topUp := GetTopUpByTradeNo(tradeNo)
	if topUp == nil {
		return ErrTopUpNotFound
	}

	if topUp.Status != TopUpStatusPending {
		return ErrTopUpStatusInvalid
	}

	// 计算充值额度
	quota := int64(float64(amount) * config.QuotaPerUnit)

	// 更新订单状态
	topUp.Status = TopUpStatusSuccess
	topUp.CompleteTime = helper.GetTimestamp()
	err := topUp.Update()
	if err != nil {
		return err
	}

	// 增加用户额度
	err = IncreaseUserQuota(userId, quota)
	if err != nil {
		logger.SysError("failed to increase user quota: " + err.Error())
		return err
	}

	// 记录日志
	ctx := context.Background()
	RecordLog(ctx, userId, LogTypeTopup, "在线充值成功，充值额度: "+common.LogQuota(quota))

	return nil
}
