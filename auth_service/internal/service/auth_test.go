package service

import (
	"testing"

	"github.com/jacl-coder/telegramlite/auth_service/pkg"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register(t *testing.T) {
	// 创建JWT管理器用于测试
	jwtManager := pkg.NewJWTManager("test-secret", 3600, 7*24*3600)
	authService := NewAuthService(jwtManager)

	tests := []struct {
		name    string
		req     *RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing phone and email",
			req: &RegisterRequest{
				Phone:       "",
				Email:       "",
				Username:    "testuser",
				Password:    "password123",
				DeviceToken: "device123",
				DeviceType:  "ios",
			},
			wantErr: true,
			errMsg:  "手机号或邮箱必须提供一个",
		},
		{
			name: "weak password",
			req: &RegisterRequest{
				Phone:       "+1234567890",
				Username:    "testuser",
				Password:    "123", // 太短
				DeviceToken: "device123",
				DeviceType:  "ios",
			},
			wantErr: true,
			errMsg:  "密码长度至少6位",
		},
		{
			name: "invalid device type",
			req: &RegisterRequest{
				Phone:       "+1234567890",
				Username:    "testuser",
				Password:    "password123",
				DeviceToken: "device123",
				DeviceType:  "invalid_type",
			},
			wantErr: true,
			errMsg:  "无效的设备类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.Register(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				// 注意: 在没有初始化数据库的情况下，这里可能会失败
				// 这只是一个基本的结构测试
				if err != nil {
					t.Logf("Expected no error for valid input, but got: %v", err)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	jwtManager := pkg.NewJWTManager("test-secret", 3600, 7*24*3600)
	authService := NewAuthService(jwtManager)

	tests := []struct {
		name    string
		req     *LoginRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing phone and email",
			req: &LoginRequest{
				Phone:       "",
				Email:       "",
				Password:    "password123",
				DeviceToken: "device123",
				DeviceType:  "ios",
			},
			wantErr: true,
			errMsg:  "手机号或邮箱必须提供一个",
		},
		{
			name: "invalid device type",
			req: &LoginRequest{
				Phone:       "+1234567890",
				Password:    "password123",
				DeviceToken: "device123",
				DeviceType:  "invalid_type",
			},
			wantErr: true,
			errMsg:  "无效的设备类型",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.Login(tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	jwtManager := pkg.NewJWTManager("test-secret", 3600, 7*24*3600)
	authService := NewAuthService(jwtManager)

	tests := []struct {
		name         string
		refreshToken string
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "empty refresh token",
			refreshToken: "",
			wantErr:      true,
			errMsg:       "刷新token不能为空",
		},
		{
			name:         "invalid refresh token",
			refreshToken: "invalid_token",
			wantErr:      true,
			errMsg:       "无效的刷新token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.RefreshToken(tt.refreshToken)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			}
		})
	}
}

func TestAuthService_GetUserInfo(t *testing.T) {
	jwtManager := pkg.NewJWTManager("test-secret", 3600, 7*24*3600)
	authService := NewAuthService(jwtManager)

	tests := []struct {
		name    string
		token   string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
			errMsg:  "token不能为空",
		},
		{
			name:    "invalid token",
			token:   "invalid_token",
			wantErr: true,
			errMsg:  "无效的token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := authService.GetUserInfo(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			}
		})
	}
}
