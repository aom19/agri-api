package middleware

import (
	"agri-api/internal/repository"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePermission returnează un middleware care verifică în DB dacă rolul
// utilizatorului autentificat are permisiunea cerută.
// Trebuie aplicat după AuthMiddleware (depinde de "role_id" din context).
func RequirePermission(permRepo repository.PermissionRepository, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDRaw, exists := c.Get("role_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}

		// JWT MapClaims stochează numerele ca float64
		roleIDFloat, ok := roleIDRaw.(float64)
		if !ok || roleIDFloat <= 0 || roleIDFloat > math.MaxInt64 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid role in token"})
			return
		}
		roleID := int64(roleIDFloat)

		allowed, err := permRepo.HasPermission(roleID, permission)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "authorization check failed"})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}
