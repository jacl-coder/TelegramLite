package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
	"github.com/jacl-coder/telegramlite/user_service/internal/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo       *repository.UserRepository
	friendshipRepo *repository.FriendshipRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{
		userRepo:       repository.NewUserRepository(),
		friendshipRepo: repository.NewFriendshipRepository(),
	}
}

// GetUserProfile 获取用户资料
func (s *UserService) GetUserProfile(userID uint) (*model.UserProfile, error) {
	profile, err := s.userRepo.GetUserProfileByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile == nil {
		return nil, errors.New("user profile not found")
	}
	return profile, nil
}

// UpdateUserProfile 更新用户资料
func (s *UserService) UpdateUserProfile(userID uint, req *UpdateProfileRequest) (*model.UserProfile, error) {
	// 获取现有资料
	profile, err := s.userRepo.GetUserProfileByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// 如果没有资料，创建新的
	if profile == nil {
		profile = &model.UserProfile{
			UserID: userID,
		}
	}

	// 更新字段
	if req.Nickname != nil {
		profile.Nickname = *req.Nickname
	}
	if req.FirstName != nil {
		profile.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		profile.LastName = *req.LastName
	}
	if req.Bio != nil {
		profile.Bio = *req.Bio
	}
	if req.Avatar != nil {
		profile.Avatar = *req.Avatar
	}
	if req.Status != nil {
		profile.Status = *req.Status
	}
	if req.Birthday != nil {
		profile.Birthday = req.Birthday
	}
	if req.Gender != nil {
		profile.Gender = *req.Gender
	}
	if req.Language != nil {
		profile.Language = *req.Language
	}
	if req.Timezone != nil {
		profile.Timezone = *req.Timezone
	}

	// 保存资料
	if profile.ID == 0 {
		err = s.userRepo.CreateUserProfile(profile)
	} else {
		err = s.userRepo.UpdateUserProfile(profile)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to save user profile: %w", err)
	}

	return profile, nil
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string, limit int) ([]*model.UserProfile, error) {
	if keyword == "" {
		return []*model.UserProfile{}, nil
	}

	profiles, err := s.userRepo.SearchUsersByKeyword(keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return profiles, nil
}

// UpdateUserStatus 更新用户在线状态
func (s *UserService) UpdateUserStatus(userID uint, status string) error {
	now := time.Now()
	err := s.userRepo.UpdateUserStatus(userID, status, &now)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}
	return nil
}

// GetUserSettings 获取用户设置
func (s *UserService) GetUserSettings(userID uint) (*model.UserSetting, error) {
	settings, err := s.userRepo.GetUserSettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return settings, nil
}

// UpdateUserSettings 更新用户设置
func (s *UserService) UpdateUserSettings(userID uint, req *UpdateSettingsRequest) (*model.UserSetting, error) {
	// 获取现有设置
	settings, err := s.userRepo.GetUserSettings(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}

	// 更新字段
	if req.AllowFriendRequests != nil {
		settings.AllowFriendRequests = *req.AllowFriendRequests
	}
	if req.AllowBeingSearched != nil {
		settings.AllowBeingSearched = *req.AllowBeingSearched
	}
	if req.ShowOnlineStatus != nil {
		settings.ShowOnlineStatus = *req.ShowOnlineStatus
	}
	if req.ShowLastSeen != nil {
		settings.ShowLastSeen = *req.ShowLastSeen
	}
	if req.MessageNotifications != nil {
		settings.MessageNotifications = *req.MessageNotifications
	}
	if req.FriendNotifications != nil {
		settings.FriendNotifications = *req.FriendNotifications
	}

	// 保存设置
	err = s.userRepo.UpdateUserSettings(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to update user settings: %w", err)
	}

	return settings, nil
}

// DTO 结构体

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Nickname  *string    `json:"nickname"`
	FirstName *string    `json:"first_name"`
	LastName  *string    `json:"last_name"`
	Bio       *string    `json:"bio"`
	Avatar    *string    `json:"avatar"`
	Status    *string    `json:"status"`
	Birthday  *time.Time `json:"birthday"`
	Gender    *string    `json:"gender"`
	Language  *string    `json:"language"`
	Timezone  *string    `json:"timezone"`
}

// UpdateSettingsRequest 更新设置请求
type UpdateSettingsRequest struct {
	AllowFriendRequests  *bool `json:"allow_friend_requests"`
	AllowBeingSearched   *bool `json:"allow_being_searched"`
	ShowOnlineStatus     *bool `json:"show_online_status"`
	ShowLastSeen         *bool `json:"show_last_seen"`
	MessageNotifications *bool `json:"message_notifications"`
	FriendNotifications  *bool `json:"friend_notifications"`
}
