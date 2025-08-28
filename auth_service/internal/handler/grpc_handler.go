package handler

import (
	"context"

	pb "telegramlite/auth_service/internal/pb/proto"
	"telegramlite/auth_service/internal/service"
)

type GRPCAuthHandler struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewGRPCAuthHandler(authSvc service.AuthService) *GRPCAuthHandler {
	return &GRPCAuthHandler{
		authService: authSvc,
	}
}

func (h *GRPCAuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	err := h.authService.Register(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.RegisterResponse{
			Success: false,
			Message: "Registration failed",
			Error:   err.Error(),
		}, nil
	}

	return &pb.RegisterResponse{
		Success: true,
		Message: "User registered successfully",
	}, nil
}

func (h *GRPCAuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	accessToken, refreshToken, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return &pb.LoginResponse{
			Success: false,
			Message: "Login failed",
			Error:   err.Error(),
		}, nil
	}

	return &pb.LoginResponse{
		Success:      true,
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *GRPCAuthHandler) Refresh(ctx context.Context, req *pb.RefreshRequest) (*pb.RefreshResponse, error) {
	accessToken, refreshToken, err := h.authService.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return &pb.RefreshResponse{
			Success: false,
			Message: "Token refresh failed",
			Error:   err.Error(),
		}, nil
	}

	return &pb.RefreshResponse{
		Success:      true,
		Message:      "Token refreshed successfully",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (h *GRPCAuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.authService.Logout(ctx, req.RefreshToken)
	if err != nil {
		return &pb.LogoutResponse{
			Success: false,
			Message: "Logout failed",
			Error:   err.Error(),
		}, nil
	}

	return &pb.LogoutResponse{
		Success: true,
		Message: "Logged out successfully",
	}, nil
}