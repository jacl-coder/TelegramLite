package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/jacl-coder/telegramlite/user_service/api/proto"
	"github.com/jacl-coder/telegramlite/user_service/internal/model"
	"github.com/jacl-coder/telegramlite/user_service/internal/service"
)

// convertUserProfileToProto 将内部模型转换为Proto消息
func convertUserProfileToProto(profile *model.UserProfile) *proto.UserProfile {
	if profile == nil {
		return nil
	}

	protoProfile := &proto.UserProfile{
		Id:        uint32(profile.ID),
		UserId:    uint32(profile.UserID),
		Nickname:  profile.Nickname,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Bio:       profile.Bio,
		Avatar:    profile.Avatar,
		Status:    profile.Status,
		Gender:    profile.Gender,
		Language:  profile.Language,
		Timezone:  profile.Timezone,
		IsOnline:  profile.IsOnline,
		CreatedAt: timestamppb.New(profile.CreatedAt),
		UpdatedAt: timestamppb.New(profile.UpdatedAt),
	}

	// 处理可能为nil的时间字段
	if profile.Birthday != nil {
		protoProfile.Birthday = timestamppb.New(*profile.Birthday)
	}

	if profile.LastSeenAt != nil {
		protoProfile.LastSeenAt = timestamppb.New(*profile.LastSeenAt)
	}

	return protoProfile
}

// convertUserSettingsToProto 将内部设置模型转换为Proto消息
func convertUserSettingsToProto(settings *model.UserSetting) *proto.UserSettings {
	if settings == nil {
		return nil
	}

	return &proto.UserSettings{
		Id:               uint32(settings.ID),
		UserId:           uint32(settings.UserID),
		PrivacyLevel:     "normal", // 默认隐私级别
		AllowSearch:      settings.AllowBeingSearched,
		AllowFriendReq:   settings.AllowFriendRequests,
		ShowOnlineStatus: settings.ShowOnlineStatus,
		ShowLastSeen:     settings.ShowLastSeen,
		MessagePreview:   settings.MessageNotifications,
		SoundEnabled:     settings.FriendNotifications,
		VibrateEnabled:   true, // 默认值
		PushEnabled:      true, // 默认值
		CreatedAt:        timestamppb.New(settings.CreatedAt),
		UpdatedAt:        timestamppb.New(settings.UpdatedAt),
	}
}

// convertFriendshipToProto 将好友关系转换为Proto消息
func convertFriendshipToProto(friendship *model.Friendship) *proto.Friendship {
	if friendship == nil {
		return nil
	}

	return &proto.Friendship{
		Id:        uint32(friendship.ID),
		UserId:    uint32(friendship.UserID),
		FriendId:  uint32(friendship.FriendID),
		Status:    string(friendship.Status),
		CreatedAt: timestamppb.New(friendship.CreatedAt),
		UpdatedAt: timestamppb.New(friendship.UpdatedAt),
	}
}

// convertUpdateUserProfileRequest 转换更新用户资料请求
func convertUpdateUserProfileRequest(req *proto.UpdateUserProfileRequest) *service.UpdateProfileRequest {
	result := &service.UpdateProfileRequest{}

	if req.Nickname != "" {
		result.Nickname = &req.Nickname
	}
	if req.FirstName != "" {
		result.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		result.LastName = &req.LastName
	}
	if req.Bio != "" {
		result.Bio = &req.Bio
	}
	if req.Avatar != "" {
		result.Avatar = &req.Avatar
	}
	if req.Status != "" {
		result.Status = &req.Status
	}
	if req.Gender != "" {
		result.Gender = &req.Gender
	}
	if req.Language != "" {
		result.Language = &req.Language
	}
	if req.Timezone != "" {
		result.Timezone = &req.Timezone
	}
	if req.Birthday != nil {
		birthday := req.Birthday.AsTime()
		result.Birthday = &birthday
	}

	return result
}

// convertUpdateUserSettingsRequest 转换更新用户设置请求
func convertUpdateUserSettingsRequest(req *proto.UpdateUserSettingsRequest) *service.UpdateSettingsRequest {
	return &service.UpdateSettingsRequest{
		AllowFriendRequests:  &req.AllowFriendReq,
		AllowBeingSearched:   &req.AllowSearch,
		ShowOnlineStatus:     &req.ShowOnlineStatus,
		ShowLastSeen:         &req.ShowLastSeen,
		MessageNotifications: &req.MessagePreview,
		FriendNotifications:  &req.SoundEnabled,
	}
}
