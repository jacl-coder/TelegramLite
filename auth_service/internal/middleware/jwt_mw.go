package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jacl-coder/auth_service/pkg/jwtutil"
)

func jwtMiddleware(jwtm *jwtutil.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tok := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwtm.Parse(tok)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// inject user info
		if sub, ok := claims["sub"]; ok {
			c.Set("user", sub)
		}
		c.Next()
	}
}
