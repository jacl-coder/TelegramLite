package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户repository
func NewUserRepository() *UserRepository {
	return &UserRepository{
		db: GetDB(),
	}
}

// GetUserByID 根据ID获取用户基本信息
func (r *UserRepository) GetUserByID(userID uint) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND is_active = ?", userID, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserProfileByID 根据用户ID获取用户资料
func (r *UserRepository) GetUserProfileByID(userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	err := r.db.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// CreateUserProfile 创建用户资料
func (r *UserRepository) CreateUserProfile(profile *model.UserProfile) error {
	return r.db.Create(profile).Error
}

// UpdateUserProfile 更新用户资料
func (r *UserRepository) UpdateUserProfile(profile *model.UserProfile) error {
	return r.db.Save(profile).Error
}

// SearchUsersByKeyword 根据关键字搜索用户
func (r *UserRepository) SearchUsersByKeyword(keyword string, limit int) ([]*model.UserProfile, error) {
	var profiles []*model.UserProfile
	err := r.db.Joins("JOIN users ON users.id = user_profiles.user_id").
		Where("users.is_active = ? AND (user_profiles.nickname ILIKE ?)",
			true, "%"+keyword+"%").
		Limit(limit).
		Find(&profiles).Error

	if err != nil {
		return nil, err
	}
	return profiles, nil
} // GetUsersByIDs 批量获取用户信息
func (r *UserRepository) GetUsersByIDs(userIDs []uint) ([]*model.User, error) {
	var users []*model.User
	err := r.db.Where("id IN ? AND is_active = ?", userIDs, true).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// UpdateUserStatus 更新用户在线状态
func (r *UserRepository) UpdateUserStatus(userID uint, status string, lastSeenAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if lastSeenAt != nil {
		updates["last_seen_at"] = *lastSeenAt
	}

	return r.db.Model(&model.UserProfile{}).
		Where("user_id = ?", userID).
		Updates(updates).Error
}

// GetUserSettings 获取用户设置
func (r *UserRepository) GetUserSettings(userID uint) (*model.UserSetting, error) {
	var settings model.UserSetting
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有设置记录，创建默认设置
			defaultSettings := &model.UserSetting{
				UserID:               userID,
				AllowFriendRequests:  true,
				AllowBeingSearched:   true,
				MessageNotifications: true,
				FriendNotifications:  true,
			}
			if createErr := r.db.Create(defaultSettings).Error; createErr != nil {
				return nil, createErr
			}
			return defaultSettings, nil
		}
		return nil, err
	}
	return &settings, nil
}

// UpdateUserSettings 更新用户设置
func (r *UserRepository) UpdateUserSettings(settings *model.UserSetting) error {
	return r.db.Save(settings).Error
}
