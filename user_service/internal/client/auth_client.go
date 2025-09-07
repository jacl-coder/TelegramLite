package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authpb "github.com/jacl-coder/telegramlite/auth_service/api/proto"
)

// AuthClient Auth Service客户端
type AuthClient struct {
	client authpb.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthClient 创建Auth Service客户端
func NewAuthClient(authServiceURL string) (*AuthClient, error) {
	// 建立gRPC连接
	conn, err := grpc.NewClient(authServiceURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(4*1024*1024)), // 4MB
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	client := authpb.NewAuthServiceClient(conn)

	return &AuthClient{
		client: client,
		conn:   conn,
	}, nil
}

// VerifyToken 验证Token
func (c *AuthClient) VerifyToken(ctx context.Context, token string) (*authpb.VerifyTokenData, error) {
	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 调用Auth Service验证Token
	resp, err := c.client.VerifyToken(ctx, &authpb.VerifyTokenRequest{
		AccessToken: token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// 检查响应
	if resp.Response.Code != 0 {
		return nil, fmt.Errorf("token verification failed: %s", resp.Response.Message)
	}

	return resp.Data, nil
}

// GetUserInfo 获取用户信息
func (c *AuthClient) GetUserInfo(ctx context.Context, token string) (*authpb.UserInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.client.GetUserInfo(ctx, &authpb.GetUserInfoRequest{
		AccessToken: token,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if resp.Response.Code != 0 {
		return nil, fmt.Errorf("get user info failed: %s", resp.Response.Message)
	}

	return resp.User, nil
}

// Close 关闭连接
func (c *AuthClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
