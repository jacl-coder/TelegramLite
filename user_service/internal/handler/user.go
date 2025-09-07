package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jacl-coder/telegramlite/user_service/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Response 通用响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Code       int             `json:"code"`
	Message    string          `json:"message"`
	Data       interface{}     `json:"data,omitempty"`
	Pagination *PaginationInfo `json:"pagination,omitempty"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	profile, err := h.userService.GetUserProfile(uint(userID))
	if err != nil {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    profile,
	})
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	profile, err := h.userService.UpdateUserProfile(uint(userID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    profile,
	})
}

// SearchUsers 搜索用户
func (h *UserHandler) SearchUsers(c *gin.Context) {
	keyword := c.Query("keyword")
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	// 获取当前用户ID（可能为0，表示匿名用户）
	var currentUserID uint = 0
	if userIDValue, exists := c.Get("user_id"); exists {
		if uid, ok := userIDValue.(uint); ok {
			currentUserID = uid
		}
	}

	profiles, total, err := h.userService.SearchUsersWithPagination(keyword, limit, offset, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, PaginatedResponse{
		Code:    200,
		Message: "success",
		Data:    profiles,
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// UpdateStatus 更新用户状态
func (h *UserHandler) UpdateStatus(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	err = h.userService.UpdateUserStatus(uint(userID), req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "status updated successfully",
	})
}

// GetSettings 获取用户设置
func (h *UserHandler) GetSettings(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	settings, err := h.userService.GetUserSettings(uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    settings,
	})
}

// UpdateSettings 更新用户设置
func (h *UserHandler) UpdateSettings(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	var req service.UpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	settings, err := h.userService.UpdateUserSettings(uint(userID), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "success",
		Data:    settings,
	})
}

// BlockUser 屏蔽用户
func (h *UserHandler) BlockUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	blockedIDStr := c.Param("blocked_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	blockedID, err := strconv.ParseUint(blockedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid blocked user ID",
		})
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	err = h.userService.BlockUser(uint(userID), uint(blockedID), req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "用户屏蔽成功",
	})
}

// UnblockUser 取消屏蔽用户
func (h *UserHandler) UnblockUser(c *gin.Context) {
	userIDStr := c.Param("user_id")
	blockedIDStr := c.Param("blocked_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	blockedID, err := strconv.ParseUint(blockedIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid blocked user ID",
		})
		return
	}

	err = h.userService.UnblockUser(uint(userID), uint(blockedID))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "取消屏蔽成功",
	})
}

// GetBlockedUsers 获取屏蔽用户列表
func (h *UserHandler) GetBlockedUsers(c *gin.Context) {
	userIDStr := c.Param("user_id")
	limitStr := c.DefaultQuery("limit", "20")
	pageStr := c.DefaultQuery("page", "1")

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	// 计算偏移量
	offset := (page - 1) * limit

	profiles, total, err := h.userService.GetBlockedUsers(uint(userID), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	c.JSON(http.StatusOK, PaginatedResponse{
		Code:    200,
		Message: "success",
		Data:    profiles,
		Pagination: &PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}
