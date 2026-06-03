package middleware

import (
	"agri-api/internal/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(jwt *auth.JWTService, blacklist *auth.Blacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwt.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if jti, ok := claims["jti"].(string); ok && jti != "" {
			revoked, err := blacklist.IsBlacklisted(c.Request.Context(), jti)
			if err != nil || revoked {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
				return
			}
		}

		c.Set("user_id", claims["user_id"])
		c.Set("role_id", claims["role_id"])
		c.Set("role_code", claims["role_code"])
		c.Set("role_name", claims["role_name"])
		c.Next()
	}
}
