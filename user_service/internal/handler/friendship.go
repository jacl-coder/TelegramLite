package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jacl-coder/telegramlite/user_service/internal/service"
)

// FriendshipHandler 好友关系处理器
type FriendshipHandler struct {
	friendshipService *service.FriendshipService
}

// NewFriendshipHandler 创建好友关系处理器
func NewFriendshipHandler(friendshipService *service.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService: friendshipService,
	}
}

// SendFriendRequest 发送好友请求
func (h *FriendshipHandler) SendFriendRequest(c *gin.Context) {
	fromIDStr := c.Param("user_id")
	fromID, err := strconv.ParseUint(fromIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	var req struct {
		ToID    uint   `json:"to_id" binding:"required"`
		Message string `json:"message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	err = h.friendshipService.SendFriendRequest(uint(fromID), req.ToID, req.Message)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "friend request sent successfully",
	})
}

// GetPendingRequests 获取待处理的好友请求
func (h *FriendshipHandler) GetPendingRequests(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	requests, err := h.friendshipService.GetPendingFriendRequests(uint(userID))
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
		Data:    requests,
	})
}

// AcceptFriendRequest 接受好友请求
func (h *FriendshipHandler) AcceptFriendRequest(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	requestIDStr := c.Param("request_id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid request ID",
		})
		return
	}

	err = h.friendshipService.AcceptFriendRequest(uint(requestID), uint(userID))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "friend request accepted",
	})
}

// RejectFriendRequest 拒绝好友请求
func (h *FriendshipHandler) RejectFriendRequest(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	requestIDStr := c.Param("request_id")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid request ID",
		})
		return
	}

	err = h.friendshipService.RejectFriendRequest(uint(requestID), uint(userID))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "friend request rejected",
	})
}

// GetFriendsList 获取好友列表
func (h *FriendshipHandler) GetFriendsList(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	friends, err := h.friendshipService.GetFriendsList(uint(userID), page, pageSize)
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
		Data:    friends,
	})
}

// DeleteFriend 删除好友
func (h *FriendshipHandler) DeleteFriend(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	friendIDStr := c.Param("friend_id")
	friendID, err := strconv.ParseUint(friendIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid friend ID",
		})
		return
	}

	err = h.friendshipService.DeleteFriend(uint(userID), uint(friendID))
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "friend deleted successfully",
	})
}

// GetMutualFriends 获取共同好友
func (h *FriendshipHandler) GetMutualFriends(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid user ID",
		})
		return
	}

	otherUserIDStr := c.Param("other_user_id")
	otherUserID, err := strconv.ParseUint(otherUserIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "invalid other user ID",
		})
		return
	}

	mutualFriends, err := h.friendshipService.GetMutualFriends(uint(userID), uint(otherUserID))
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
		Data:    mutualFriends,
	})
}
