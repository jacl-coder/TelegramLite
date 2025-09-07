package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/jacl-coder/telegramlite/user_service/api/proto"
	"github.com/jacl-coder/telegramlite/user_service/internal/service"
)

// UserGRPCHandler gRPC处理器
type UserGRPCHandler struct {
	pb.UnimplementedUserServiceServer
	userService       *service.UserService
	friendshipService *service.FriendshipService
}

// NewUserGRPCHandler 创建新的gRPC处理器
func NewUserGRPCHandler(userSvc *service.UserService, friendshipSvc *service.FriendshipService) *UserGRPCHandler {
	return &UserGRPCHandler{
		userService:       userSvc,
		friendshipService: friendshipSvc,
	}
}

// GetUserProfile 获取用户资料
func (h *UserGRPCHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.GetUserProfileResponse, error) {
	profile, err := h.userService.GetUserProfile(uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user profile: %v", err)
	}

	return &pb.GetUserProfileResponse{
		Profile: convertUserProfileToProto(profile),
	}, nil
}

// UpdateUserProfile 更新用户资料
func (h *UserGRPCHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateUserProfileRequest) (*pb.UpdateUserProfileResponse, error) {
	// 转换proto到内部模型
	profileReq := convertUpdateUserProfileRequest(req)

	updatedProfile, err := h.userService.UpdateUserProfile(uint(req.UserId), profileReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user profile: %v", err)
	}

	return &pb.UpdateUserProfileResponse{
		Profile: convertUserProfileToProto(updatedProfile),
	}, nil
}

// SearchUsers 搜索用户
func (h *UserGRPCHandler) SearchUsers(ctx context.Context, req *pb.SearchUsersRequest) (*pb.SearchUsersResponse, error) {
	// 注意：proto定义中的query字段对应我们的keyword参数
	users, err := h.userService.SearchUsers(req.Query, int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to search users: %v", err)
	}

	protoUsers := make([]*pb.UserProfile, len(users))
	for i, user := range users {
		protoUsers[i] = convertUserProfileToProto(user)
	}

	return &pb.SearchUsersResponse{
		Users:    protoUsers,
		Total:    uint32(len(users)), // 简化处理，实际应该从数据库获取总数
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// SendFriendRequest 发送好友请求
func (h *UserGRPCHandler) SendFriendRequest(ctx context.Context, req *pb.SendFriendRequestRequest) (*pb.SendFriendRequestResponse, error) {
	err := h.friendshipService.SendFriendRequest(uint(req.UserId), uint(req.FriendId), req.Message)
	if err != nil {
		return &pb.SendFriendRequestResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.SendFriendRequestResponse{
		Success: true,
		Message: "friend request sent successfully",
	}, nil
}

// HandleFriendRequest 处理好友请求
func (h *UserGRPCHandler) HandleFriendRequest(ctx context.Context, req *pb.HandleFriendRequestRequest) (*pb.HandleFriendRequestResponse, error) {
	var err error
	if req.Accept {
		err = h.friendshipService.AcceptFriendRequest(uint(req.FriendshipId), uint(req.UserId))
	} else {
		err = h.friendshipService.RejectFriendRequest(uint(req.FriendshipId), uint(req.UserId))
	}

	if err != nil {
		return &pb.HandleFriendRequestResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	message := "friend request rejected"
	if req.Accept {
		message = "friend request accepted"
	}

	return &pb.HandleFriendRequestResponse{
		Success: true,
		Message: message,
	}, nil
}

// GetFriendsList 获取好友列表
func (h *UserGRPCHandler) GetFriendsList(ctx context.Context, req *pb.GetFriendsListRequest) (*pb.GetFriendsListResponse, error) {
	// 根据status参数决定调用哪个方法
	if req.Status == "pending" {
		// 获取待处理的好友请求
		requests, err := h.friendshipService.GetPendingFriendRequests(uint(req.UserId))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get pending friend requests: %v", err)
		}

		// 转换为Friendship格式（简化处理）
		friendships := make([]*pb.Friendship, len(requests))
		for i, request := range requests {
			friendships[i] = &pb.Friendship{
				Id:       uint32(request.ID),
				UserId:   uint32(request.FromID),
				FriendId: uint32(request.ToID),
				Status:   request.Status,
			}
		}

		return &pb.GetFriendsListResponse{
			Friendships: friendships,
			Total:       uint32(len(requests)),
			Page:        req.Page,
			PageSize:    req.PageSize,
		}, nil
	} else {
		// 获取已接受的好友列表
		friends, err := h.friendshipService.GetFriendsList(uint(req.UserId), int(req.Page), int(req.PageSize))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get friends list: %v", err)
		}

		protoFriends := make([]*pb.Friendship, len(friends))
		for i, friend := range friends {
			protoFriends[i] = convertFriendshipToProto(friend)
		}

		return &pb.GetFriendsListResponse{
			Friendships: protoFriends,
			Total:       uint32(len(friends)),
			Page:        req.Page,
			PageSize:    req.PageSize,
		}, nil
	}
}

// RemoveFriend 删除好友
func (h *UserGRPCHandler) RemoveFriend(ctx context.Context, req *pb.RemoveFriendRequest) (*pb.RemoveFriendResponse, error) {
	err := h.friendshipService.DeleteFriend(uint(req.UserId), uint(req.FriendId))
	if err != nil {
		return &pb.RemoveFriendResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &pb.RemoveFriendResponse{
		Success: true,
		Message: "friend deleted successfully",
	}, nil
}

// BlockUser 屏蔽用户 (暂未实现)
func (h *UserGRPCHandler) BlockUser(ctx context.Context, req *pb.BlockUserRequest) (*pb.BlockUserResponse, error) {
	// TODO: 实现用户屏蔽功能
	return &pb.BlockUserResponse{
		Success: false,
		Message: "block user feature not implemented yet",
	}, nil
}

// GetUserSettings 获取用户设置
func (h *UserGRPCHandler) GetUserSettings(ctx context.Context, req *pb.GetUserSettingsRequest) (*pb.GetUserSettingsResponse, error) {
	settings, err := h.userService.GetUserSettings(uint(req.UserId))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user settings: %v", err)
	}

	return &pb.GetUserSettingsResponse{
		Settings: convertUserSettingsToProto(settings),
	}, nil
}

// UpdateUserSettings 更新用户设置
func (h *UserGRPCHandler) UpdateUserSettings(ctx context.Context, req *pb.UpdateUserSettingsRequest) (*pb.UpdateUserSettingsResponse, error) {
	// 转换proto到内部模型
	settingsReq := convertUpdateUserSettingsRequest(req)

	updatedSettings, err := h.userService.UpdateUserSettings(uint(req.UserId), settingsReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user settings: %v", err)
	}

	return &pb.UpdateUserSettingsResponse{
		Settings: convertUserSettingsToProto(updatedSettings),
	}, nil
}

// UpdateOnlineStatus 更新在线状态
func (h *UserGRPCHandler) UpdateOnlineStatus(ctx context.Context, req *pb.UpdateOnlineStatusRequest) (*pb.UpdateOnlineStatusResponse, error) {
	status := "offline"
	if req.IsOnline {
		status = "online"
	}

	err := h.userService.UpdateUserStatus(uint(req.UserId), status)
	if err != nil {
		return &pb.UpdateOnlineStatusResponse{
			Success: false,
		}, nil
	}

	return &pb.UpdateOnlineStatusResponse{
		Success: true,
	}, nil
}
