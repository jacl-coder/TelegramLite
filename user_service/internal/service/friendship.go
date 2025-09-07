package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jacl-coder/telegramlite/user_service/internal/model"
	"github.com/jacl-coder/telegramlite/user_service/internal/repository"
)

// FriendshipService 好友关系服务
type FriendshipService struct {
	friendshipRepo *repository.FriendshipRepository
	userRepo       *repository.UserRepository
	cacheRepo      *repository.UserCacheRepository
}

// NewFriendshipService 创建好友关系服务
func NewFriendshipService() *FriendshipService {
	var cacheRepo *repository.UserCacheRepository
	if redisClient := repository.GetRedis(); redisClient != nil {
		cacheRepo = repository.NewUserCacheRepository(redisClient)
	}

	return &FriendshipService{
		friendshipRepo: repository.NewFriendshipRepository(),
		userRepo:       repository.NewUserRepository(),
		cacheRepo:      cacheRepo,
	}
}

// SendFriendRequest 发送好友请求
func (s *FriendshipService) SendFriendRequest(fromID, toID uint, message string) error {
	// 检查是否已经是好友
	isFriend, err := s.friendshipRepo.CheckFriendship(fromID, toID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}
	if isFriend {
		return errors.New("already friends")
	}

	// 检查是否已有待处理请求
	existingReq, err := s.friendshipRepo.GetFriendRequest(fromID, toID)
	if err != nil {
		return fmt.Errorf("failed to check existing request: %w", err)
	}
	if existingReq != nil {
		return errors.New("friend request already sent")
	}

	// 检查目标用户是否允许好友请求
	toUserSettings, err := s.userRepo.GetUserSettings(toID)
	if err != nil {
		return fmt.Errorf("failed to get user settings: %w", err)
	}
	if !toUserSettings.AllowFriendRequests {
		return errors.New("user does not allow friend requests")
	}

	// 创建好友请求
	request := &model.FriendRequest{
		FromID:  fromID,
		ToID:    toID,
		Message: message,
		Status:  "pending",
	}

	err = s.friendshipRepo.SendFriendRequest(request)
	if err != nil {
		return fmt.Errorf("failed to send friend request: %w", err)
	}

	return nil
}

// GetPendingFriendRequests 获取待处理的好友请求
func (s *FriendshipService) GetPendingFriendRequests(userID uint) ([]*model.FriendRequest, error) {
	requests, err := s.friendshipRepo.GetPendingFriendRequests(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	return requests, nil
}

// AcceptFriendRequest 接受好友请求
func (s *FriendshipService) AcceptFriendRequest(requestID, userID uint) error {
	// 获取好友请求
	request, err := s.getAndValidateRequest(requestID, userID)
	if err != nil {
		return err
	}

	// 更新请求状态
	err = s.friendshipRepo.UpdateFriendRequestStatus(requestID, "accepted")
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	// 创建好友关系
	err = s.friendshipRepo.CreateFriendship(request.FromID, request.ToID)
	if err != nil {
		return fmt.Errorf("failed to create friendship: %w", err)
	}

	// 清除双方的好友列表缓存
	if s.cacheRepo != nil {
		ctx := context.Background()
		go func() {
			s.cacheRepo.InvalidateFriendshipCache(ctx, request.FromID, request.ToID)
		}()
	}

	return nil
}

// RejectFriendRequest 拒绝好友请求
func (s *FriendshipService) RejectFriendRequest(requestID, userID uint) error {
	// 验证请求
	_, err := s.getAndValidateRequest(requestID, userID)
	if err != nil {
		return err
	}

	// 更新请求状态
	err = s.friendshipRepo.UpdateFriendRequestStatus(requestID, "rejected")
	if err != nil {
		return fmt.Errorf("failed to update request status: %w", err)
	}

	return nil
}

// GetFriendsList 获取好友列表
func (s *FriendshipService) GetFriendsList(userID uint, page, pageSize int) ([]*model.Friendship, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	ctx := context.Background()

	// 只缓存第一页的好友列表，因为第一页是最常访问的
	if page == 1 && s.cacheRepo != nil {
		if cachedFriends, err := s.cacheRepo.GetFriendsList(ctx, userID); err == nil && cachedFriends != nil {
			// 限制返回数量
			if len(cachedFriends) > pageSize {
				cachedFriends = cachedFriends[:pageSize]
			}
			return cachedFriends, nil
		}
	}

	// 从数据库获取
	friendships, err := s.friendshipRepo.GetFriendsList(userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends list: %w", err)
	}

	// 缓存第一页的结果
	if page == 1 && s.cacheRepo != nil && len(friendships) > 0 {
		go func() {
			if err := s.cacheRepo.SetFriendsList(ctx, userID, friendships); err != nil {
				fmt.Printf("Failed to cache friends list: %v\n", err)
			}
		}()
	}

	return friendships, nil
}

// DeleteFriend 删除好友
func (s *FriendshipService) DeleteFriend(userID, friendID uint) error {
	// 检查是否为好友关系
	friendship, err := s.friendshipRepo.GetFriendship(userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to check friendship: %w", err)
	}
	if friendship == nil {
		return errors.New("not friends")
	}

	// 删除好友关系
	err = s.friendshipRepo.DeleteFriendship(userID, friendID)
	if err != nil {
		return fmt.Errorf("failed to delete friendship: %w", err)
	}

	// 清除双方的好友列表缓存
	if s.cacheRepo != nil {
		ctx := context.Background()
		go func() {
			s.cacheRepo.InvalidateFriendshipCache(ctx, userID, friendID)
		}()
	}

	return nil
}

// GetMutualFriends 获取共同好友
func (s *FriendshipService) GetMutualFriends(userID1, userID2 uint) ([]*model.UserProfile, error) {
	mutualFriends, err := s.friendshipRepo.GetMutualFriends(userID1, userID2)
	if err != nil {
		return nil, fmt.Errorf("failed to get mutual friends: %w", err)
	}
	return mutualFriends, nil
}

// 私有辅助方法
func (s *FriendshipService) getAndValidateRequest(requestID, userID uint) (*model.FriendRequest, error) {
	// 获取好友请求
	request, err := s.friendshipRepo.GetFriendRequestByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friend request: %w", err)
	}
	if request == nil {
		return nil, errors.New("friend request not found")
	}

	// 验证请求接收者
	if request.ToID != userID {
		return nil, errors.New("not authorized to handle this request")
	}

	// 验证请求状态
	if request.Status != "pending" {
		return nil, errors.New("request is not pending")
	}

	return request, nil
}
