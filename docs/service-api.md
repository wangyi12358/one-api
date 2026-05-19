# One-API Service 层接口文档

## 目录

1. [架构概述](#架构概述)
2. [服务容器](#服务容器)
3. [接口定义](#接口定义)
   - [IUserService](#iuserservice)
   - [ITokenService](#itokenservice)
   - [IChannelService](#ichannelservice)
   - [IRedemptionService](#iredemptionservice)
   - [IAuthService](#iauthservice)
   - [ILogService](#ilogservice)
   - [IOptionService](#ioptionservice)
   - [IChannelBillingService](#ichannelbillingservice)
   - [IChannelTestService](#ichanneltestservice)
4. [使用示例](#使用示例)
5. [测试指南](#测试指南)

---

## 架构概述

### 分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                        Router 层                            │
│                    (路由定义和中间件)                         │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                      Controller 层                          │
│              (请求解析、参数验证、响应构造)                    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                       Service 层                            │
│         (业务逻辑、数据处理、外部服务调用)                     │
│                                                             │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐          │
│  │ UserSvc │ │TokenSvc │ │ChannelSvc│ │AuthSvc │ ...      │
│  └─────────┘ └─────────┘ └─────────┘ └─────────┘          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                        Model 层                             │
│              (数据模型定义、数据库 CRUD 操作)                  │
└─────────────────────────────────────────────────────────────┘
```

### 设计原则

1. **单一职责**：每个服务负责一个特定的业务领域
2. **依赖倒置**：Controller 依赖接口而非具体实现
3. **开闭原则**：通过接口扩展功能，无需修改现有代码
4. **接口隔离**：接口保持精简，只暴露必要的方法

---

## 服务容器

### 获取服务实例

```go
import "github.com/songquanpeng/one-api/service"

// 获取用户服务
userService := service.GetUserService()

// 获取令牌服务
tokenService := service.GetTokenService()

// 获取渠道服务
channelService := service.GetChannelService()

// 获取兑换码服务
redemptionService := service.GetRedemptionService()

// 获取认证服务
authService := service.GetAuthService()

// 获取日志服务
logService := service.GetLogService()

// 获取配置服务
optionService := service.GetOptionService()

// 获取渠道余额服务
channelBillingService := service.GetChannelBillingService()

// 获取渠道测试服务
channelTestService := service.GetChannelTestService()
```

### 依赖注入（用于测试）

```go
// 获取服务容器
container := service.GetContainer()

// 注入 mock 实现
container.SetUserService(mockUserService)
```

---

## 接口定义

### IUserService

用户服务接口，负责用户注册、管理、认证等功能。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `Register` | `ctx context.Context, req *RegisterRequest` | `error` | 用户注册 |
| `ManageUser` | `req *ManageUserRequest, operatorRole int` | `(*model.User, error)` | 管理用户（禁用/启用/删除/提升/降级） |
| `UpdateUser` | `params *UserUpdateParams, operatorId int, operatorRole int` | `error` | 更新用户信息 |
| `GenerateAccessToken` | `userId int` | `(string, error)` | 生成访问令牌 |
| `DeleteUser` | `userId int, operatorRole int` | `error` | 删除用户 |
| `DeleteSelf` | `userId int` | `error` | 用户自行删除账户 |
| `EmailBind` | `userId int, email, code string` | `error` | 绑定邮箱 |
| `TopUp` | `ctx context.Context, userId int, key string` | `(int64, error)` | 用户充值 |
| `AdminTopUp` | `ctx context.Context, userId int, quota int, remark string` | `error` | 管理员充值 |

#### 请求结构体

```go
// RegisterRequest 注册请求
type RegisterRequest struct {
    Username         string `json:"username"`
    Password         string `json:"password"`
    DisplayName      string `json:"display_name"`
    Email            string `json:"email"`
    VerificationCode string `json:"verification_code"`
    AffCode          string `json:"aff_code"`
}

// ManageUserRequest 管理用户请求
type ManageUserRequest struct {
    Username string       `json:"username"`
    Action   ManageAction `json:"action"` // disable/enable/delete/promote/demote
}

// UserUpdateParams 用户更新参数
type UserUpdateParams struct {
    Id          int    `json:"id"`
    Username    string `json:"username"`
    Password    string `json:"password"`
    DisplayName string `json:"display_name"`
    Role        int    `json:"role"`
    Quota       int64  `json:"quota"`
}
```

#### 使用示例

```go
func Register(c *gin.Context) {
    var req service.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiErrorMsg(c, "参数错误")
        return
    }

    userService := service.GetUserService()
    err := userService.Register(c.Request.Context(), &req)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, nil)
}
```

---

### ITokenService

令牌服务接口，负责令牌的创建、更新、删除和额度管理。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CreateToken` | `userId int, req *CreateTokenRequest` | `(*model.Token, error)` | 创建令牌 |
| `UpdateToken` | `userId int, req *UpdateTokenRequest, statusOnly bool` | `(*model.Token, error)` | 更新令牌 |
| `DeleteToken` | `id, userId int` | `error` | 删除令牌 |
| `GetToken` | `id, userId int` | `(*model.Token, error)` | 获取令牌 |
| `GetAllTokens` | `userId, startIdx, num int, order string` | `([]*model.Token, error)` | 获取用户所有令牌 |
| `SearchTokens` | `userId int, keyword string` | `([]*model.Token, error)` | 搜索令牌 |
| `PreConsumeTokenQuota` | `tokenId int, quota int64` | `error` | 预扣费令牌额度 |
| `PostConsumeTokenQuota` | `tokenId int, quota int64` | `error` | 后处理令牌额度 |

#### 请求结构体

```go
// CreateTokenRequest 创建令牌请求
type CreateTokenRequest struct {
    Name           string  `json:"name"`
    ExpiredTime    int64   `json:"expired_time"`
    RemainQuota    int64   `json:"remain_quota"`
    UnlimitedQuota bool    `json:"unlimited_quota"`
    Models         *string `json:"models"`
    Subnet         *string `json:"subnet"`
}

// UpdateTokenRequest 更新令牌请求
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
```

#### 使用示例

```go
func AddToken(c *gin.Context) {
    var req service.CreateTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiError(c, err)
        return
    }

    userId := c.GetInt(ctxkey.Id)
    tokenService := service.GetTokenService()
    token, err := tokenService.CreateToken(userId, &req)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, token)
}
```

---

### IChannelService

渠道服务接口，负责渠道的创建、更新、删除和查询。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CreateChannel` | `req *CreateChannelRequest` | `error` | 创建渠道（支持多 key 批量创建） |
| `UpdateChannel` | `channel *model.Channel` | `error` | 更新渠道 |
| `DeleteChannel` | `id int` | `error` | 删除渠道 |
| `DeleteDisabledChannels` | - | `(int64, error)` | 删除禁用的渠道 |
| `GetChannel` | `id int` | `(*model.Channel, error)` | 获取渠道 |
| `GetAllChannels` | `startIdx, num int, status string` | `([]*model.Channel, error)` | 获取所有渠道 |
| `SearchChannels` | `keyword string` | `([]*model.Channel, error)` | 搜索渠道 |

#### 请求结构体

```go
// CreateChannelRequest 创建渠道请求
type CreateChannelRequest struct {
    Name         string  `json:"name"`
    Type         int     `json:"type"`
    Key          string  `json:"key"`           // 支持多 key，用换行符分隔
    BaseURL      *string `json:"base_url"`
    Other        *string `json:"other"`
    Models       string  `json:"models"`
    Group        string  `json:"group"`
    Priority     *int64  `json:"priority"`
    Weight       *uint   `json:"weight"`
    Balance      float64 `json:"balance"`
    ModelMapping *string `json:"model_mapping"`
    Config       string  `json:"config"`
    SystemPrompt *string `json:"system_prompt"`
}
```

#### 使用示例

```go
func AddChannel(c *gin.Context) {
    var req service.CreateChannelRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiError(c, err)
        return
    }

    channelService := service.GetChannelService()
    err := channelService.CreateChannel(&req)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, nil)
}
```

---

### IRedemptionService

兑换码服务接口，负责兑换码的创建、更新、删除和兑换。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CreateRedemptions` | `userId int, req *CreateRedemptionRequest` | `([]string, error)` | 批量创建兑换码 |
| `UpdateRedemption` | `req *UpdateRedemptionRequest, statusOnly bool` | `(*model.Redemption, error)` | 更新兑换码 |
| `DeleteRedemption` | `id int` | `error` | 删除兑换码 |
| `GetRedemption` | `id int` | `(*model.Redemption, error)` | 获取兑换码 |
| `GetAllRedemptions` | `startIdx, num int` | `([]*model.Redemption, error)` | 获取所有兑换码 |
| `SearchRedemptions` | `keyword string` | `([]*model.Redemption, error)` | 搜索兑换码 |
| `RedeemCode` | `ctx context.Context, userId int, key string` | `(int64, error)` | 兑换码兑换 |

#### 请求结构体

```go
// CreateRedemptionRequest 创建兑换码请求
type CreateRedemptionRequest struct {
    Name  string `json:"name"`   // 长度 1-20
    Quota int64  `json:"quota"`
    Count int    `json:"count"`  // 1-100
}

// UpdateRedemptionRequest 更新兑换码请求
type UpdateRedemptionRequest struct {
    Id     int    `json:"id"`
    Name   string `json:"name"`
    Status int    `json:"status"`
    Quota  int64  `json:"quota"`
}
```

#### 使用示例

```go
func AddRedemption(c *gin.Context) {
    var req service.CreateRedemptionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiError(c, err)
        return
    }

    userId := c.GetInt(ctxkey.Id)
    redemptionService := service.GetRedemptionService()
    keys, err := redemptionService.CreateRedemptions(userId, &req)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, keys)
}
```

---

### IAuthService

认证服务接口，负责 OAuth 登录和账号绑定。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetGitHubUserInfoByCode` | `code string` | `(*OAuthUserInfo, error)` | 通过 code 获取 GitHub 用户信息 |
| `OAuthLogin` | `ctx context.Context, provider OAuthProvider, userInfo *OAuthUserInfo` | `(*model.User, error)` | OAuth 登录 |
| `OAuthBind` | `provider OAuthProvider, userInfo *OAuthUserInfo, userId int` | `error` | OAuth 绑定账号 |
| `GenerateOAuthState` | - | `string` | 生成 OAuth state |

#### 类型定义

```go
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
```

#### 使用示例

```go
func GitHubOAuth(c *gin.Context) {
    code := c.Query("code")
    authService := service.GetAuthService()

    // 获取用户信息
    userInfo, err := authService.GetGitHubUserInfoByCode(code)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    // 登录
    user, err := authService.OAuthLogin(c.Request.Context(), service.ProviderGitHub, userInfo)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    // 设置登录 session
    controller.SetupLogin(user, c)
}
```

---

### ILogService

日志服务接口，负责日志的查询、统计和清理。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetAllLogs` | `query *LogQuery` | `([]*model.Log, error)` | 获取所有日志 |
| `GetUserLogs` | `userId int, query *LogQuery` | `([]*model.Log, error)` | 获取用户日志 |
| `SearchAllLogs` | `keyword string` | `([]*model.Log, error)` | 搜索所有日志 |
| `SearchUserLogs` | `userId int, keyword string` | `([]*model.Log, error)` | 搜索用户日志 |
| `GetLogsStat` | `query *LogQuery` | `(*LogStat, error)` | 获取日志统计 |
| `DeleteHistoryLogs` | `targetTimestamp int64` | `(int64, error)` | 删除历史日志 |

#### 请求结构体

```go
// LogQuery 日志查询参数
type LogQuery struct {
    Page           int
    PageSize       int
    LogType        int
    StartTimestamp  int64
    EndTimestamp    int64
    ModelName      string
    Username       string
    TokenName      string
    Channel        int
}

// LogStat 日志统计结果
type LogStat struct {
    Quota int64 `json:"quota"`
}
```

#### 使用示例

```go
func GetAllLogs(c *gin.Context) {
    query := &service.LogQuery{
        Page:           0,
        PageSize:       config.ItemsPerPage,
        LogType:        0,
        StartTimestamp:  0,
        EndTimestamp:    0,
    }

    logService := service.GetLogService()
    logs, err := logService.GetAllLogs(query)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, logs)
}
```

---

### IOptionService

配置服务接口，负责系统配置的获取和更新。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `GetPublicOptions` | - | `[]*model.Option` | 获取公开配置项（过滤敏感信息） |
| `UpdateOption` | `key, value string` | `error` | 更新配置项 |

#### 使用示例

```go
func GetOptions(c *gin.Context) {
    optionService := service.GetOptionService()
    options := optionService.GetPublicOptions()
    common.ApiSuccess(c, options)
}

func UpdateOption(c *gin.Context) {
    var option model.Option
    if err := c.ShouldBindJSON(&option); err != nil {
        common.ApiErrorMsg(c, "参数错误")
        return
    }

    optionService := service.GetOptionService()
    err := optionService.UpdateOption(option.Key, option.Value)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, nil)
}
```

---

### IChannelBillingService

渠道余额服务接口，负责渠道余额的查询和更新。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `UpdateChannelBalance` | `channel *model.Channel` | `(float64, error)` | 更新单个渠道余额 |
| `UpdateAllChannelsBalance` | - | `error` | 更新所有渠道余额 |
| `AutomaticallyUpdateChannels` | `frequency int` | - | 自动更新渠道余额 |

#### 使用示例

```go
func UpdateChannelBalance(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    channel, err := model.GetChannelById(id, true)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    channelBillingService := service.GetChannelBillingService()
    balance, err := channelBillingService.UpdateChannelBalance(channel)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "balance": balance,
    })
}
```

---

### IChannelTestService

渠道测试服务接口，负责渠道的测试和监控。

#### 方法列表

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `TestChannel` | `ctx context.Context, channelId int, modelName string` | `(*TestResult, error)` | 测试单个渠道 |
| `TestAllChannels` | `ctx context.Context, notify bool, scope string` | `error` | 测试所有渠道 |
| `AutomaticallyTestChannels` | `frequency int` | - | 自动测试渠道 |

#### 返回结构体

```go
// TestResult 测试结果
type TestResult struct {
    Success         bool    `json:"success"`
    Message         string  `json:"message"`
    Time            float64 `json:"time"`
    ModelName       string  `json:"modelName"`
    ResponseMessage string  `json:"responseMessage,omitempty"`
}
```

#### 使用示例

```go
func TestChannel(c *gin.Context) {
    id, _ := strconv.Atoi(c.Param("id"))
    modelName := c.Query("model")

    channelTestService := service.GetChannelTestService()
    result, err := channelTestService.TestChannel(c.Request.Context(), id, modelName)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success":   result.Success,
        "message":   result.Message,
        "time":      result.Time,
        "modelName": result.ModelName,
    })
}
```

---

## 使用示例

### 完整的 Controller 示例

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/songquanpeng/one-api/common"
    "github.com/songquanpeng/one-api/service"
)

func Register(c *gin.Context) {
    var req service.RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        common.ApiErrorMsg(c, "参数错误")
        return
    }

    userService := service.GetUserService()
    err := userService.Register(c.Request.Context(), &req)
    if err != nil {
        common.ApiError(c, err)
        return
    }

    common.ApiSuccess(c, nil)
}
```

### 依赖注入示例（用于测试）

```go
package service_test

import (
    "testing"
    "github.com/songquanpeng/one-api/service"
    "github.com/stretchr/testify/assert"
)

// MockUserService 模拟用户服务
type MockUserService struct {
    RegisterFunc func(ctx context.Context, req *service.RegisterRequest) error
}

func (m *MockUserService) Register(ctx context.Context, req *service.RegisterRequest) error {
    return m.RegisterFunc(ctx, req)
}

// ... 实现其他接口方法

func TestRegister(t *testing.T) {
    // 注入 mock
    container := service.GetContainer()
    mockService := &MockUserService{
        RegisterFunc: func(ctx context.Context, req *service.RegisterRequest) error {
            return nil
        },
    }
    container.SetUserService(mockService)

    // 测试
    userService := service.GetUserService()
    err := userService.Register(context.Background(), &service.RegisterRequest{
        Username: "test",
        Password: "password",
    })
    assert.NoError(t, err)
}
```

---

## 测试指南

### 运行测试

```bash
# 运行所有服务测试
go test ./service/... -v

# 运行特定测试
go test ./service/... -run TestTokenService_validateToken -v

# 查看覆盖率
go test ./service/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### 编写测试

```go
package service

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestTokenService_validateToken(t *testing.T) {
    service := NewTokenService()

    tests := []struct {
        name    string
        tokenName string
        subnet  *string
        wantErr error
    }{
        {
            name:    "valid name without subnet",
            tokenName: "test-token",
            subnet:  nil,
            wantErr: nil,
        },
        {
            name:    "name too long",
            tokenName: "this-is-a-very-long-token-name-that-exceeds-thirty-characters",
            subnet:  nil,
            wantErr: ErrTokenNameTooLong,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.validateToken(tt.tokenName, tt.subnet)
            if tt.wantErr != nil {
                assert.ErrorIs(t, err, tt.wantErr)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## 错误处理

### 统一错误定义

每个服务都在文件开头定义了业务错误：

```go
// service/user.go
var (
    ErrRegisterDisabled       = errors.New("管理员关闭了新用户注册")
    ErrPasswordRegisterDisabled = errors.New("管理员关闭了通过密码进行注册")
    ErrUserNotFound           = errors.New("用户不存在")
    // ...
)
```

### 错误使用示例

```go
func (s *UserService) Register(ctx context.Context, req *RegisterRequest) error {
    if !config.RegisterEnabled {
        return ErrRegisterDisabled
    }
    // ...
}
```

### Controller 错误处理

```go
func Register(c *gin.Context) {
    userService := service.GetUserService()
    err := userService.Register(c.Request.Context(), &req)
    if err != nil {
        // 使用统一的错误响应
        common.ApiError(c, err)
        return
    }
    common.ApiSuccess(c, nil)
}
```
