package service

import (
	"errors"
	"fmt"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"github.com/songquanpeng/one-api/common/network"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrTokenNameTooLong       = errors.New("令牌名称过长")
	ErrInvalidSubnet          = errors.New("无效的网段")
	ErrTokenExpired           = errors.New("令牌已过期，无法启用，请先修改令牌过期时间，或者设置为永不过期")
	ErrTokenExhausted         = errors.New("令牌可用额度已用尽，无法启用，请先修改令牌剩余额度，或者设置为无限额度")
	ErrTokenQuotaInsufficient = errors.New("令牌额度不足")
	ErrUserQuotaInsufficient  = errors.New("用户额度不足")
)

// CreateTokenRequest 创建令牌请求参数
type CreateTokenRequest struct {
	Name           string  `json:"name"`
	ExpiredTime    int64   `json:"expired_time"`
	RemainQuota    int64   `json:"remain_quota"`
	UnlimitedQuota bool    `json:"unlimited_quota"`
	Models         *string `json:"models"`
	Subnet         *string `json:"subnet"`
}

// UpdateTokenRequest 更新令牌请求参数
type UpdateTokenRequest struct {
	Id             int     `json:"id"`
	Name           string  `json:"name"`
	Status         int     `json:"status"`
	ExpiredTime    int64   `json:"expired_time"`
	RemainQuota    int64   `json:"remain_quota"`
	UnlimitedQuota bool    `json:"unlimited_quota"`
	Models         *string `json:"models"`
	Subnet         *string `json:"subnet"`
}

// TokenService 令牌服务
type TokenService struct{}

// NewTokenService 创建令牌服务实例
func NewTokenService() *TokenService {
	return &TokenService{}
}

// validateToken 验证令牌参数
func (s *TokenService) validateToken(name string, subnet *string) error {
	if len(name) > 30 {
		return ErrTokenNameTooLong
	}
	if subnet != nil && *subnet != "" {
		if err := network.IsValidSubnets(*subnet); err != nil {
			return fmt.Errorf("%w: %s", ErrInvalidSubnet, err.Error())
		}
	}
	return nil
}

// CreateToken 创建令牌
func (s *TokenService) CreateToken(userId int, req *CreateTokenRequest) (*model.Token, error) {
	// 1. 参数验证
	if err := s.validateToken(req.Name, req.Subnet); err != nil {
		return nil, err
	}

	// 2. 创建令牌
	token := &model.Token{
		UserId:         userId,
		Name:           req.Name,
		Key:            random.GenerateKey(),
		CreatedTime:    helper.GetTimestamp(),
		AccessedTime:   helper.GetTimestamp(),
		ExpiredTime:    req.ExpiredTime,
		RemainQuota:    req.RemainQuota,
		UnlimitedQuota: req.UnlimitedQuota,
		Models:         req.Models,
		Subnet:         req.Subnet,
	}

	if err := token.Insert(); err != nil {
		return nil, err
	}

	return token, nil
}

// UpdateToken 更新令牌
func (s *TokenService) UpdateToken(userId int, req *UpdateTokenRequest, statusOnly bool) (*model.Token, error) {
	// 1. 参数验证
	if err := s.validateToken(req.Name, req.Subnet); err != nil {
		return nil, err
	}

	// 2. 获取现有令牌
	cleanToken, err := model.GetTokenByIds(req.Id, userId)
	if err != nil {
		return nil, err
	}

	// 3. 状态机逻辑验证
	if req.Status == model.TokenStatusEnabled {
		if cleanToken.Status == model.TokenStatusExpired && cleanToken.ExpiredTime <= helper.GetTimestamp() && cleanToken.ExpiredTime != -1 {
			return nil, ErrTokenExpired
		}
		if cleanToken.Status == model.TokenStatusExhausted && cleanToken.RemainQuota <= 0 && !cleanToken.UnlimitedQuota {
			return nil, ErrTokenExhausted
		}
	}

	// 4. 更新字段
	if statusOnly {
		cleanToken.Status = req.Status
	} else {
		cleanToken.Name = req.Name
		cleanToken.ExpiredTime = req.ExpiredTime
		cleanToken.RemainQuota = req.RemainQuota
		cleanToken.UnlimitedQuota = req.UnlimitedQuota
		cleanToken.Models = req.Models
		cleanToken.Subnet = req.Subnet
	}

	// 5. 保存更新
	if err := cleanToken.Update(); err != nil {
		return nil, err
	}

	return cleanToken, nil
}

// DeleteToken 删除令牌
func (s *TokenService) DeleteToken(id, userId int) error {
	return model.DeleteTokenById(id, userId)
}

// GetToken 获取令牌
func (s *TokenService) GetToken(id, userId int) (*model.Token, error) {
	return model.GetTokenByIds(id, userId)
}

// GetAllTokens 获取用户所有令牌
func (s *TokenService) GetAllTokens(userId, startIdx, num int, order string) ([]*model.Token, error) {
	return model.GetAllUserTokens(userId, startIdx, num, order)
}

// SearchTokens 搜索令牌
func (s *TokenService) SearchTokens(userId int, keyword string) ([]*model.Token, error) {
	return model.SearchUserTokens(userId, keyword)
}

// PreConsumeTokenQuota 预扣费令牌额度
func (s *TokenService) PreConsumeTokenQuota(tokenId int, quota int64) error {
	if quota < 0 {
		return errors.New("quota 不能为负数")
	}

	// 1. 获取令牌
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		return err
	}

	// 2. 检查令牌额度
	if !token.UnlimitedQuota && token.RemainQuota < quota {
		return ErrTokenQuotaInsufficient
	}

	// 3. 检查用户额度
	userQuota, err := model.GetUserQuota(token.UserId)
	if err != nil {
		return err
	}
	if userQuota < quota {
		return ErrUserQuotaInsufficient
	}

	// 4. 发送低额度提醒
	s.checkAndSendQuotaNotification(token.UserId, userQuota, quota)

	// 5. 扣除额度
	if !token.UnlimitedQuota {
		if err := model.DecreaseTokenQuota(tokenId, quota); err != nil {
			return err
		}
	}
	return model.DecreaseUserQuota(token.UserId, quota)
}

// checkAndSendQuotaNotification 检查并发送额度通知
func (s *TokenService) checkAndSendQuotaNotification(userId int, userQuota, quota int64) {
	quotaTooLow := userQuota >= config.QuotaRemindThreshold && userQuota-quota < config.QuotaRemindThreshold
	noMoreQuota := userQuota-quota <= 0

	if quotaTooLow || noMoreQuota {
		go func() {
			email, err := model.GetUserEmail(userId)
			if err != nil {
				logger.SysError("failed to fetch user email: " + err.Error())
				return
			}
			if email == "" {
				return
			}

			prompt := "额度提醒"
			var contentText string
			if noMoreQuota {
				contentText = "您的额度已用尽"
			} else {
				contentText = "您的额度即将用尽"
			}

			topUpLink := fmt.Sprintf("%s/topup", config.ServerAddress)
			content := message.EmailTemplate(
				prompt,
				fmt.Sprintf(`
					<p>您好！</p>
					<p>%s，当前剩余额度为 <strong>%d</strong>。</p>
					<p>为了不影响您的使用，请及时充值。</p>
					<p style="text-align: center; margin: 30px 0;">
						<a href="%s" style="background-color: #007bff; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">立即充值</a>
					</p>
					<p style="color: #666;">如果按钮无法点击，请复制以下链接到浏览器中打开：</p>
					<p style="background-color: #f8f8f8; padding: 10px; border-radius: 4px; word-break: break-all;">%s</p>
				`, contentText, userQuota, topUpLink, topUpLink),
			)
			if err := message.SendEmail(prompt, email, content); err != nil {
				logger.SysError("failed to send email: " + err.Error())
			}
		}()
	}
}

// PostConsumeTokenQuota 后处理令牌额度
func (s *TokenService) PostConsumeTokenQuota(tokenId int, quota int64) error {
	token, err := model.GetTokenById(tokenId)
	if err != nil {
		return err
	}

	if quota > 0 {
		if err := model.DecreaseUserQuota(token.UserId, quota); err != nil {
			return err
		}
	} else {
		if err := model.IncreaseUserQuota(token.UserId, -quota); err != nil {
			return err
		}
	}

	if !token.UnlimitedQuota {
		if quota > 0 {
			return model.DecreaseTokenQuota(tokenId, quota)
		}
		return model.IncreaseTokenQuota(tokenId, -quota)
	}

	return nil
}
