package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	applogger "github.com/jacl-coder/TelegramLite/common/go/logger"
	"github.com/jacl-coder/telegramlite/user_service/internal/config"
	"github.com/jacl-coder/telegramlite/user_service/internal/model"
)

// RedisClient Redis客户端实例
var RedisClient *redis.Client

// RedisKeys Redis键名常量
const (
	// 用户信息缓存键前缀
	UserProfileKey = "user:profile:%d" // user:profile:123
	UserOnlineKey  = "user:online:%d"  // user:online:123
	UserSettingKey = "user:setting:%d" // user:setting:123

	// 好友关系缓存键前缀
	FriendsListKey   = "user:friends:%d"    // user:friends:123
	BlockedUsersKey  = "user:blocked:%d"    // user:blocked:123
	FriendRequestKey = "user:friend_req:%d" // user:friend_req:123

	// 搜索缓存键前缀
	SearchResultKey = "search:users:%s" // search:users:keyword

	// 缓存过期时间
	UserCacheTTL    = 30 * time.Minute // 用户信息缓存30分钟
	OnlineCacheTTL  = 5 * time.Minute  // 在线状态缓存5分钟
	FriendsCacheTTL = 15 * time.Minute // 好友列表缓存15分钟
	SearchCacheTTL  = 10 * time.Minute // 搜索结果缓存10分钟
)

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	// 创建Redis客户端
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	RedisClient = rdb

	// 使用统一日志系统
	log := applogger.GetDefault()
	if log != nil {
		log.Info("Redis connected successfully", applogger.Fields{
			"addr": cfg.Addr,
			"db":   cfg.DB,
		})
	}

	return nil
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}

// UserCacheRepository 用户缓存仓储
type UserCacheRepository struct {
	redis *redis.Client
}

// NewUserCacheRepository 创建用户缓存仓储实例
func NewUserCacheRepository(redis *redis.Client) *UserCacheRepository {
	return &UserCacheRepository{
		redis: redis,
	}
}

// 用户资料缓存相关方法

// SetUserProfile 缓存用户资料
func (r *UserCacheRepository) SetUserProfile(ctx context.Context, userID uint, profile *model.UserProfile) error {
	key := fmt.Sprintf(UserProfileKey, userID)
	data, err := json.Marshal(profile)
	if err != nil {
		return fmt.Errorf("failed to marshal user profile: %w", err)
	}

	return r.redis.Set(ctx, key, data, UserCacheTTL).Err()
}

// GetUserProfile 获取缓存的用户资料
func (r *UserCacheRepository) GetUserProfile(ctx context.Context, userID uint) (*model.UserProfile, error) {
	key := fmt.Sprintf(UserProfileKey, userID)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to get user profile from cache: %w", err)
	}

	var profile model.UserProfile
	if err := json.Unmarshal([]byte(data), &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user profile: %w", err)
	}

	return &profile, nil
}

// DeleteUserProfile 删除用户资料缓存
func (r *UserCacheRepository) DeleteUserProfile(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(UserProfileKey, userID)
	return r.redis.Del(ctx, key).Err()
}

// 在线状态缓存相关方法

// SetUserOnlineStatus 设置用户在线状态
func (r *UserCacheRepository) SetUserOnlineStatus(ctx context.Context, userID uint, isOnline bool, lastSeen time.Time) error {
	key := fmt.Sprintf(UserOnlineKey, userID)

	status := map[string]interface{}{
		"is_online": isOnline,
		"last_seen": lastSeen.Unix(),
	}

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal online status: %w", err)
	}

	return r.redis.Set(ctx, key, data, OnlineCacheTTL).Err()
}

// GetUserOnlineStatus 获取用户在线状态
func (r *UserCacheRepository) GetUserOnlineStatus(ctx context.Context, userID uint) (isOnline bool, lastSeen time.Time, err error) {
	key := fmt.Sprintf(UserOnlineKey, userID)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return false, time.Time{}, nil // 缓存不存在，返回默认值
		}
		return false, time.Time{}, fmt.Errorf("failed to get online status from cache: %w", err)
	}

	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return false, time.Time{}, fmt.Errorf("failed to unmarshal online status: %w", err)
	}

	isOnline, _ = status["is_online"].(bool)
	if lastSeenUnix, ok := status["last_seen"].(float64); ok {
		lastSeen = time.Unix(int64(lastSeenUnix), 0)
	}

	return isOnline, lastSeen, nil
}

// 好友关系缓存相关方法

// SetFriendsList 缓存好友列表
func (r *UserCacheRepository) SetFriendsList(ctx context.Context, userID uint, friends []*model.Friendship) error {
	key := fmt.Sprintf(FriendsListKey, userID)
	data, err := json.Marshal(friends)
	if err != nil {
		return fmt.Errorf("failed to marshal friends list: %w", err)
	}

	return r.redis.Set(ctx, key, data, FriendsCacheTTL).Err()
}

// GetFriendsList 获取缓存的好友列表
func (r *UserCacheRepository) GetFriendsList(ctx context.Context, userID uint) ([]*model.Friendship, error) {
	key := fmt.Sprintf(FriendsListKey, userID)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to get friends list from cache: %w", err)
	}

	var friends []*model.Friendship
	if err := json.Unmarshal([]byte(data), &friends); err != nil {
		return nil, fmt.Errorf("failed to unmarshal friends list: %w", err)
	}

	return friends, nil
}

// DeleteFriendsList 删除好友列表缓存
func (r *UserCacheRepository) DeleteFriendsList(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(FriendsListKey, userID)
	return r.redis.Del(ctx, key).Err()
}

// SetBlockedUsers 缓存屏蔽用户列表
func (r *UserCacheRepository) SetBlockedUsers(ctx context.Context, userID uint, blockedUsers []*model.BlockedUser) error {
	key := fmt.Sprintf(BlockedUsersKey, userID)
	data, err := json.Marshal(blockedUsers)
	if err != nil {
		return fmt.Errorf("failed to marshal blocked users: %w", err)
	}

	return r.redis.Set(ctx, key, data, FriendsCacheTTL).Err()
}

// GetBlockedUsers 获取缓存的屏蔽用户列表
func (r *UserCacheRepository) GetBlockedUsers(ctx context.Context, userID uint) ([]*model.BlockedUser, error) {
	key := fmt.Sprintf(BlockedUsersKey, userID)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to get blocked users from cache: %w", err)
	}

	var blockedUsers []*model.BlockedUser
	if err := json.Unmarshal([]byte(data), &blockedUsers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal blocked users: %w", err)
	}

	return blockedUsers, nil
}

// DeleteBlockedUsers 删除屏蔽用户列表缓存
func (r *UserCacheRepository) DeleteBlockedUsers(ctx context.Context, userID uint) error {
	key := fmt.Sprintf(BlockedUsersKey, userID)
	return r.redis.Del(ctx, key).Err()
}

// 搜索缓存相关方法

// SetSearchResult 缓存搜索结果
func (r *UserCacheRepository) SetSearchResult(ctx context.Context, keyword string, users []*model.UserProfile) error {
	key := fmt.Sprintf(SearchResultKey, keyword)
	data, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("failed to marshal search result: %w", err)
	}

	return r.redis.Set(ctx, key, data, SearchCacheTTL).Err()
}

// GetSearchResult 获取缓存的搜索结果
func (r *UserCacheRepository) GetSearchResult(ctx context.Context, keyword string) ([]*model.UserProfile, error) {
	key := fmt.Sprintf(SearchResultKey, keyword)
	data, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存不存在
		}
		return nil, fmt.Errorf("failed to get search result from cache: %w", err)
	}

	var users []*model.UserProfile
	if err := json.Unmarshal([]byte(data), &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %w", err)
	}

	return users, nil
}

// 批量删除缓存的辅助方法

// InvalidateUserCache 使用户相关的所有缓存失效
func (r *UserCacheRepository) InvalidateUserCache(ctx context.Context, userID uint) error {
	keys := []string{
		fmt.Sprintf(UserProfileKey, userID),
		fmt.Sprintf(UserOnlineKey, userID),
		fmt.Sprintf(UserSettingKey, userID),
		fmt.Sprintf(FriendsListKey, userID),
		fmt.Sprintf(BlockedUsersKey, userID),
		fmt.Sprintf(FriendRequestKey, userID),
	}

	return r.redis.Del(ctx, keys...).Err()
}

// InvalidateFriendshipCache 使好友关系缓存失效
func (r *UserCacheRepository) InvalidateFriendshipCache(ctx context.Context, userID1, userID2 uint) error {
	keys := []string{
		fmt.Sprintf(FriendsListKey, userID1),
		fmt.Sprintf(FriendsListKey, userID2),
		fmt.Sprintf(FriendRequestKey, userID1),
		fmt.Sprintf(FriendRequestKey, userID2),
	}

	return r.redis.Del(ctx, keys...).Err()
}
