package handler

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/jacl-coder/telegramlite/auth_service/api/proto"
	"github.com/jacl-coder/telegramlite/auth_service/internal/service"
)

// GRPCAuthHandler gRPC认证处理器
type GRPCAuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService *service.AuthService
}

// NewGRPCAuthHandler 创建新的gRPC认证处理器
func NewGRPCAuthHandler(authService *service.AuthService) *GRPCAuthHandler {
	return &GRPCAuthHandler{
		authService: authService,
	}
}

// Register 用户注册
func (h *GRPCAuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// 转换请求到service层
	serviceReq := &service.RegisterRequest{
		Phone:       req.Phone,
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		DeviceToken: req.DeviceToken,
		DeviceType:  convertDeviceTypeToDomain(req.DeviceType),
		DeviceName:  req.DeviceName,
	}

	// 调用业务逻辑
	resp, err := h.authService.Register(serviceReq)
	if err != nil {
		return &pb.RegisterResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   err.Error(),
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.RegisterResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "注册成功",
			Timestamp: timestamppb.Now(),
		},
		Data: &pb.RegisterData{
			User:   convertUserToProto(resp.User),
			Device: convertDeviceToProto(resp.Device),
			Token:  convertTokenToProto(resp.Token),
		},
	}, nil
}

// Login 用户登录
func (h *GRPCAuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	serviceReq := &service.LoginRequest{
		DeviceToken: req.DeviceToken,
		Password:    req.Password,
		DeviceType:  convertDeviceTypeToDomain(req.DeviceType),
		DeviceName:  req.DeviceName,
	}

	// 处理登录凭证
	switch cred := req.Credential.(type) {
	case *pb.LoginRequest_Phone:
		serviceReq.Phone = cred.Phone
	case *pb.LoginRequest_Email:
		serviceReq.Email = cred.Email
	default:
		return &pb.LoginResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   "请提供有效的登录凭证",
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	// 调用业务逻辑
	resp, err := h.authService.Login(serviceReq)
	if err != nil {
		return &pb.LoginResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   err.Error(),
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.LoginResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "登录成功",
			Timestamp: timestamppb.Now(),
		},
		Data: &pb.LoginData{
			User:   convertUserToProto(resp.User),
			Device: convertDeviceToProto(resp.Device),
			Token:  convertTokenToProto(resp.Token),
		},
	}, nil
}

// RefreshToken 刷新Token
func (h *GRPCAuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	token, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		return &pb.RefreshTokenResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   err.Error(),
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.RefreshTokenResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "Token刷新成功",
			Timestamp: timestamppb.Now(),
		},
		Token: convertTokenToProto(token),
	}, nil
}

// Logout 用户注销
func (h *GRPCAuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	// 首先需要根据device_token获取device_id
	// 这里我们需要添加一个helper方法到service
	err := h.authService.LogoutByDeviceToken(req.DeviceToken)
	if err != nil {
		return &pb.LogoutResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   err.Error(),
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.LogoutResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "注销成功",
			Timestamp: timestamppb.Now(),
		},
	}, nil
}

// VerifyToken 验证Token (给其他服务调用)
func (h *GRPCAuthHandler) VerifyToken(ctx context.Context, req *pb.VerifyTokenRequest) (*pb.VerifyTokenResponse, error) {
	claims, err := h.authService.ParseToken(req.AccessToken)
	if err != nil {
		return &pb.VerifyTokenResponse{
			Response: &pb.Response{
				Code:      401,
				Message:   "Token无效",
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.VerifyTokenResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "Token验证成功",
			Timestamp: timestamppb.Now(),
		},
		Data: &pb.VerifyTokenData{
			Valid:       true,
			UserId:      uint64(claims.UserID),
			DeviceId:    uint64(claims.DeviceID),
			DeviceToken: claims.DeviceToken,
			ExpiresAt:   timestamppb.New(claims.ExpiresAt.Time),
		},
	}, nil
}

// GetUserInfo 获取用户信息
func (h *GRPCAuthHandler) GetUserInfo(ctx context.Context, req *pb.GetUserInfoRequest) (*pb.GetUserInfoResponse, error) {
	user, err := h.authService.GetUserByToken(req.AccessToken)
	if err != nil {
		return &pb.GetUserInfoResponse{
			Response: &pb.Response{
				Code:      400,
				Message:   err.Error(),
				Timestamp: timestamppb.Now(),
			},
		}, nil
	}

	return &pb.GetUserInfoResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "获取用户信息成功",
			Timestamp: timestamppb.Now(),
		},
		User: convertUserToProto(user),
	}, nil
}

// Health 健康检查
func (h *GRPCAuthHandler) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Response: &pb.Response{
			Code:      0,
			Message:   "Auth Service is running",
			Timestamp: timestamppb.Now(),
		},
		Data: &pb.HealthData{
			Service:   "auth-service",
			Status:    "healthy",
			Timestamp: timestamppb.Now(),
		},
	}, nil
}
