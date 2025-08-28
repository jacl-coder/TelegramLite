package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"telegramlite/auth_service/config"
	"telegramlite/auth_service/internal/repository"
	"telegramlite/auth_service/pkg/hash"
	"telegramlite/auth_service/pkg/jwtutil"
	"github.com/redis/go-redis/v9"
)

type AuthService interface {
	Register(ctx context.Context, username, password string) error
	Login(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
	Refresh(ctx context.Context, refreshToken string) (newAccess string, newRefresh string, err error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	users repository.UserRepo
	rdb   *redis.Client
	jwt   *jwtutil.JWTManager
	cfg   *ConfigWrapper
}

type ConfigWrapper struct {
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewAuthService(u repository.UserRepo, rdb *redis.Client, jwtm *jwtutil.JWTManager, cfg *config.Config) AuthService {
	return &authService{
		users: u,
		rdb:   rdb,
		jwt:   jwtm,
		cfg: &ConfigWrapper{
			AccessTTL:  time.Duration(cfg.AccessTTLMin) * time.Minute,
			RefreshTTL: time.Duration(cfg.RefreshTTLHr) * time.Hour,
		},
	}
}

func (s *authService) Register(ctx context.Context, username, password string) error {
	// check exists
	if _, err := s.users.GetByUsername(ctx, username); err == nil {
		return errors.New("user exists")
	}
	h, err := hash.HashPassword(password)
	if err != nil {
		return err
	}
	u := &repository.User{Username: username, PasswordHash: h, Role: "user"}
	return s.users.Create(ctx, u)
}

func (s *authService) Login(ctx context.Context, username, password string) (string, string, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	if !hash.CheckPassword(u.PasswordHash, password) {
		return "", "", errors.New("invalid credentials")
	}
	at, err := s.jwt.Mint(u.ID, u.Username, u.Role)
	if err != nil {
		return "", "", err
	}
	// create refresh token (simple random string)
	rt := fmt.Sprintf("rt-%d-%d", u.ID, time.Now().UnixNano())
	// store to redis
	if err := s.rdb.Set(ctx, "refresh:"+rt, u.ID, s.cfg.RefreshTTL).Err(); err != nil {
		return "", "", err
	}
	return at, rt, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	key := "refresh:" + refreshToken
	uid, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", "", errors.New("invalid refresh")
	}
	// delete old -> rotate
	if err := s.rdb.Del(ctx, key).Err(); err != nil {
		return "", "", err
	}
	// convert uid to int and lookup user
	var userID int64
	_, err = fmt.Sscanf(uid, "%d", &userID)
	if err != nil {
		return "", "", err
	}
	// fetch user by ID to get username and role
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return "", "", errors.New("user not found")
	}
	// mint new access token with correct user info
	at, err := s.jwt.Mint(u.ID, u.Username, u.Role)
	if err != nil {
		return "", "", err
	}
	// generate new refresh token
	newRT := fmt.Sprintf("rt-%d-%d", u.ID, time.Now().UnixNano())
	if err := s.rdb.Set(ctx, "refresh:"+newRT, u.ID, s.cfg.RefreshTTL).Err(); err != nil {
		return "", "", err
	}
	return at, newRT, nil
}

func (s *authService) Logout(ctx context.Context, refreshToken string) error {
	return s.rdb.Del(ctx, "refresh:"+refreshToken).Err()
}
