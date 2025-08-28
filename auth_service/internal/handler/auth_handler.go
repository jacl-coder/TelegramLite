package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"telegramlite/auth_service/internal/service"
	"telegramlite/auth_service/pkg/jwtutil"
)

type authReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterRoutes(r *gin.Engine, svc service.AuthService, jwtm *jwtutil.JWTManager) {
	auth := r.Group("/auth")
	{
		auth.POST("/register", makeRegisterHandler(svc))
		auth.POST("/login", makeLoginHandler(svc))
		auth.POST("/refresh", makeRefreshHandler(svc))
		auth.POST("/logout", makeLogoutHandler(svc))
	}

	// protected example
	r.GET("/protected", jwtMiddleware(jwtm), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"msg": "ok", "user": c.GetString("user")})
	})
}

func makeRegisterHandler(svc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Register(context.Background(), req.Username, req.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "created"})
	}
}

func makeLoginHandler(svc service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		at, rt, err := svc.Login(context.Background(), req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": at, "refresh_token": rt})
	}
}

func makeRefreshHandler(svc service.AuthService) gin.HandlerFunc {
	type req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	return func(c *gin.Context) {
		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		at, rt, err := svc.Refresh(context.Background(), r.RefreshToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"access_token": at, "refresh_token": rt})
	}
}

func makeLogoutHandler(svc service.AuthService) gin.HandlerFunc {
	type req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	return func(c *gin.Context) {
		var r req
		if err := c.ShouldBindJSON(&r); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Logout(context.Background(), r.RefreshToken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "logged out"})
	}
}
