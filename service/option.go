package service

import (
	"errors"
	"strings"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrInvalidTheme                 = errors.New("无效的主题")
	ErrGitHubOAuthClientIdRequired  = errors.New("无法启用 GitHub OAuth，请先填入 GitHub Client Id 以及 GitHub Client Secret")
	ErrEmailDomainWhitelistRequired = errors.New("无法启用邮箱域名限制，请先填入限制的邮箱域名")
	ErrWeChatConfigRequired         = errors.New("无法启用微信登录，请先填入微信登录相关配置信息")
	ErrTurnstileConfigRequired      = errors.New("无法启用 Turnstile 校验，请先填入 Turnstile 校验相关配置信息")
)

// OptionService 配置服务
type OptionService struct{}

// NewOptionService 创建配置服务实例
func NewOptionService() *OptionService {
	return &OptionService{}
}

// GetPublicOptions 获取公开配置项（过滤敏感信息）
func (s *OptionService) GetPublicOptions() []*model.Option {
	var options []*model.Option
	config.OptionMapRWMutex.Lock()
	defer config.OptionMapRWMutex.Unlock()

	for k, v := range config.OptionMap {
		// 过滤敏感字段
		if strings.HasSuffix(k, "Token") || strings.HasSuffix(k, "Secret") {
			continue
		}
		options = append(options, &model.Option{
			Key:   k,
			Value: helper.Interface2String(v),
		})
	}
	return options
}

// UpdateOption 更新配置项
func (s *OptionService) UpdateOption(key, value string) error {
	// 验证配置项
	if err := s.validateOption(key, value); err != nil {
		return err
	}

	// 更新配置
	return model.UpdateOption(key, value)
}

// validateOption 验证配置项
func (s *OptionService) validateOption(key, value string) error {
	switch key {
	case "Theme":
		if !config.ValidThemes[value] {
			return ErrInvalidTheme
		}
	case "GitHubOAuthEnabled":
		if value == "true" && config.GitHubClientId == "" {
			return ErrGitHubOAuthClientIdRequired
		}
	case "EmailDomainRestrictionEnabled":
		if value == "true" && len(config.EmailDomainWhitelist) == 0 {
			return ErrEmailDomainWhitelistRequired
		}
	case "WeChatAuthEnabled":
		if value == "true" && config.WeChatServerAddress == "" {
			return ErrWeChatConfigRequired
		}
	case "TurnstileCheckEnabled":
		if value == "true" && config.TurnstileSiteKey == "" {
			return ErrTurnstileConfigRequired
		}
	}
	return nil
}
