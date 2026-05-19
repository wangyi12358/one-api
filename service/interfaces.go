package service

import (
	"context"

	"github.com/songquanpeng/one-api/model"
)

// UserService 用户服务接口
type IUserService interface {
	// Register 用户注册
	Register(ctx context.Context, req *RegisterRequest) error
	// ManageUser 管理用户（禁用/启用/删除/提升/降级）
	ManageUser(req *ManageUserRequest, operatorRole int) (*model.User, error)
	// UpdateUser 更新用户信息
	UpdateUser(params *UserUpdateParams, operatorId int, operatorRole int) error
	// GenerateAccessToken 生成访问令牌
	GenerateAccessToken(userId int) (string, error)
	// DeleteUser 删除用户
	DeleteUser(userId int, operatorRole int) error
	// DeleteSelf 用户自行删除账户
	DeleteSelf(userId int) error
	// EmailBind 绑定邮箱
	EmailBind(userId int, email, code string) error
	// TopUp 用户充值
	TopUp(ctx context.Context, userId int, key string) (int64, error)
	// AdminTopUp 管理员充值
	AdminTopUp(ctx context.Context, userId int, quota int, remark string) error
}

// TokenService 令牌服务接口
type ITokenService interface {
	// CreateToken 创建令牌
	CreateToken(userId int, req *CreateTokenRequest) (*model.Token, error)
	// UpdateToken 更新令牌
	UpdateToken(userId int, req *UpdateTokenRequest, statusOnly bool) (*model.Token, error)
	// DeleteToken 删除令牌
	DeleteToken(id, userId int) error
	// GetToken 获取令牌
	GetToken(id, userId int) (*model.Token, error)
	// GetAllTokens 获取用户所有令牌
	GetAllTokens(userId, startIdx, num int, order string) ([]*model.Token, error)
	// SearchTokens 搜索令牌
	SearchTokens(userId int, keyword string) ([]*model.Token, error)
	// PreConsumeTokenQuota 预扣费令牌额度
	PreConsumeTokenQuota(tokenId int, quota int64) error
	// PostConsumeTokenQuota 后处理令牌额度
	PostConsumeTokenQuota(tokenId int, quota int64) error
}

// ChannelService 渠道服务接口
type IChannelService interface {
	// CreateChannel 创建渠道（支持多 key 批量创建）
	CreateChannel(req *CreateChannelRequest) error
	// UpdateChannel 更新渠道
	UpdateChannel(channel *model.Channel) error
	// DeleteChannel 删除渠道
	DeleteChannel(id int) error
	// DeleteDisabledChannels 删除禁用的渠道
	DeleteDisabledChannels() (int64, error)
	// GetChannel 获取渠道
	GetChannel(id int) (*model.Channel, error)
	// GetAllChannels 获取所有渠道
	GetAllChannels(startIdx, num int, status string) ([]*model.Channel, error)
	// SearchChannels 搜索渠道
	SearchChannels(keyword string) ([]*model.Channel, error)
}

// RedemptionService 兑换码服务接口
type IRedemptionService interface {
	// CreateRedemptions 批量创建兑换码
	CreateRedemptions(userId int, req *CreateRedemptionRequest) ([]string, error)
	// UpdateRedemption 更新兑换码
	UpdateRedemption(req *UpdateRedemptionRequest, statusOnly bool) (*model.Redemption, error)
	// DeleteRedemption 删除兑换码
	DeleteRedemption(id int) error
	// GetRedemption 获取兑换码
	GetRedemption(id int) (*model.Redemption, error)
	// GetAllRedemptions 获取所有兑换码
	GetAllRedemptions(startIdx, num int) ([]*model.Redemption, error)
	// SearchRedemptions 搜索兑换码
	SearchRedemptions(keyword string) ([]*model.Redemption, error)
	// RedeemCode 兑换码兑换
	RedeemCode(ctx context.Context, userId int, key string) (int64, error)
}

// AuthService 认证服务接口
type IAuthService interface {
	// GetGitHubUserInfoByCode 通过 code 获取 GitHub 用户信息
	GetGitHubUserInfoByCode(code string) (*OAuthUserInfo, error)
	// OAuthLogin OAuth 登录
	OAuthLogin(ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo) (*model.User, error)
	// OAuthBind OAuth 绑定账号
	OAuthBind(provider OAuthProvider, userInfo *OAuthUserInfo, userId int) error
	// GenerateOAuthState 生成 OAuth state
	GenerateOAuthState() string
}

// LogService 日志服务接口
type ILogService interface {
	// GetAllLogs 获取所有日志
	GetAllLogs(query *LogQuery) ([]*model.Log, error)
	// GetUserLogs 获取用户日志
	GetUserLogs(userId int, query *LogQuery) ([]*model.Log, error)
	// SearchAllLogs 搜索所有日志
	SearchAllLogs(keyword string) ([]*model.Log, error)
	// SearchUserLogs 搜索用户日志
	SearchUserLogs(userId int, keyword string) ([]*model.Log, error)
	// GetLogsStat 获取日志统计
	GetLogsStat(query *LogQuery) (*LogStat, error)
	// DeleteHistoryLogs 删除历史日志
	DeleteHistoryLogs(targetTimestamp int64) (int64, error)
}

// OptionService 配置服务接口
type IOptionService interface {
	// GetPublicOptions 获取公开配置项
	GetPublicOptions() []*model.Option
	// UpdateOption 更新配置项
	UpdateOption(key, value string) error
}

// ChannelBillingService 渠道余额服务接口
type IChannelBillingService interface {
	// UpdateChannelBalance 更新单个渠道余额
	UpdateChannelBalance(channel *model.Channel) (float64, error)
	// UpdateAllChannelsBalance 更新所有渠道余额
	UpdateAllChannelsBalance() error
	// AutomaticallyUpdateChannels 自动更新渠道余额
	AutomaticallyUpdateChannels(frequency int)
}

// ChannelTestService 渠道测试服务接口
type IChannelTestService interface {
	// TestChannel 测试单个渠道
	TestChannel(ctx context.Context, channelId int, modelName string) (*TestResult, error)
	// TestAllChannels 测试所有渠道
	TestAllChannels(ctx context.Context, notify bool, scope string) error
	// AutomaticallyTestChannels 自动测试渠道
	AutomaticallyTestChannels(frequency int)
}
