package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jacl-coder/telegramlite/auth_service/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "注册成功",
		Data:    result,
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "登录成功",
		Data:    result,
	})
}

// RefreshToken 刷新token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	result, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "刷新成功",
		Data:    result,
	})
}

// Logout 登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 从JWT中获取设备ID (这里简化处理，实际应该从middleware中获取)
	deviceID := c.GetUint("device_id")
	if deviceID == 0 {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的设备信息",
		})
		return
	}

	err := h.authService.Logout(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "登出成功",
	})
}

// GetUserInfo 获取用户信息
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// 从Header获取Access Token
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "access token is required",
		})
		return
	}

	// 去掉 Bearer 前缀
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 验证Token并获取用户信息
	userInfo, err := h.authService.GetUserInfo(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, Response{
			Code:    401,
			Message: "invalid or expired token",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "user info retrieved successfully",
		Data:    userInfo,
	})
}

// Health 健康检查
func (h *AuthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "Auth Service is running",
		Data: gin.H{
			"service": "auth-service",
			"status":  "healthy",
		},
	})
}
