package service

import (
	"errors"
	"strings"

	"github.com/jacl-coder/telegramlite/auth_service/internal/model"
	"github.com/jacl-coder/telegramlite/auth_service/internal/repository"
	"github.com/jacl-coder/telegramlite/auth_service/pkg"
)

// AuthService 认证服务
type AuthService struct {
	userRepo        *repository.UserRepository
	deviceRepo      *repository.DeviceRepository
	jwtManager      *pkg.JWTManager
	passwordManager *pkg.PasswordManager
}

// NewAuthService 创建认证服务
func NewAuthService(jwtManager *pkg.JWTManager) *AuthService {
	return &AuthService{
		userRepo:        repository.NewUserRepository(),
		deviceRepo:      repository.NewDeviceRepository(),
		jwtManager:      jwtManager,
		passwordManager: pkg.NewPasswordManager(),
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Phone       string `json:"phone" binding:"required"`
	Email       string `json:"email"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=6"`
	DeviceToken string `json:"device_token" binding:"required"`
	DeviceType  string `json:"device_type" binding:"required"`
	DeviceName  string `json:"device_name"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Password    string `json:"password" binding:"required"`
	DeviceToken string `json:"device_token" binding:"required"`
	DeviceType  string `json:"device_type" binding:"required"`
	DeviceName  string `json:"device_name"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User   *model.User        `json:"user"`
	Device *model.Device      `json:"device"`
	Token  *pkg.TokenResponse `json:"token"`
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	// 验证输入
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("手机号或邮箱必须提供一个")
	}

	if !s.passwordManager.IsValidPassword(req.Password) {
		return nil, errors.New("密码长度至少6位")
	}

	if !isValidDeviceType(req.DeviceType) {
		return nil, errors.New("无效的设备类型")
	}

	// 检查用户是否已存在
	if req.Phone != "" {
		existingUser, err := s.userRepo.GetUserByPhone(req.Phone)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.New("手机号已被注册")
		}
	}

	if req.Email != "" {
		existingUser, err := s.userRepo.GetUserByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.New("邮箱已被注册")
		}
	}

	// 加密密码
	hashedPassword, err := s.passwordManager.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Phone:        req.Phone,
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: hashedPassword,
		IsActive:     true,
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	// 创建设备
	device := &model.Device{
		UserID:      user.ID,
		DeviceToken: req.DeviceToken,
		DeviceType:  req.DeviceType,
		DeviceName:  req.DeviceName,
		IsOnline:    true,
	}

	if err := s.deviceRepo.CreateDevice(device); err != nil {
		return nil, err
	}

	// 生成token
	tokenResponse, err := s.jwtManager.GenerateTokenPair(user.ID, device.ID, device.DeviceToken)
	if err != nil {
		return nil, err
	}

	// 隐藏密码
	user.PasswordHash = ""

	return &AuthResponse{
		User:   user,
		Device: device,
		Token:  tokenResponse,
	}, nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	// 验证输入
	if req.Phone == "" && req.Email == "" {
		return nil, errors.New("手机号或邮箱必须提供一个")
	}

	if !isValidDeviceType(req.DeviceType) {
		return nil, errors.New("无效的设备类型")
	}

	// 获取用户
	var user *model.User
	var err error

	if req.Phone != "" {
		user, err = s.userRepo.GetUserByPhone(req.Phone)
	} else {
		user, err = s.userRepo.GetUserByEmail(req.Email)
	}

	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 验证密码
	if err := s.passwordManager.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, errors.New("密码错误")
	}

	// 检查或创建设备
	device, err := s.deviceRepo.GetDeviceByToken(req.DeviceToken)
	if err != nil {
		return nil, err
	}

	if device == nil {
		// 创建新设备
		device = &model.Device{
			UserID:      user.ID,
			DeviceToken: req.DeviceToken,
			DeviceType:  req.DeviceType,
			DeviceName:  req.DeviceName,
			IsOnline:    true,
		}
		if err := s.deviceRepo.CreateDevice(device); err != nil {
			return nil, err
		}
	} else {
		// 验证设备是否属于该用户
		if device.UserID != user.ID {
			return nil, errors.New("设备已被其他用户使用")
		}
		// 更新设备在线状态
		if err := s.deviceRepo.UpdateDeviceOnlineStatus(device.ID, true); err != nil {
			return nil, err
		}
	}

	// 更新最后登录时间
	if err := s.userRepo.UpdateLastLoginAt(user.ID); err != nil {
		return nil, err
	}

	// 生成token
	tokenResponse, err := s.jwtManager.GenerateTokenPair(user.ID, device.ID, device.DeviceToken)
	if err != nil {
		return nil, err
	}

	// 隐藏密码
	user.PasswordHash = ""

	return &AuthResponse{
		User:   user,
		Device: device,
		Token:  tokenResponse,
	}, nil
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(refreshToken string) (*pkg.TokenResponse, error) {
	// 输入验证
	if refreshToken == "" {
		return nil, errors.New("刷新token不能为空")
	}

	// 验证刷新token
	claims, err := s.jwtManager.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("无效的刷新token")
	}

	// 获取设备信息
	device, err := s.deviceRepo.GetDeviceByToken("")
	if err != nil {
		return nil, err
	}

	if device == nil || device.ID != claims.DeviceID {
		return nil, errors.New("设备不存在")
	}

	// 生成新的token对
	return s.jwtManager.GenerateTokenPair(claims.UserID, claims.DeviceID, device.DeviceToken)
}

// Logout 登出
func (s *AuthService) Logout(deviceID uint) error {
	return s.deviceRepo.UpdateDeviceOnlineStatus(deviceID, false)
}

// isValidDeviceType 验证设备类型
func isValidDeviceType(deviceType string) bool {
	deviceType = strings.ToLower(deviceType)
	for _, validType := range model.ValidDeviceTypes {
		if deviceType == validType {
			return true
		}
	}
	return false
}

// LogoutByDeviceToken 根据设备Token注销
func (s *AuthService) LogoutByDeviceToken(deviceToken string) error {
	device, err := s.deviceRepo.GetDeviceByToken(deviceToken)
	if err != nil {
		return err
	}
	if device == nil {
		return errors.New("设备不存在")
	}
	return s.deviceRepo.UpdateDeviceOnlineStatus(device.ID, false)
}

// ParseToken 解析Token并验证
func (s *AuthService) ParseToken(tokenString string) (*pkg.Claims, error) {
	return s.jwtManager.VerifyToken(tokenString)
}

// GetUserByToken 通过Token获取用户信息
func (s *AuthService) GetUserByToken(tokenString string) (*model.User, error) {
	// 解析token
	claims, err := s.jwtManager.VerifyToken(tokenString)
	if err != nil {
		return nil, errors.New("无效的token")
	}

	// 获取用户信息
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("用户不存在")
	}

	// 隐藏密码
	user.PasswordHash = ""
	return user, nil
}

// GetUserInfo 获取用户信息 (为HTTP API提供)
func (s *AuthService) GetUserInfo(tokenString string) (*model.User, error) {
	// 输入验证
	if tokenString == "" {
		return nil, errors.New("token不能为空")
	}

	return s.GetUserByToken(tokenString)
}
