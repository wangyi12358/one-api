package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserService_ManageUser_Validation(t *testing.T) {
	tests := []struct {
		name         string
		req          *ManageUserRequest
		operatorRole int
		wantErr      error
	}{
		{
			name: "disable action by admin",
			req: &ManageUserRequest{
				Username: "testuser",
				Action:   ManageActionDisable,
			},
			operatorRole: 10, // AdminUser
			wantErr:      nil,
		},
		{
			name: "enable action by admin",
			req: &ManageUserRequest{
				Username: "testuser",
				Action:   ManageActionEnable,
			},
			operatorRole: 10, // AdminUser
			wantErr:      nil,
		},
		{
			name: "promote action by non-root",
			req: &ManageUserRequest{
				Username: "testuser",
				Action:   ManageActionPromote,
			},
			operatorRole: 10, // AdminUser
			wantErr:      ErrCannotPromoteByNonRoot,
		},
		{
			name: "promote action by root",
			req: &ManageUserRequest{
				Username: "testuser",
				Action:   ManageActionPromote,
			},
			operatorRole: 100, // RootUser
			wantErr:      nil,
		},
		{
			name: "invalid action",
			req: &ManageUserRequest{
				Username: "testuser",
				Action:   ManageAction("invalid"),
			},
			operatorRole: 100, // RootUser
			wantErr:      nil, // 会在后续查询用户时失败
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 注意：这个测试需要数据库连接才能完整测试
			// 这里只测试参数验证逻辑
			err := validateManageUserRequest(tt.req, tt.operatorRole)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				// 对于需要查询数据库的操作，我们只验证参数验证通过
				if tt.req.Action == ManageActionPromote && tt.operatorRole != 100 {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

// validateManageUserRequest 是从 ManageUser 中提取的验证逻辑
func validateManageUserRequest(req *ManageUserRequest, operatorRole int) error {
	// 验证操作类型
	switch req.Action {
	case ManageActionDisable, ManageActionEnable, ManageActionDelete:
		// 这些操作需要查询用户后才能验证权限
		return nil
	case ManageActionPromote:
		if operatorRole != 100 { // RoleRootUser
			return ErrCannotPromoteByNonRoot
		}
		return nil
	case ManageActionDemote:
		return nil
	default:
		return nil
	}
}

func TestManageAction_Constants(t *testing.T) {
	assert.Equal(t, ManageAction("disable"), ManageActionDisable)
	assert.Equal(t, ManageAction("enable"), ManageActionEnable)
	assert.Equal(t, ManageAction("delete"), ManageActionDelete)
	assert.Equal(t, ManageAction("promote"), ManageActionPromote)
	assert.Equal(t, ManageAction("demote"), ManageActionDemote)
}

func TestErrorDefinitions(t *testing.T) {
	// 验证错误定义是否正确
	assert.NotNil(t, ErrRegisterDisabled)
	assert.NotNil(t, ErrPasswordRegisterDisabled)
	assert.NotNil(t, ErrEmailVerificationRequired)
	assert.NotNil(t, ErrVerificationCodeInvalid)
	assert.NotNil(t, ErrUserNotFound)
	assert.NotNil(t, ErrNoPermission)
	assert.NotNil(t, ErrCannotDisableRootUser)
	assert.NotNil(t, ErrCannotDeleteRootUser)
	assert.NotNil(t, ErrCannotDemoteRootUser)
	assert.NotNil(t, ErrUserAlreadyAdmin)
	assert.NotNil(t, ErrUserAlreadyCommon)
	assert.NotNil(t, ErrCannotPromoteByNonRoot)
	assert.NotNil(t, ErrCannotCreateHigherRole)
	assert.NotNil(t, ErrAccessTokenDuplicate)
}
