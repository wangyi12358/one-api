package service

import (
	"errors"

	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrTargetTimestampRequired = errors.New("target timestamp is required")
)

// LogQuery 日志查询参数
type LogQuery struct {
	Page           int
	PageSize       int
	LogType        int
	StartTimestamp int64
	EndTimestamp   int64
	ModelName      string
	Username       string
	TokenName      string
	Channel        int
}

// LogStat 日志统计结果
type LogStat struct {
	Quota int64 `json:"quota"`
}

// LogService 日志服务
type LogService struct{}

// NewLogService 创建日志服务实例
func NewLogService() *LogService {
	return &LogService{}
}

// GetAllLogs 获取所有日志
func (s *LogService) GetAllLogs(query *LogQuery) ([]*model.Log, error) {
	return model.GetAllLogs(
		query.LogType,
		query.StartTimestamp,
		query.EndTimestamp,
		query.ModelName,
		query.Username,
		query.TokenName,
		query.Page*query.PageSize,
		query.PageSize,
		query.Channel,
	)
}

// GetUserLogs 获取用户日志
func (s *LogService) GetUserLogs(userId int, query *LogQuery) ([]*model.Log, error) {
	return model.GetUserLogs(
		userId,
		query.LogType,
		query.StartTimestamp,
		query.EndTimestamp,
		query.ModelName,
		query.TokenName,
		query.Page*query.PageSize,
		query.PageSize,
	)
}

// SearchAllLogs 搜索所有日志
func (s *LogService) SearchAllLogs(keyword string) ([]*model.Log, error) {
	return model.SearchAllLogs(keyword)
}

// SearchUserLogs 搜索用户日志
func (s *LogService) SearchUserLogs(userId int, keyword string) ([]*model.Log, error) {
	return model.SearchUserLogs(userId, keyword)
}

// GetLogsStat 获取日志统计
func (s *LogService) GetLogsStat(query *LogQuery) (*LogStat, error) {
	quotaNum := model.SumUsedQuota(
		query.LogType,
		query.StartTimestamp,
		query.EndTimestamp,
		query.ModelName,
		query.Username,
		query.TokenName,
		query.Channel,
	)
	return &LogStat{Quota: quotaNum}, nil
}

// DeleteHistoryLogs 删除历史日志
func (s *LogService) DeleteHistoryLogs(targetTimestamp int64) (int64, error) {
	if targetTimestamp == 0 {
		return 0, ErrTargetTimestampRequired
	}
	return model.DeleteOldLog(targetTimestamp)
}
