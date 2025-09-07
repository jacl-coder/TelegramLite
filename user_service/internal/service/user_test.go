package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
	"github.com/jacl-coder/telegramlite/user_service/internal/repository"
)

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	return &s
}

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	// 使用内存SQLite数据库进行测试
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移测试表
	err = db.AutoMigrate(
		&model.User{},
		&model.UserProfile{},
		&model.FriendRequest{},
		&model.Friendship{},
		&model.UserSetting{},
		&model.BlockedUser{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestUserService_GetUserProfile(t *testing.T) {
	// 设置测试数据库
	testDB := setupTestDB(t)

	// 临时替换全局DB实例用于测试
	originalDB := repository.DB
	repository.DB = testDB
	defer func() {
		repository.DB = originalDB
	}()

	// 创建测试数据
	testUser := &model.User{
		ID:       1,
		Username: "testuser",
		Phone:    "13812345678",
		Email:    "test@example.com",
		IsActive: true,
	}
	testDB.Create(testUser)

	testProfile := &model.UserProfile{
		UserID:    1,
		Nickname:  "Test User",
		FirstName: "Test",
		LastName:  "User",
		Bio:       "Test bio",
	}
	testDB.Create(testProfile)

	tests := []struct {
		name    string
		userID  uint
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid user ID - zero",
			userID:  0,
			wantErr: true,
			errMsg:  "user ID cannot be zero",
		},
		{
			name:    "valid user ID",
			userID:  1,
			wantErr: false,
		},
		{
			name:    "user not found",
			userID:  999,
			wantErr: true,
			errMsg:  "user profile not found",
		},
	}

	service := NewUserService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetUserProfile(tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.UserID)
			}
		})
	}
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	// 设置测试数据库
	testDB := setupTestDB(t)

	// 临时替换全局DB实例用于测试
	originalDB := repository.DB
	repository.DB = testDB
	defer func() {
		repository.DB = originalDB
	}()

	// 创建测试数据
	testUser := &model.User{
		ID:       1,
		Username: "testuser",
		Phone:    "13812345678",
		Email:    "test@example.com",
		IsActive: true,
	}
	testDB.Create(testUser)

	testProfile := &model.UserProfile{
		UserID:    1,
		Nickname:  "Test User",
		FirstName: "Test",
		LastName:  "User",
		Bio:       "Test bio",
	}
	testDB.Create(testProfile)

	tests := []struct {
		name    string
		userID  uint
		req     *UpdateProfileRequest
		wantErr bool
		errMsg  string
	}{
		{
			name:    "invalid user ID - zero",
			userID:  0,
			req:     &UpdateProfileRequest{},
			wantErr: true,
			errMsg:  "user ID cannot be zero",
		},
		{
			name:    "nil request",
			userID:  1,
			req:     nil,
			wantErr: true,
			errMsg:  "update request cannot be nil",
		},
		{
			name:   "valid update",
			userID: 1,
			req: &UpdateProfileRequest{
				Nickname:  stringPtr("Updated Nickname"),
				FirstName: stringPtr("Updated"),
				LastName:  stringPtr("User"),
				Bio:       stringPtr("Updated bio"),
			},
			wantErr: false,
		},
	}

	service := NewUserService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.UpdateUserProfile(tt.userID, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, result)
			}
		})
	}
}

func TestUserService_BlockUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    uint
		blockedID uint
		reason    string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "invalid user ID - zero",
			userID:    0,
			blockedID: 1,
			reason:    "test",
			wantErr:   true,
			errMsg:    "invalid user ID",
		},
		{
			name:      "invalid blocked ID - zero",
			userID:    1,
			blockedID: 0,
			reason:    "test",
			wantErr:   true,
			errMsg:    "invalid user ID",
		},
		{
			name:      "cannot block yourself",
			userID:    1,
			blockedID: 1,
			reason:    "test",
			wantErr:   true,
			errMsg:    "cannot block yourself",
		},
	}

	service := NewUserService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.BlockUser(tt.userID, tt.blockedID, tt.reason)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			}
		})
	}
}
