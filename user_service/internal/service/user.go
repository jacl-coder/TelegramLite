package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
	"github.com/jacl-coder/telegramlite/user_service/internal/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo       *repository.UserRepository
	friendshipRepo *repository.FriendshipRepository
	cacheRepo      *repository.UserCacheRepository
}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	var cacheRepo *repository.UserCacheRepository
	if redisClient := repository.GetRedis(); redisClient != nil {
		cacheRepo = repository.NewUserCacheRepository(redisClient)
	}

	return &UserService{
		userRepo:       repository.NewUserRepository(),
		friendshipRepo: repository.NewFriendshipRepository(),
		cacheRepo:      cacheRepo,
	}
}

// GetUserProfile 获取用户资料
func (s *UserService) GetUserProfile(userID uint) (*model.UserProfile, error) {
	ctx := context.Background()

	// 先尝试从缓存获取
	if s.cacheRepo != nil {
		if profile, err := s.cacheRepo.GetUserProfile(ctx, userID); err == nil && profile != nil {
			return profile, nil
		}
	}

	// 缓存未命中，从数据库获取
	profile, err := s.userRepo.GetUserProfileByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	if profile == nil {
		return nil, errors.New("user profile not found")
	}

	// 缓存到Redis
	if s.cacheRepo != nil {
		go func() {
			if err := s.cacheRepo.SetUserProfile(ctx, userID, profile); err != nil {
				// 记录日志但不影响主流程
				fmt.Printf("Failed to cache user profile: %v\n", err)
			}
		}()
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

	// 更新缓存
	if s.cacheRepo != nil {
		ctx := context.Background()
		go func() {
			// 删除旧缓存，下次访问时会重新缓存
			if err := s.cacheRepo.DeleteUserProfile(ctx, userID); err != nil {
				fmt.Printf("Failed to invalidate user profile cache: %v\n", err)
			}
		}()
	}

	return profile, nil
}

// SearchUsers 搜索用户
func (s *UserService) SearchUsers(keyword string, limit int) ([]*model.UserProfile, error) {
	if keyword == "" {
		return []*model.UserProfile{}, nil
	}

	ctx := context.Background()

	// 先尝试从缓存获取
	if s.cacheRepo != nil {
		if profiles, err := s.cacheRepo.GetSearchResult(ctx, keyword); err == nil && profiles != nil {
			// 限制返回数量
			if len(profiles) > limit {
				profiles = profiles[:limit]
			}
			return profiles, nil
		}
	}

	// 缓存未命中，从数据库搜索
	profiles, err := s.userRepo.SearchUsersByKeyword(keyword, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// 缓存搜索结果
	if s.cacheRepo != nil && len(profiles) > 0 {
		go func() {
			if err := s.cacheRepo.SetSearchResult(ctx, keyword, profiles); err != nil {
				fmt.Printf("Failed to cache search result: %v\n", err)
			}
		}()
	}

	return profiles, nil
}

// SearchUsersWithPagination 搜索用户（带分页）
func (s *UserService) SearchUsersWithPagination(keyword string, limit, offset int, currentUserID uint) ([]*model.UserProfile, int64, error) {
	if keyword == "" {
		return []*model.UserProfile{}, 0, nil
	}

	profiles, total, err := s.userRepo.SearchUsersByKeywordWithPagination(keyword, limit, offset, currentUserID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search users with pagination: %w", err)
	}

	return profiles, total, nil
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

// BlockUser 屏蔽用户
func (s *UserService) BlockUser(userID, blockedID uint, reason string) error {
	// 验证参数
	if userID == 0 || blockedID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	if userID == blockedID {
		return fmt.Errorf("cannot block yourself")
	}

	// 检查被屏蔽用户是否存在
	_, err := s.userRepo.GetUserByID(blockedID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("target user not found")
		}
		return fmt.Errorf("failed to check target user: %w", err)
	}

	// 执行屏蔽
	err = s.userRepo.BlockUser(userID, blockedID, reason)
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	// 清除相关缓存
	if s.cacheRepo != nil {
		ctx := context.Background()
		go func() {
			s.cacheRepo.DeleteBlockedUsers(ctx, userID)
		}()
	}

	return nil
}

// UnblockUser 取消屏蔽用户
func (s *UserService) UnblockUser(userID, blockedID uint) error {
	// 验证参数
	if userID == 0 || blockedID == 0 {
		return fmt.Errorf("invalid user ID")
	}

	// 执行取消屏蔽
	err := s.userRepo.UnblockUser(userID, blockedID)
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	// 清除相关缓存
	if s.cacheRepo != nil {
		ctx := context.Background()
		go func() {
			s.cacheRepo.DeleteBlockedUsers(ctx, userID)
		}()
	}

	return nil
}

// GetBlockedUsers 获取屏蔽用户列表
func (s *UserService) GetBlockedUsers(userID uint, limit, offset int) ([]*model.UserProfile, int64, error) {
	// 验证参数
	if userID == 0 {
		return nil, 0, fmt.Errorf("invalid user ID")
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	// 获取屏蔽列表
	profiles, total, err := s.userRepo.GetBlockedUsers(userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get blocked users: %w", err)
	}

	return profiles, total, nil
}

// IsUserBlocked 检查用户是否被屏蔽
func (s *UserService) IsUserBlocked(userID, targetUserID uint) (bool, error) {
	if userID == 0 || targetUserID == 0 {
		return false, fmt.Errorf("invalid user ID")
	}

	return s.userRepo.IsUserBlocked(userID, targetUserID)
}

// IsBlockedBy 检查是否被某用户屏蔽
func (s *UserService) IsBlockedBy(userID, byUserID uint) (bool, error) {
	if userID == 0 || byUserID == 0 {
		return false, fmt.Errorf("invalid user ID")
	}

	return s.userRepo.IsBlockedBy(userID, byUserID)
}
