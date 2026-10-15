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
		if must, _ := claims[auth.ClaimMustChangePassword].(bool); must {
			c.Set(auth.ClaimMustChangePassword, true)
		}
		c.Next()
	}
}

// ErrCodePasswordChangeRequired e codul din răspunsul 403 când contul are parolă provizorie.
const ErrCodePasswordChangeRequired = "password_change_required"

// passwordChangeAllowed sunt rutele acceptate cât timp parola e provizorie: schimbarea ei,
// profilul (ca interfața să știe cine e conectat) și deconectarea.
var passwordChangeAllowed = map[string]bool{
	"POST /api/auth/change-password": true,
	"POST /api/auth/logout":          true,
	"GET /api/profile":               true,
	"GET /api/auth/me/permissions":   true,
}

// RequirePasswordChanged blochează restul API-ului pentru conturile cu parolă provizorie
// (revendicarea ClaimMustChangePassword din token), cu 403 și un cod pe care interfața îl
// transformă în redirecționare către schimbarea parolei.
func RequirePasswordChanged() gin.HandlerFunc {
	return func(c *gin.Context) {
		must, _ := c.Get(auth.ClaimMustChangePassword)
		if must != true {
			c.Next()
			return
		}
		if passwordChangeAllowed[c.Request.Method+" "+c.FullPath()] {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Parola provizorie trebuie schimbată înainte de a folosi aplicația.",
			"code":  ErrCodePasswordChangeRequired,
		})
	}
}
