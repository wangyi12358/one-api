package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenService_validateToken(t *testing.T) {
	service := NewTokenService()

	tests := []struct {
		name      string
		tokenName string
		subnet    *string
		wantErr   error
	}{
		{
			name:      "valid name without subnet",
			tokenName: "test-token",
			subnet:    nil,
			wantErr:   nil,
		},
		{
			name:      "valid name with empty subnet",
			tokenName: "test-token",
			subnet:    stringPtr(""),
			wantErr:   nil,
		},
		{
			name:      "valid name with valid subnet",
			tokenName: "test-token",
			subnet:    stringPtr("192.168.0.0/24"),
			wantErr:   nil,
		},
		{
			name:      "valid name with multiple subnets",
			tokenName: "test-token",
			subnet:    stringPtr("192.168.0.0/24,10.0.0.0/8"),
			wantErr:   nil,
		},
		{
			name:      "name too long",
			tokenName: "this-is-a-very-long-token-name-that-exceeds-thirty-characters",
			subnet:    nil,
			wantErr:   ErrTokenNameTooLong,
		},
		{
			name:      "invalid subnet format",
			tokenName: "test-token",
			subnet:    stringPtr("invalid-subnet"),
			wantErr:   ErrInvalidSubnet,
		},
		{
			name:      "invalid subnet CIDR",
			tokenName: "test-token",
			subnet:    stringPtr("192.168.0.0/33"),
			wantErr:   ErrInvalidSubnet,
		},
		{
			name:      "exactly 30 chars name",
			tokenName: "123456789012345678901234567890",
			subnet:    nil,
			wantErr:   nil,
		},
		{
			name:      "31 chars name",
			tokenName: "1234567890123456789012345678901",
			subnet:    nil,
			wantErr:   ErrTokenNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateToken(tt.tokenName, tt.subnet)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
