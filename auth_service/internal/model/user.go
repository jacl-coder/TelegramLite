package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
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

	// 关联关系
	Devices []Device `json:"devices,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Device 设备模型
type Device struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	UserID      uint           `json:"user_id" gorm:"not null;index;comment:用户ID"`
	DeviceToken string         `json:"device_token" gorm:"uniqueIndex;size:255;comment:设备唯一标识"`
	DeviceType  string         `json:"device_type" gorm:"size:20;comment:设备类型:ios/android/web/desktop"`
	DeviceName  string         `json:"device_name" gorm:"size:100;comment:设备名称"`
	PushToken   string         `json:"push_token" gorm:"size:255;comment:推送token"`
	IsOnline    bool           `json:"is_online" gorm:"default:false;comment:是否在线"`
	LastSeenAt  *time.Time     `json:"last_seen_at" gorm:"comment:最后活跃时间"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}

// UserSession 用户会话模型 (存储在Redis中的结构)
type UserSession struct {
	UserID      uint      `json:"user_id"`
	DeviceID    uint      `json:"device_id"`
	DeviceToken string    `json:"device_token"`
	LoginTime   time.Time `json:"login_time"`
	ExpireTime  time.Time `json:"expire_time"`
}

// DeviceType 设备类型常量
const (
	DeviceTypeIOS     = "ios"
	DeviceTypeAndroid = "android"
	DeviceTypeWeb     = "web"
	DeviceTypeDesktop = "desktop"
)

// ValidDeviceTypes 有效的设备类型列表
var ValidDeviceTypes = []string{
	DeviceTypeIOS,
	DeviceTypeAndroid,
	DeviceTypeWeb,
	DeviceTypeDesktop,
}
