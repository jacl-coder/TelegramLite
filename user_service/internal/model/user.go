package model

import (
	"time"

	"gorm.io/gorm"
)

// User 基础用户模型（引用Auth Service中的用户表）
type User struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	Phone        string         `json:"phone" gorm:"uniqueIndex;size:20;comment:手机号"`
	Email        string         `json:"email" gorm:"uniqueIndex;size:255;comment:邮箱"`
	Username     string         `json:"username" gorm:"uniqueIndex;size:50;comment:用户名"`
	PasswordHash string         `json:"-" gorm:"size:255;comment:密码哈希"`
	AvatarURL    string         `json:"avatar_url" gorm:"size:500;comment:头像URL"`
	IsActive     bool           `json:"is_active" gorm:"default:true;comment:是否激活"`
	LastLoginAt  *time.Time     `json:"last_login_at" gorm:"comment:最后登录时间"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// FriendRequest 好友请求
type FriendRequest struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	FromID    uint           `json:"from_id" gorm:"not null;index;comment:发起用户ID"`
	ToID      uint           `json:"to_id" gorm:"not null;index;comment:目标用户ID"`
	Message   string         `json:"message" gorm:"size:200;comment:请求消息"`
	Status    string         `json:"status" gorm:"size:20;default:'pending';comment:状态:pending/accepted/rejected"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (FriendRequest) TableName() string {
	return "friend_requests"
}

// UserProfile 用户档案扩展信息
type UserProfile struct {
	ID         uint           `json:"id" gorm:"primarykey"`
	UserID     uint           `json:"user_id" gorm:"uniqueIndex;not null;comment:用户ID"`
	Nickname   string         `json:"nickname" gorm:"size:50;comment:昵称"`
	FirstName  string         `json:"first_name" gorm:"size:50;comment:名"`
	LastName   string         `json:"last_name" gorm:"size:50;comment:姓"`
	Bio        string         `json:"bio" gorm:"size:500;comment:个人简介"`
	Avatar     string         `json:"avatar" gorm:"size:500;comment:头像URL"`
	Status     string         `json:"status" gorm:"size:100;comment:个性状态"`
	Birthday   *time.Time     `json:"birthday" gorm:"comment:生日"`
	Gender     string         `json:"gender" gorm:"size:10;comment:性别:male/female/other"`
	Language   string         `json:"language" gorm:"size:10;default:'zh-CN';comment:语言偏好"`
	Timezone   string         `json:"timezone" gorm:"size:50;default:'Asia/Shanghai';comment:时区"`
	IsOnline   bool           `json:"is_online" gorm:"default:false;comment:是否在线"`
	LastSeenAt *time.Time     `json:"last_seen_at" gorm:"comment:最后在线时间"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (UserProfile) TableName() string {
	return "user_profiles"
}

// FriendshipStatus 好友关系状态
type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"  // 待确认
	FriendshipAccepted FriendshipStatus = "accepted" // 已接受
	FriendshipBlocked  FriendshipStatus = "blocked"  // 已屏蔽
)

// Friendship 好友关系
type Friendship struct {
	ID        uint             `json:"id" gorm:"primarykey"`
	UserID    uint             `json:"user_id" gorm:"not null;index;comment:发起用户ID"`
	FriendID  uint             `json:"friend_id" gorm:"not null;index;comment:目标用户ID"`
	Status    FriendshipStatus `json:"status" gorm:"type:varchar(20);default:'pending';comment:关系状态"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `json:"-" gorm:"index"`
}

// TableName 指定表名
func (Friendship) TableName() string {
	return "friendships"
}

// UserSetting 用户设置
type UserSetting struct {
	ID                   uint           `json:"id" gorm:"primarykey"`
	UserID               uint           `json:"user_id" gorm:"uniqueIndex;not null;comment:用户ID"`
	AllowFriendRequests  bool           `json:"allow_friend_requests" gorm:"default:true;comment:允许好友请求"`
	AllowBeingSearched   bool           `json:"allow_being_searched" gorm:"default:true;comment:允许被搜索"`
	ShowOnlineStatus     bool           `json:"show_online_status" gorm:"default:true;comment:显示在线状态"`
	ShowLastSeen         bool           `json:"show_last_seen" gorm:"default:true;comment:显示最后在线时间"`
	MessageNotifications bool           `json:"message_notifications" gorm:"default:true;comment:消息通知"`
	FriendNotifications  bool           `json:"friend_notifications" gorm:"default:true;comment:好友通知"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (UserSetting) TableName() string {
	return "user_settings"
}

// BlockedUser 屏蔽用户
type BlockedUser struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	UserID    uint           `json:"user_id" gorm:"not null;index;comment:屏蔽发起用户ID"`
	BlockedID uint           `json:"blocked_id" gorm:"not null;index;comment:被屏蔽用户ID"`
	Reason    string         `json:"reason" gorm:"size:200;comment:屏蔽原因"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName 指定表名
func (BlockedUser) TableName() string {
	return "blocked_users"
}

// GetFullName 获取完整姓名
func (up *UserProfile) GetFullName() string {
	if up.FirstName != "" && up.LastName != "" {
		return up.FirstName + " " + up.LastName
	}
	if up.Nickname != "" {
		return up.Nickname
	}
	return ""
}

// GetDisplayName 获取显示名称
func (up *UserProfile) GetDisplayName() string {
	if up.Nickname != "" {
		return up.Nickname
	}
	return up.GetFullName()
}
