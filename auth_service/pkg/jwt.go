package pkg

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager JWT管理器
type JWTManager struct {
	secretKey       string
	tokenDuration   time.Duration
	refreshDuration time.Duration
}

// Claims JWT声明
type Claims struct {
	UserID      uint   `json:"user_id"`
	DeviceID    uint   `json:"device_id"`
	DeviceToken string `json:"device_token"`
	jwt.RegisteredClaims
}

// RefreshClaims 刷新token声明
type RefreshClaims struct {
	UserID   uint `json:"user_id"`
	DeviceID uint `json:"device_id"`
	jwt.RegisteredClaims
}

// NewJWTManager 创建JWT管理器
func NewJWTManager(secretKey string, tokenDuration, refreshDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:       secretKey,
		tokenDuration:   tokenDuration,
		refreshDuration: refreshDuration,
	}
}

// GenerateToken 生成访问token
func (manager *JWTManager) GenerateToken(userID, deviceID uint, deviceToken string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		DeviceID:    deviceID,
		DeviceToken: deviceToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "telegramlite-auth",
			Subject:   "access-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// GenerateRefreshToken 生成刷新token
func (manager *JWTManager) GenerateRefreshToken(userID, deviceID uint) (string, error) {
	claims := &RefreshClaims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.refreshDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "telegramlite-auth",
			Subject:   "refresh-token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))
}

// VerifyToken 验证访问token
func (manager *JWTManager) VerifyToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// VerifyRefreshToken 验证刷新token
func (manager *JWTManager) VerifyRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, errors.New("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok {
		return nil, errors.New("invalid refresh token claims")
	}

	return claims, nil
}

// TokenResponse token响应结构
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// GenerateTokenPair 生成token对
func (manager *JWTManager) GenerateTokenPair(userID, deviceID uint, deviceToken string) (*TokenResponse, error) {
	accessToken, err := manager.GenerateToken(userID, deviceID, deviceToken)
	if err != nil {
		return nil, err
	}

	refreshToken, err := manager.GenerateRefreshToken(userID, deviceID)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(manager.tokenDuration.Seconds()),
	}, nil
}
