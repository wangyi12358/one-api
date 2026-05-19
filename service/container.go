package service

import (
	"sync"
)

// Container 服务容器
type Container struct {
	mu sync.RWMutex

	// 服务实例
	userService           IUserService
	tokenService          ITokenService
	channelService        IChannelService
	redemptionService     IRedemptionService
	authService           IAuthService
	logService            ILogService
	optionService         IOptionService
	channelBillingService IChannelBillingService
	channelTestService    IChannelTestService
}

// 全局服务容器
var globalContainer = &Container{}

// GetContainer 获取全局服务容器
func GetContainer() *Container {
	return globalContainer
}

// SetUserService 设置用户服务
func (c *Container) SetUserService(s IUserService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userService = s
}

// GetUserService 获取用户服务
func (c *Container) GetUserService() IUserService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.userService == nil {
		c.userService = NewUserService()
	}
	return c.userService
}

// SetTokenService 设置令牌服务
func (c *Container) SetTokenService(s ITokenService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tokenService = s
}

// GetTokenService 获取令牌服务
func (c *Container) GetTokenService() ITokenService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.tokenService == nil {
		c.tokenService = NewTokenService()
	}
	return c.tokenService
}

// SetChannelService 设置渠道服务
func (c *Container) SetChannelService(s IChannelService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channelService = s
}

// GetChannelService 获取渠道服务
func (c *Container) GetChannelService() IChannelService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.channelService == nil {
		c.channelService = NewChannelService()
	}
	return c.channelService
}

// SetRedemptionService 设置兑换码服务
func (c *Container) SetRedemptionService(s IRedemptionService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.redemptionService = s
}

// GetRedemptionService 获取兑换码服务
func (c *Container) GetRedemptionService() IRedemptionService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.redemptionService == nil {
		c.redemptionService = NewRedemptionService()
	}
	return c.redemptionService
}

// SetAuthService 设置认证服务
func (c *Container) SetAuthService(s IAuthService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.authService = s
}

// GetAuthService 获取认证服务
func (c *Container) GetAuthService() IAuthService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.authService == nil {
		c.authService = NewAuthService()
	}
	return c.authService
}

// SetLogService 设置日志服务
func (c *Container) SetLogService(s ILogService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logService = s
}

// GetLogService 获取日志服务
func (c *Container) GetLogService() ILogService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.logService == nil {
		c.logService = NewLogService()
	}
	return c.logService
}

// SetOptionService 设置配置服务
func (c *Container) SetOptionService(s IOptionService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.optionService = s
}

// GetOptionService 获取配置服务
func (c *Container) GetOptionService() IOptionService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.optionService == nil {
		c.optionService = NewOptionService()
	}
	return c.optionService
}

// SetChannelBillingService 设置渠道余额服务
func (c *Container) SetChannelBillingService(s IChannelBillingService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channelBillingService = s
}

// GetChannelBillingService 获取渠道余额服务
func (c *Container) GetChannelBillingService() IChannelBillingService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.channelBillingService == nil {
		c.channelBillingService = NewChannelBillingService()
	}
	return c.channelBillingService
}

// SetChannelTestService 设置渠道测试服务
func (c *Container) SetChannelTestService(s IChannelTestService) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channelTestService = s
}

// GetChannelTestService 获取渠道测试服务
func (c *Container) GetChannelTestService() IChannelTestService {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.channelTestService == nil {
		c.channelTestService = NewChannelTestService()
	}
	return c.channelTestService
}

// 便捷函数，直接获取服务实例

// GetUserService 获取用户服务
func GetUserService() IUserService {
	return globalContainer.GetUserService()
}

// GetTokenService 获取令牌服务
func GetTokenService() ITokenService {
	return globalContainer.GetTokenService()
}

// GetChannelService 获取渠道服务
func GetChannelService() IChannelService {
	return globalContainer.GetChannelService()
}

// GetRedemptionService 获取兑换码服务
func GetRedemptionService() IRedemptionService {
	return globalContainer.GetRedemptionService()
}

// GetAuthService 获取认证服务
func GetAuthService() IAuthService {
	return globalContainer.GetAuthService()
}

// GetLogService 获取日志服务
func GetLogService() ILogService {
	return globalContainer.GetLogService()
}

// GetOptionService 获取配置服务
func GetOptionService() IOptionService {
	return globalContainer.GetOptionService()
}

// GetChannelBillingService 获取渠道余额服务
func GetChannelBillingService() IChannelBillingService {
	return globalContainer.GetChannelBillingService()
}

// GetChannelTestService 获取渠道测试服务
func GetChannelTestService() IChannelTestService {
	return globalContainer.GetChannelTestService()
}
