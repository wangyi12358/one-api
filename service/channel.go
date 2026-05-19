package service

import (
	"strings"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/model"
)

// CreateChannelRequest 创建渠道请求参数
type CreateChannelRequest struct {
	Name         string  `json:"name"`
	Type         int     `json:"type"`
	Key          string  `json:"key"`
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

// ChannelService 渠道服务
type ChannelService struct{}

// NewChannelService 创建渠道服务实例
func NewChannelService() *ChannelService {
	return &ChannelService{}
}

// CreateChannel 创建渠道（支持多 key 批量创建）
func (s *ChannelService) CreateChannel(req *CreateChannelRequest) error {
	// 1. 构建基础渠道对象
	baseChannel := model.Channel{
		Name:         req.Name,
		Type:         req.Type,
		Key:          req.Key,
		BaseURL:      req.BaseURL,
		Other:        req.Other,
		Models:       req.Models,
		Group:        req.Group,
		Priority:     req.Priority,
		Weight:       req.Weight,
		Balance:      req.Balance,
		ModelMapping: req.ModelMapping,
		Config:       req.Config,
		SystemPrompt: req.SystemPrompt,
		CreatedTime:  helper.GetTimestamp(),
	}

	// 2. 拆分多 key
	keys := strings.Split(req.Key, "\n")
	channels := make([]model.Channel, 0, len(keys))
	for _, key := range keys {
		if key == "" {
			continue
		}
		localChannel := baseChannel
		localChannel.Key = key
		channels = append(channels, localChannel)
	}

	// 3. 批量插入
	return model.BatchInsertChannels(channels)
}

// UpdateChannel 更新渠道
func (s *ChannelService) UpdateChannel(channel *model.Channel) error {
	return channel.Update()
}

// DeleteChannel 删除渠道
func (s *ChannelService) DeleteChannel(id int) error {
	channel := model.Channel{Id: id}
	return channel.Delete()
}

// DeleteDisabledChannels 删除禁用的渠道
func (s *ChannelService) DeleteDisabledChannels() (int64, error) {
	return model.DeleteDisabledChannel()
}

// GetChannel 获取渠道
func (s *ChannelService) GetChannel(id int) (*model.Channel, error) {
	return model.GetChannelById(id, false)
}

// GetAllChannels 获取所有渠道
func (s *ChannelService) GetAllChannels(startIdx, num int, status string) ([]*model.Channel, error) {
	return model.GetAllChannels(startIdx, num, status)
}

// SearchChannels 搜索渠道
func (s *ChannelService) SearchChannels(keyword string) ([]*model.Channel, error) {
	return model.SearchChannels(keyword)
}
