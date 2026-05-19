package service

import (
	"context"
	"errors"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/random"
	"github.com/songquanpeng/one-api/model"
)

// 业务错误定义
var (
	ErrRedemptionNameLength = errors.New("兑换码名称长度必须在1-20之间")
	ErrRedemptionCountZero  = errors.New("兑换码个数必须大于0")
	ErrRedemptionCountMax   = errors.New("一次兑换码批量生成的个数不能大于 100")
)

// CreateRedemptionRequest 创建兑换码请求参数
type CreateRedemptionRequest struct {
	Name  string `json:"name"`
	Quota int64  `json:"quota"`
	Count int    `json:"count"`
}

// UpdateRedemptionRequest 更新兑换码请求参数
type UpdateRedemptionRequest struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status int    `json:"status"`
	Quota  int64  `json:"quota"`
}

// RedemptionService 兑换码服务
type RedemptionService struct{}

// NewRedemptionService 创建兑换码服务实例
func NewRedemptionService() *RedemptionService {
	return &RedemptionService{}
}

// CreateRedemptions 批量创建兑换码
func (s *RedemptionService) CreateRedemptions(userId int, req *CreateRedemptionRequest) ([]string, error) {
	// 1. 参数验证
	if len(req.Name) == 0 || len(req.Name) > 20 {
		return nil, ErrRedemptionNameLength
	}
	if req.Count <= 0 {
		return nil, ErrRedemptionCountZero
	}
	if req.Count > 100 {
		return nil, ErrRedemptionCountMax
	}

	// 2. 批量生成兑换码
	keys := make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		key := random.GetUUID()
		redemption := &model.Redemption{
			UserId:      userId,
			Name:        req.Name,
			Key:         key,
			CreatedTime: helper.GetTimestamp(),
			Quota:       req.Quota,
		}
		if err := redemption.Insert(); err != nil {
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

// UpdateRedemption 更新兑换码
func (s *RedemptionService) UpdateRedemption(req *UpdateRedemptionRequest, statusOnly bool) (*model.Redemption, error) {
	// 1. 获取现有兑换码
	redemption, err := model.GetRedemptionById(req.Id)
	if err != nil {
		return nil, err
	}

	// 2. 更新字段
	if statusOnly {
		redemption.Status = req.Status
	} else {
		redemption.Name = req.Name
		redemption.Quota = req.Quota
	}

	// 3. 保存更新
	if err := redemption.Update(); err != nil {
		return nil, err
	}

	return redemption, nil
}

// DeleteRedemption 删除兑换码
func (s *RedemptionService) DeleteRedemption(id int) error {
	return model.DeleteRedemptionById(id)
}

// GetRedemption 获取兑换码
func (s *RedemptionService) GetRedemption(id int) (*model.Redemption, error) {
	return model.GetRedemptionById(id)
}

// GetAllRedemptions 获取所有兑换码
func (s *RedemptionService) GetAllRedemptions(startIdx, num int) ([]*model.Redemption, error) {
	return model.GetAllRedemptions(startIdx, num)
}

// SearchRedemptions 搜索兑换码
func (s *RedemptionService) SearchRedemptions(keyword string) ([]*model.Redemption, error) {
	return model.SearchRedemptions(keyword)
}

// RedeemCode 兑换码兑换
func (s *RedemptionService) RedeemCode(ctx context.Context, userId int, key string) (int64, error) {
	return model.Redeem(ctx, key, userId)
}
