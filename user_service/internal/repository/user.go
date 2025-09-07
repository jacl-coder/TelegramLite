package repository

import (
	"errors"
	"fmt"
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
}

// SearchUsersByKeywordWithPagination 根据关键字搜索用户（带分页）
func (r *UserRepository) SearchUsersByKeywordWithPagination(keyword string, limit, offset int, currentUserID uint) ([]*model.UserProfile, int64, error) {
	var profiles []*model.UserProfile
	var total int64

	// 构建基础查询条件
	baseQuery := r.db.Joins("JOIN users ON users.id = user_profiles.user_id").
		Where("users.is_active = ? AND (user_profiles.nickname ILIKE ?)",
			true, "%"+keyword+"%")

	// 如果当前用户已登录，排除屏蔽关系
	if currentUserID != 0 {
		// 排除被当前用户屏蔽的用户（非软删除的屏蔽记录）
		baseQuery = baseQuery.Where("user_profiles.user_id NOT IN (?)",
			r.db.Select("blocked_id").Table("blocked_users").Where("user_id = ? AND deleted_at IS NULL", currentUserID))

		// 排除屏蔽了当前用户的用户（非软删除的屏蔽记录）
		baseQuery = baseQuery.Where("user_profiles.user_id NOT IN (?)",
			r.db.Select("user_id").Table("blocked_users").Where("blocked_id = ? AND deleted_at IS NULL", currentUserID))
	}

	// 获取总数
	err := baseQuery.Model(&model.UserProfile{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = baseQuery.Offset(offset).Limit(limit).Find(&profiles).Error
	if err != nil {
		return nil, 0, err
	}

	return profiles, total, nil
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

// BlockUser 屏蔽用户
func (r *UserRepository) BlockUser(userID, blockedID uint, reason string) error {
	// 首先检查是否存在活跃的屏蔽记录
	var activeCount int64
	err := r.db.Model(&model.BlockedUser{}).
		Where("user_id = ? AND blocked_id = ? AND deleted_at IS NULL", userID, blockedID).
		Count(&activeCount).Error
	if err != nil {
		return err
	}

	if activeCount > 0 {
		return fmt.Errorf("user already blocked")
	}

	// 检查是否存在软删除的屏蔽记录
	var existingRecord model.BlockedUser
	err = r.db.Unscoped(). // Unscoped 查询包含软删除的记录
				Where("user_id = ? AND blocked_id = ?", userID, blockedID).
				First(&existingRecord).Error

	if err == nil {
		// 存在软删除的记录，恢复它
		existingRecord.Reason = reason
		existingRecord.DeletedAt = gorm.DeletedAt{} // 清空 deleted_at
		return r.db.Unscoped().Save(&existingRecord).Error
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 数据库查询错误
		return err
	}

	// 不存在任何记录，创建新的屏蔽记录
	blockedUser := &model.BlockedUser{
		UserID:    userID,
		BlockedID: blockedID,
		Reason:    reason,
	}

	return r.db.Create(blockedUser).Error
}

// UnblockUser 取消屏蔽用户
func (r *UserRepository) UnblockUser(userID, blockedID uint) error {
	result := r.db.Where("user_id = ? AND blocked_id = ?", userID, blockedID).
		Delete(&model.BlockedUser{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not blocked or already unblocked")
	}

	return nil
}

// GetBlockedUsers 获取用户屏蔽列表
func (r *UserRepository) GetBlockedUsers(userID uint, limit, offset int) ([]*model.UserProfile, int64, error) {
	var blockedProfiles []*model.UserProfile
	var total int64

	// 构建查询：获取被屏蔽用户的档案信息，排除软删除的屏蔽记录
	baseQuery := r.db.Table("user_profiles").
		Joins("JOIN blocked_users ON blocked_users.blocked_id = user_profiles.user_id").
		Where("blocked_users.user_id = ? AND blocked_users.deleted_at IS NULL", userID)

	// 获取总数
	err := baseQuery.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err = baseQuery.Offset(offset).Limit(limit).Find(&blockedProfiles).Error
	if err != nil {
		return nil, 0, err
	}

	return blockedProfiles, total, nil
}

// IsUserBlocked 检查用户是否被屏蔽
func (r *UserRepository) IsUserBlocked(userID, targetUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.BlockedUser{}).
		Where("user_id = ? AND blocked_id = ? AND deleted_at IS NULL", userID, targetUserID).
		Count(&count).Error

	return count > 0, err
}

// IsBlockedBy 检查是否被某用户屏蔽
func (r *UserRepository) IsBlockedBy(userID, byUserID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.BlockedUser{}).
		Where("user_id = ? AND blocked_id = ? AND deleted_at IS NULL", byUserID, userID).
		Count(&count).Error

	return count > 0, err
}
