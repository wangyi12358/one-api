package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrOAuthDisabled         = errors.New("管理员未开启此登录方式")
	ErrOAuthStateInvalid     = errors.New("state is empty or not same")
	ErrUserBanned            = errors.New("用户已被封禁")
	ErrOAuthRegisterDisabled = errors.New("管理员关闭了新用户注册")
	ErrAccountAlreadyBound   = errors.New("该账户已被绑定")
	ErrOAuthCodeInvalid      = errors.New("无效的参数")
	ErrOAuthServerConnect    = errors.New("无法连接至服务器，请稍后重试")
)

// OAuthProvider OAuth 提供商类型
type OAuthProvider string

const (
	ProviderGitHub OAuthProvider = "github"
	ProviderWeChat OAuthProvider = "wechat"
	ProviderLark   OAuthProvider = "lark"
	ProviderOIDC   OAuthProvider = "oidc"
)

// OAuthUserInfo OAuth 用户信息
type OAuthUserInfo struct {
	Provider    OAuthProvider
	ProviderId  string
	Username    string
	DisplayName string
	Email       string
}

// AuthService 认证服务
type AuthService struct{}

// NewAuthService 创建认证服务实例
func NewAuthService() *AuthService {
	return &AuthService{}
}

// GitHubOAuthResponse GitHub OAuth 响应
type GitHubOAuthResponse struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// GitHubUser GitHub 用户信息
type GitHubUser struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// GetGitHubUserInfoByCode 通过 code 获取 GitHub 用户信息
func (s *AuthService) GetGitHubUserInfoByCode(code string) (*OAuthUserInfo, error) {
	if code == "" {
		return nil, ErrOAuthCodeInvalid
	}

	values := map[string]string{
		"client_id":     config.GitHubClientId,
		"client_secret": config.GitHubClientSecret,
		"code":          code,
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, ErrOAuthServerConnect
	}
	defer res.Body.Close()

	var oAuthResponse GitHubOAuthResponse
	err = json.NewDecoder(res.Body).Decode(&oAuthResponse)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", oAuthResponse.AccessToken))

	res2, err := client.Do(req)
	if err != nil {
		logger.SysLog(err.Error())
		return nil, ErrOAuthServerConnect
	}
	defer res2.Body.Close()

	var githubUser GitHubUser
	err = json.NewDecoder(res2.Body).Decode(&githubUser)
	if err != nil {
		return nil, err
	}

	if githubUser.Login == "" {
		return nil, errors.New("返回值非法，用户字段为空，请稍后重试")
	}

	return &OAuthUserInfo{
		Provider:    ProviderGitHub,
		ProviderId:  githubUser.Login,
		Username:    "github_" + githubUser.Login,
		DisplayName: githubUser.Name,
		Email:       githubUser.Email,
	}, nil
}

// OAuthLogin OAuth 登录
func (s *AuthService) OAuthLogin(ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo) (*model.User, error) {
	// 1. 检查 OAuth 是否启用
	if !s.isOAuthEnabled(provider) {
		return nil, ErrOAuthDisabled
	}

	// 2. 查找或创建用户
	user, err := s.findOrCreateUser(ctx, provider, userInfo)
	if err != nil {
		return nil, err
	}

	// 3. 检查用户状态
	if user.Status != model.UserStatusEnabled {
		return nil, ErrUserBanned
	}

	return user, nil
}

// OAuthBind OAuth 绑定账号
func (s *AuthService) OAuthBind(provider OAuthProvider, userInfo *OAuthUserInfo, userId int) error {
	// 1. 检查 OAuth 是否启用
	if !s.isOAuthEnabled(provider) {
		return ErrOAuthDisabled
	}

	// 2. 检查是否已被绑定
	if s.isProviderIdTaken(provider, userInfo.ProviderId) {
		return ErrAccountAlreadyBound
	}

	// 3. 获取用户并绑定
	user := &model.User{Id: userId}
	if err := user.FillUserById(); err != nil {
		return err
	}

	// 4. 设置 provider id
	s.setProviderId(user, provider, userInfo.ProviderId)

	// 5. 保存更新
	return user.Update(false)
}

// GenerateOAuthState 生成 OAuth state
func (s *AuthService) GenerateOAuthState() string {
	return random.GetRandomString(12)
}

// isOAuthEnabled 检查 OAuth 是否启用
func (s *AuthService) isOAuthEnabled(provider OAuthProvider) bool {
	switch provider {
	case ProviderGitHub:
		return config.GitHubOAuthEnabled
	case ProviderWeChat:
		return config.WeChatAuthEnabled
	case ProviderLark:
		return config.LarkClientId != "" && config.LarkClientSecret != ""
	case ProviderOIDC:
		return config.OidcEnabled
	default:
		return false
	}
}

// findOrCreateUser 查找或创建用户
func (s *AuthService) findOrCreateUser(ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo) (*model.User, error) {
	// 1. 查找已有用户
	if s.isProviderIdTaken(provider, userInfo.ProviderId) {
		user, err := s.getUserByProviderId(provider, userInfo.ProviderId)
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	// 2. 检查注册是否启用
	if !config.RegisterEnabled {
		return nil, ErrOAuthRegisterDisabled
	}

	// 3. 创建新用户
	user := &model.User{
		Username:    userInfo.Username,
		DisplayName: userInfo.DisplayName,
		Email:       userInfo.Email,
		Role:        model.RoleCommonUser,
		Status:      model.UserStatusEnabled,
	}

	// 设置 provider id
	s.setProviderId(user, provider, userInfo.ProviderId)

	// 生成随机密码
	user.Password = random.GetRandomString(16)

	if err := user.Insert(ctx, 0); err != nil {
		return nil, err
	}

	return user, nil
}

// isProviderIdTaken 检查 provider id 是否已被使用
func (s *AuthService) isProviderIdTaken(provider OAuthProvider, providerId string) bool {
	switch provider {
	case ProviderGitHub:
		return model.IsGitHubIdAlreadyTaken(providerId)
	case ProviderWeChat:
		return model.IsWeChatIdAlreadyTaken(providerId)
	case ProviderLark:
		return model.IsLarkIdAlreadyTaken(providerId)
	case ProviderOIDC:
		return model.IsOidcIdAlreadyTaken(providerId)
	default:
		return false
	}
}

// getUserByProviderId 通过 provider id 获取用户
func (s *AuthService) getUserByProviderId(provider OAuthProvider, providerId string) (*model.User, error) {
	user := &model.User{}
	switch provider {
	case ProviderGitHub:
		user.GitHubId = providerId
		err := user.FillUserByGitHubId()
		return user, err
	case ProviderWeChat:
		user.WeChatId = providerId
		err := user.FillUserByWeChatId()
		return user, err
	case ProviderLark:
		user.LarkId = providerId
		err := user.FillUserByLarkId()
		return user, err
	case ProviderOIDC:
		user.OidcId = providerId
		err := user.FillUserByOidcId()
		return user, err
	default:
		return nil, errors.New("unsupported provider")
	}
}

// setProviderId 设置 provider id
func (s *AuthService) setProviderId(user *model.User, provider OAuthProvider, providerId string) {
	switch provider {
	case ProviderGitHub:
		user.GitHubId = providerId
	case ProviderWeChat:
		user.WeChatId = providerId
	case ProviderLark:
		user.LarkId = providerId
	case ProviderOIDC:
		user.OidcId = providerId
	}
}

// GetMaxUserId 获取最大用户 ID（用于生成用户名）
func (s *AuthService) GetMaxUserId() int {
	return model.GetMaxUserId()
}
