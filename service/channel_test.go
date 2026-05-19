package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateChannelRequest_KeySplitting(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		expectedKeys []string
	}{
		{
			name:         "single key",
			key:          "sk-1234567890",
			expectedKeys: []string{"sk-1234567890"},
		},
		{
			name:         "multiple keys",
			key:          "sk-1234567890\nsk-0987654321",
			expectedKeys: []string{"sk-1234567890", "sk-0987654321"},
		},
		{
			name:         "keys with empty lines",
			key:          "sk-1234567890\n\nsk-0987654321\n",
			expectedKeys: []string{"sk-1234567890", "sk-0987654321"},
		},
		{
			name:         "single key with newline",
			key:          "sk-1234567890\n",
			expectedKeys: []string{"sk-1234567890"},
		},
		{
			name:         "empty key",
			key:          "",
			expectedKeys: []string{},
		},
		{
			name:         "only newlines",
			key:          "\n\n\n",
			expectedKeys: []string{},
		},
		{
			name:         "three keys",
			key:          "sk-1\nsk-2\nsk-3",
			expectedKeys: []string{"sk-1", "sk-2", "sk-3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 key 拆分逻辑
			keys := splitKeys(tt.key)
			assert.Equal(t, tt.expectedKeys, keys)
		})
	}
}

// splitKeys 是从 CreateChannel 中提取的 key 拆分逻辑
func splitKeys(key string) []string {
	keys := make([]string, 0)
	for _, k := range splitAndTrim(key, "\n") {
		if k != "" {
			keys = append(keys, k)
		}
	}
	return keys
}

// splitAndTrim 分割字符串并去除空白
func splitAndTrim(s, sep string) []string {
	result := make([]string, 0)
	for _, part := range splitString(s, sep) {
		trimmed := trimSpace(part)
		result = append(result, trimmed)
	}
	return result
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	result := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if string(s[i]) == sep {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

// trimSpace 去除首尾空白
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for start < end && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func TestCreateChannelRequest_StructInitialization(t *testing.T) {
	baseURL := "https://api.example.com"
	priority := int64(10)
	weight := uint(5)
	modelMapping := `{"gpt-4": "gpt-4-turbo"}`
	systemPrompt := "You are a helpful assistant"

	req := &CreateChannelRequest{
		Name:         "test-channel",
		Type:         1,
		Key:          "sk-1234567890",
		BaseURL:      &baseURL,
		Models:       "gpt-3.5-turbo,gpt-4",
		Group:        "default",
		Priority:     &priority,
		Weight:       &weight,
		Balance:      100.0,
		ModelMapping: &modelMapping,
		Config:       "{}",
		SystemPrompt: &systemPrompt,
	}

	assert.Equal(t, "test-channel", req.Name)
	assert.Equal(t, 1, req.Type)
	assert.Equal(t, "sk-1234567890", req.Key)
	assert.Equal(t, &baseURL, req.BaseURL)
	assert.Equal(t, "gpt-3.5-turbo,gpt-4", req.Models)
	assert.Equal(t, "default", req.Group)
	assert.Equal(t, &priority, req.Priority)
	assert.Equal(t, &weight, req.Weight)
	assert.Equal(t, 100.0, req.Balance)
	assert.Equal(t, &modelMapping, req.ModelMapping)
	assert.Equal(t, "{}", req.Config)
	assert.Equal(t, &systemPrompt, req.SystemPrompt)
}

func TestChannelService_NewChannelService(t *testing.T) {
	service := NewChannelService()
	assert.NotNil(t, service)
}
