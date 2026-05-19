package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/songquanpeng/one-api/common/client"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/monitor"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

// 响应结构体定义
type OpenAISubscriptionResponse struct {
	Object             string  `json:"object"`
	HasPaymentMethod   bool    `json:"has_payment_method"`
	SoftLimitUSD       float64 `json:"soft_limit_usd"`
	HardLimitUSD       float64 `json:"hard_limit_usd"`
	SystemHardLimitUSD float64 `json:"system_hard_limit_usd"`
	AccessUntil        int64   `json:"access_until"`
}

type OpenAIUsageResponse struct {
	Object     string  `json:"object"`
	TotalUsage float64 `json:"total_usage"`
}

type OpenAICreditGrants struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalAvailable float64 `json:"total_available"`
}

type OpenAISBUsageResponse struct {
	Msg  string `json:"msg"`
	Data *struct {
		Credit string `json:"credit"`
	} `json:"data"`
}

type AIProxyUserOverviewResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
	Data      struct {
		TotalPoints float64 `json:"totalPoints"`
	} `json:"data"`
}

type API2GPTUsageResponse struct {
	Object         string  `json:"object"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
	TotalRemaining float64 `json:"total_remaining"`
}

type AIGC2DUsageResponse struct {
	Object         string  `json:"object"`
	TotalAvailable float64 `json:"total_available"`
	TotalGranted   float64 `json:"total_granted"`
	TotalUsed      float64 `json:"total_used"`
}

type SiliconFlowUsageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  bool   `json:"status"`
	Data    struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Balance      string `json:"balance"`
		TotalBalance string `json:"totalBalance"`
	} `json:"data"`
}

type DeepSeekUsageResponse struct {
	IsAvailable  bool `json:"is_available"`
	BalanceInfos []struct {
		Currency     string `json:"currency"`
		TotalBalance string `json:"total_balance"`
	} `json:"balance_infos"`
}

type OpenRouterResponse struct {
	Data struct {
		TotalCredits float64 `json:"total_credits"`
		TotalUsage   float64 `json:"total_usage"`
	} `json:"data"`
}

// ChannelBillingService 渠道余额服务
type ChannelBillingService struct{}

// NewChannelBillingService 创建渠道余额服务实例
func NewChannelBillingService() *ChannelBillingService {
	return &ChannelBillingService{}
}

// GetAuthHeader 获取认证头
func (s *ChannelBillingService) GetAuthHeader(token string) http.Header {
	h := http.Header{}
	h.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return h
}

// GetResponseBody 获取响应内容
func (s *ChannelBillingService) GetResponseBody(method, url string, channel *model.Channel, headers http.Header) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	for k := range headers {
		req.Header.Add(k, headers.Get(k))
	}
	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}

// UpdateChannelBalance 更新单个渠道余额
func (s *ChannelBillingService) UpdateChannelBalance(channel *model.Channel) (float64, error) {
	baseURL := channeltype.ChannelBaseURLs[channel.Type]
	if channel.GetBaseURL() == "" {
		channel.BaseURL = &baseURL
	}

	switch channel.Type {
	case channeltype.OpenAI, channeltype.Custom:
		if channel.GetBaseURL() != "" {
			baseURL = channel.GetBaseURL()
		}
		return s.updateOpenAIBalance(channel, baseURL)
	case channeltype.CloseAI:
		return s.updateCloseAIBalance(channel)
	case channeltype.OpenAISB:
		return s.updateOpenAISBBalance(channel)
	case channeltype.AIProxy:
		return s.updateAIProxyBalance(channel)
	case channeltype.API2GPT:
		return s.updateAPI2GPTBalance(channel)
	case channeltype.AIGC2D:
		return s.updateAIGC2DBalance(channel)
	case channeltype.SiliconFlow:
		return s.updateSiliconFlowBalance(channel)
	case channeltype.DeepSeek:
		return s.updateDeepSeekBalance(channel)
	case channeltype.OpenRouter:
		return s.updateOpenRouterBalance(channel)
	case channeltype.Azure:
		return 0, errors.New("尚未实现")
	default:
		return 0, errors.New("尚未实现")
	}
}

// UpdateAllChannelsBalance 更新所有渠道余额
func (s *ChannelBillingService) UpdateAllChannelsBalance() error {
	channels, err := model.GetAllChannels(0, 0, "all")
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if channel.Status != model.ChannelStatusEnabled {
			continue
		}
		if channel.Type != channeltype.OpenAI && channel.Type != channeltype.Custom {
			continue
		}
		balance, err := s.UpdateChannelBalance(channel)
		if err != nil {
			continue
		} else {
			if balance <= 0 {
				monitor.DisableChannel(channel.Id, channel.Name, "余额不足")
			}
		}
		time.Sleep(config.RequestInterval)
	}
	return nil
}

// AutomaticallyUpdateChannels 自动更新渠道余额
func (s *ChannelBillingService) AutomaticallyUpdateChannels(frequency int) {
	for {
		time.Sleep(time.Duration(frequency) * time.Minute)
		logger.SysLog("updating all channels")
		_ = s.UpdateAllChannelsBalance()
		logger.SysLog("channels update done")
	}
}

// updateOpenAIBalance 更新 OpenAI 余额
func (s *ChannelBillingService) updateOpenAIBalance(channel *model.Channel, baseURL string) (float64, error) {
	url := fmt.Sprintf("%s/v1/dashboard/billing/subscription", baseURL)
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	subscription := OpenAISubscriptionResponse{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	startDate := fmt.Sprintf("%s-01", now.Format("2006-01"))
	endDate := now.Format("2006-01-02")
	if !subscription.HasPaymentMethod {
		startDate = now.AddDate(0, 0, -100).Format("2006-01-02")
	}
	url = fmt.Sprintf("%s/v1/dashboard/billing/usage?start_date=%s&end_date=%s", baseURL, startDate, endDate)
	body, err = s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	usage := OpenAIUsageResponse{}
	err = json.Unmarshal(body, &usage)
	if err != nil {
		return 0, err
	}
	balance := subscription.HardLimitUSD - usage.TotalUsage/100
	channel.UpdateBalance(balance)
	return balance, nil
}

// updateCloseAIBalance 更新 CloseAI 余额
func (s *ChannelBillingService) updateCloseAIBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("%s/dashboard/billing/credit_grants", channel.GetBaseURL())
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenAICreditGrants{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalAvailable)
	return response.TotalAvailable, nil
}

// updateOpenAISBBalance 更新 OpenAISB 余额
func (s *ChannelBillingService) updateOpenAISBBalance(channel *model.Channel) (float64, error) {
	url := fmt.Sprintf("https://api.openai-sb.com/sb-api/user/status?api_key=%s", channel.Key)
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenAISBUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if response.Data == nil {
		return 0, errors.New(response.Msg)
	}
	balance, err := strconv.ParseFloat(response.Data.Credit, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

// updateAIProxyBalance 更新 AIProxy 余额
func (s *ChannelBillingService) updateAIProxyBalance(channel *model.Channel) (float64, error) {
	url := "https://aiproxy.io/api/report/getUserOverview"
	headers := http.Header{}
	headers.Add("Api-Key", channel.Key)
	body, err := s.GetResponseBody("GET", url, channel, headers)
	if err != nil {
		return 0, err
	}
	response := AIProxyUserOverviewResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if !response.Success {
		return 0, fmt.Errorf("code: %d, message: %s", response.ErrorCode, response.Message)
	}
	channel.UpdateBalance(response.Data.TotalPoints)
	return response.Data.TotalPoints, nil
}

// updateAPI2GPTBalance 更新 API2GPT 余额
func (s *ChannelBillingService) updateAPI2GPTBalance(channel *model.Channel) (float64, error) {
	url := "https://api.api2gpt.com/dashboard/billing/credit_grants"
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := API2GPTUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalRemaining)
	return response.TotalRemaining, nil
}

// updateAIGC2DBalance 更新 AIGC2D 余额
func (s *ChannelBillingService) updateAIGC2DBalance(channel *model.Channel) (float64, error) {
	url := "https://api.aigc2d.com/dashboard/billing/credit_grants"
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := AIGC2DUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(response.TotalAvailable)
	return response.TotalAvailable, nil
}

// updateSiliconFlowBalance 更新 SiliconFlow 余额
func (s *ChannelBillingService) updateSiliconFlowBalance(channel *model.Channel) (float64, error) {
	url := "https://api.siliconflow.cn/v1/user/info"
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := SiliconFlowUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	if response.Code != 20000 {
		return 0, fmt.Errorf("code: %d, message: %s", response.Code, response.Message)
	}
	balance, err := strconv.ParseFloat(response.Data.TotalBalance, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

// updateDeepSeekBalance 更新 DeepSeek 余额
func (s *ChannelBillingService) updateDeepSeekBalance(channel *model.Channel) (float64, error) {
	url := "https://api.deepseek.com/user/balance"
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := DeepSeekUsageResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	index := -1
	for i, balanceInfo := range response.BalanceInfos {
		if balanceInfo.Currency == "CNY" {
			index = i
			break
		}
	}
	if index == -1 {
		return 0, errors.New("currency CNY not found")
	}
	balance, err := strconv.ParseFloat(response.BalanceInfos[index].TotalBalance, 64)
	if err != nil {
		return 0, err
	}
	channel.UpdateBalance(balance)
	return balance, nil
}

// updateOpenRouterBalance 更新 OpenRouter 余额
func (s *ChannelBillingService) updateOpenRouterBalance(channel *model.Channel) (float64, error) {
	url := "https://openrouter.ai/api/v1/credits"
	body, err := s.GetResponseBody("GET", url, channel, s.GetAuthHeader(channel.Key))
	if err != nil {
		return 0, err
	}
	response := OpenRouterResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}
	balance := response.Data.TotalCredits - response.Data.TotalUsage
	channel.UpdateBalance(balance)
	return balance, nil
}
