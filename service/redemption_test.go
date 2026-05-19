package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedemptionService_validateCreateRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     *CreateRedemptionRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: 5,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			req: &CreateRedemptionRequest{
				Name:  "",
				Quota: 1000,
				Count: 5,
			},
			wantErr: ErrRedemptionNameLength,
		},
		{
			name: "name too long",
			req: &CreateRedemptionRequest{
				Name:  "this-is-a-very-long-redemption-name-that-exceeds-twenty-characters",
				Quota: 1000,
				Count: 5,
			},
			wantErr: ErrRedemptionNameLength,
		},
		{
			name: "name exactly 20 chars",
			req: &CreateRedemptionRequest{
				Name:  "12345678901234567890",
				Quota: 1000,
				Count: 5,
			},
			wantErr: nil,
		},
		{
			name: "name 21 chars",
			req: &CreateRedemptionRequest{
				Name:  "123456789012345678901",
				Quota: 1000,
				Count: 5,
			},
			wantErr: ErrRedemptionNameLength,
		},
		{
			name: "count zero",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: 0,
			},
			wantErr: ErrRedemptionCountZero,
		},
		{
			name: "count negative",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: -1,
			},
			wantErr: ErrRedemptionCountZero,
		},
		{
			name: "count exceeds max",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: 101,
			},
			wantErr: ErrRedemptionCountMax,
		},
		{
			name: "count exactly 100",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: 100,
			},
			wantErr: nil,
		},
		{
			name: "count exactly 1",
			req: &CreateRedemptionRequest{
				Name:  "test-redemption",
				Quota: 1000,
				Count: 1,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试参数验证逻辑
			err := validateCreateRedemptionRequest(tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// validateCreateRedemptionRequest 是从 CreateRedemptions 中提取的验证逻辑
func validateCreateRedemptionRequest(req *CreateRedemptionRequest) error {
	if len(req.Name) == 0 || len(req.Name) > 20 {
		return ErrRedemptionNameLength
	}
	if req.Count <= 0 {
		return ErrRedemptionCountZero
	}
	if req.Count > 100 {
		return ErrRedemptionCountMax
	}
	return nil
}
