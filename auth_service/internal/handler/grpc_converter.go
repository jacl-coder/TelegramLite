package handler

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/jacl-coder/telegramlite/auth_service/api/proto"
	"github.com/jacl-coder/telegramlite/auth_service/internal/model"
	"github.com/jacl-coder/telegramlite/auth_service/pkg"
)

// convertDeviceTypeToDomain 转换protobuf设备类型到领域模型
func convertDeviceTypeToDomain(deviceType pb.DeviceType) string {
	switch deviceType {
	case pb.DeviceType_DEVICE_TYPE_WEB:
		return "web"
	case pb.DeviceType_DEVICE_TYPE_IOS:
		return "ios"
	case pb.DeviceType_DEVICE_TYPE_ANDROID:
		return "android"
	case pb.DeviceType_DEVICE_TYPE_DESKTOP:
		return "desktop"
	default:
		return "unknown"
	}
}

// convertDeviceTypeToProto 转换领域模型设备类型到protobuf
func convertDeviceTypeToProto(deviceType string) pb.DeviceType {
	switch deviceType {
	case "web":
		return pb.DeviceType_DEVICE_TYPE_WEB
	case "ios":
		return pb.DeviceType_DEVICE_TYPE_IOS
	case "android":
		return pb.DeviceType_DEVICE_TYPE_ANDROID
	case "desktop":
		return pb.DeviceType_DEVICE_TYPE_DESKTOP
	default:
		return pb.DeviceType_DEVICE_TYPE_UNSPECIFIED
	}
}

// convertUserToProto 转换用户模型到protobuf
func convertUserToProto(user *model.User) *pb.UserInfo {
	if user == nil {
		return nil
	}

	var lastLoginAt *timestamppb.Timestamp
	if user.LastLoginAt != nil {
		lastLoginAt = timestamppb.New(*user.LastLoginAt)
	}

	return &pb.UserInfo{
		Id:          uint64(user.ID),
		Phone:       user.Phone,
		Email:       user.Email,
		Username:    user.Username,
		AvatarUrl:   user.AvatarURL,
		IsActive:    user.IsActive,
		LastLoginAt: lastLoginAt,
		CreatedAt:   timestamppb.New(user.CreatedAt),
		UpdatedAt:   timestamppb.New(user.UpdatedAt),
	}
}

// convertDeviceToProto 转换设备模型到protobuf
func convertDeviceToProto(device *model.Device) *pb.DeviceInfo {
	if device == nil {
		return nil
	}

	var lastSeenAt *timestamppb.Timestamp
	if device.LastSeenAt != nil {
		lastSeenAt = timestamppb.New(*device.LastSeenAt)
	}

	return &pb.DeviceInfo{
		Id:          uint64(device.ID),
		UserId:      uint64(device.UserID),
		DeviceToken: device.DeviceToken,
		DeviceType:  convertDeviceTypeToProto(device.DeviceType),
		DeviceName:  device.DeviceName,
		PushToken:   device.PushToken,
		IsOnline:    device.IsOnline,
		LastSeenAt:  lastSeenAt,
		CreatedAt:   timestamppb.New(device.CreatedAt),
	}
}

// convertTokenToProto 转换Token响应到protobuf
func convertTokenToProto(token *pkg.TokenResponse) *pb.TokenInfo {
	if token == nil {
		return nil
	}

	return &pb.TokenInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		ExpiresIn:    token.ExpiresIn,
	}
}
