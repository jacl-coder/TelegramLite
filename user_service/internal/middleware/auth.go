package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jacl-coder/telegramlite/user_service/internal/client"
)

// AuthMiddleware JWT身份验证中间件
type AuthMiddleware struct {
	authClient *client.AuthClient
}

// NewAuthMiddleware 创建身份验证中间件
func NewAuthMiddleware(authClient *client.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

// RequireAuth 要求身份验证的中间件
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization header获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 验证Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 调用Auth Service验证token
		tokenData, err := m.authClient.VerifyToken(context.Background(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Invalid or expired token",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到context中
		c.Set("user_id", uint(tokenData.UserId))
		c.Set("device_id", uint(tokenData.DeviceId))
		c.Set("device_token", tokenData.DeviceToken)
		c.Set("token", token)

		c.Next()
	}
}

// OptionalAuth 可选身份验证的中间件（不强制要求登录）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有token，继续处理但不设置用户信息
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// token格式错误，继续处理但不设置用户信息
			c.Next()
			return
		}

		token := parts[1]
		tokenData, err := m.authClient.VerifyToken(context.Background(), token)
		if err != nil {
			// token无效，继续处理但不设置用户信息
			c.Next()
			return
		}

		// 设置用户信息
		c.Set("user_id", uint(tokenData.UserId))
		c.Set("device_id", uint(tokenData.DeviceId))
		c.Set("device_token", tokenData.DeviceToken)
		c.Set("token", token)

		c.Next()
	}
}

// GetUserID 从context获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	return userID.(uint), true
}

// GetDeviceID 从context获取设备ID
func GetDeviceID(c *gin.Context) (uint, bool) {
	deviceID, exists := c.Get("device_id")
	if !exists {
		return 0, false
	}
	return deviceID.(uint), true
}

// GetToken 从context获取token
func GetToken(c *gin.Context) (string, bool) {
	token, exists := c.Get("token")
	if !exists {
		return "", false
	}
	return token.(string), true
}
