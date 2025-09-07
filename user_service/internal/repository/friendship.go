package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
)

// FriendshipRepository 好友关系数据访问层
type FriendshipRepository struct {
	db *gorm.DB
}

// NewFriendshipRepository 创建好友关系repository
func NewFriendshipRepository() *FriendshipRepository {
	return &FriendshipRepository{
		db: GetDB(),
	}
}

// SendFriendRequest 发送好友请求
func (r *FriendshipRepository) SendFriendRequest(req *model.FriendRequest) error {
	return r.db.Create(req).Error
}

// GetFriendRequest 获取好友请求
func (r *FriendshipRepository) GetFriendRequest(fromID, toID uint) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.Where("from_id = ? AND to_id = ? AND status = 'pending'", fromID, toID).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// GetPendingFriendRequests 获取待处理的好友请求
func (r *FriendshipRepository) GetPendingFriendRequests(userID uint) ([]*model.FriendRequest, error) {
	var requests []*model.FriendRequest
	err := r.db.Where("to_id = ? AND status = 'pending'", userID).Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return requests, nil
}

// UpdateFriendRequestStatus 更新好友请求状态
func (r *FriendshipRepository) UpdateFriendRequestStatus(requestID uint, status string) error {
	return r.db.Model(&model.FriendRequest{}).
		Where("id = ?", requestID).
		Update("status", status).Error
}

// CreateFriendship 创建好友关系
func (r *FriendshipRepository) CreateFriendship(userID, friendID uint) error {
	// 创建双向好友关系
	friendships := []*model.Friendship{
		{
			UserID:   userID,
			FriendID: friendID,
			Status:   model.FriendshipAccepted,
		},
		{
			UserID:   friendID,
			FriendID: userID,
			Status:   model.FriendshipAccepted,
		},
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, friendship := range friendships {
			if err := tx.Create(friendship).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetFriendship 获取好友关系
func (r *FriendshipRepository) GetFriendship(userID, friendID uint) (*model.Friendship, error) {
	var friendship model.Friendship
	err := r.db.Where("user_id = ? AND friend_id = ? AND status = 'accepted'", userID, friendID).First(&friendship).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &friendship, nil
}

// GetFriendsList 获取好友列表
func (r *FriendshipRepository) GetFriendsList(userID uint, page, pageSize int) ([]*model.Friendship, error) {
	var friendships []*model.Friendship
	offset := (page - 1) * pageSize

	err := r.db.Where("user_id = ? AND status = 'accepted'", userID).
		Offset(offset).
		Limit(pageSize).
		Find(&friendships).Error

	if err != nil {
		return nil, err
	}
	return friendships, nil
} // DeleteFriendship 删除好友关系
func (r *FriendshipRepository) DeleteFriendship(userID, friendID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除双向好友关系
		if err := tx.Where("user_id = ? AND friend_id = ?", userID, friendID).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? AND friend_id = ?", friendID, userID).Delete(&model.Friendship{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// CheckFriendship 检查是否为好友关系
func (r *FriendshipRepository) CheckFriendship(userID, friendID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Friendship{}).
		Where("user_id = ? AND friend_id = ? AND status = 'accepted'", userID, friendID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetFriendRequestByID 根据ID获取好友请求
func (r *FriendshipRepository) GetFriendRequestByID(requestID uint) (*model.FriendRequest, error) {
	var req model.FriendRequest
	err := r.db.Where("id = ?", requestID).First(&req).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &req, nil
}

// GetMutualFriends 获取共同好友
func (r *FriendshipRepository) GetMutualFriends(userID1, userID2 uint) ([]*model.UserProfile, error) {
	var mutualFriends []*model.UserProfile

	err := r.db.Table("friendships f1").
		Select("up.*").
		Joins("JOIN friendships f2 ON f1.friend_id = f2.friend_id").
		Joins("JOIN user_profiles up ON f1.friend_id = up.user_id").
		Where("f1.user_id = ? AND f2.user_id = ? AND f1.status = 'accepted' AND f2.status = 'accepted'", userID1, userID2).
		Find(&mutualFriends).Error

	if err != nil {
		return nil, err
	}
	return mutualFriends, nil
}
