package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"telegramlite/auth_service/pkg/jwtutil"
)

func jwtMiddleware(jwtm *jwtutil.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtm.Parse(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user", claims["name"])
		c.Set("role", claims["role"])
		c.Next()
	}
}
