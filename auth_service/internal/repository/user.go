package repository

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/jacl-coder/telegramlite/auth_service/internal/model"
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

// CreateUser 创建用户
func (r *UserRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

// GetUserByPhone 根据手机号获取用户
func (r *UserRepository) GetUserByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ? AND is_active = ?", phone, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (r *UserRepository) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 根据ID获取用户
func (r *UserRepository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (r *UserRepository) UpdateUser(user *model.User) error {
	return r.db.Save(user).Error
}

// UpdateLastLoginAt 更新最后登录时间
func (r *UserRepository) UpdateLastLoginAt(userID uint) error {
	now := time.Now()
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("last_login_at", &now).Error
}

// DeviceRepository 设备数据访问层
type DeviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 创建设备repository
func NewDeviceRepository() *DeviceRepository {
	return &DeviceRepository{
		db: GetDB(),
	}
}

// CreateDevice 创建设备
func (r *DeviceRepository) CreateDevice(device *model.Device) error {
	return r.db.Create(device).Error
}

// GetDeviceByToken 根据设备token获取设备
func (r *DeviceRepository) GetDeviceByToken(deviceToken string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("device_token = ?", deviceToken).First(&device).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

// GetUserDevices 获取用户的所有设备
func (r *DeviceRepository) GetUserDevices(userID uint) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Where("user_id = ?", userID).Find(&devices).Error
	return devices, err
}

// UpdateDeviceOnlineStatus 更新设备在线状态
func (r *DeviceRepository) UpdateDeviceOnlineStatus(deviceID uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
	}
	if !isOnline {
		updates["last_seen_at"] = time.Now()
	}
	return r.db.Model(&model.Device{}).Where("id = ?", deviceID).Updates(updates).Error
}

// DeleteDevice 删除设备
func (r *DeviceRepository) DeleteDevice(deviceID uint) error {
	return r.db.Delete(&model.Device{}, deviceID).Error
}
