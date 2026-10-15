package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/delivery/http/middleware"
	"agri-api/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

func init() { gin.SetMode(gin.TestMode) }

type permissionRepo struct {
	allowed bool
	err     error
}

func (p *permissionRepo) GetAllPermissions() ([]domain.Permission, error)     { return nil, nil }
func (p *permissionRepo) GetPermissionByID(int64) (*domain.Permission, error) { return nil, nil }
func (p *permissionRepo) HasPermission(int64, string) (bool, error)           { return p.allowed, p.err }

func TestAuthMiddleware(t *testing.T) {
	jwtService := auth.NewJWTService("secret", 15*time.Minute, time.Hour)
	blacklist := auth.NewBlacklist(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond}))

	router := gin.New()
	router.Use(middleware.AuthMiddleware(jwtService, blacklist))
	router.GET("/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetString("user_id"), "role_id": c.MustGet("role_id")})
	})
	call := func(header string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/me", nil)
		if header != "" {
			req.Header.Set("Authorization", header)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	if rec := call(""); rec.Code != http.StatusUnauthorized {
		t.Errorf("fără header: %d", rec.Code)
	}
	if rec := call("Basic abc"); rec.Code != http.StatusUnauthorized {
		t.Errorf("schemă greșită: %d", rec.Code)
	}
	if rec := call("Bearer nu.e.token"); rec.Code != http.StatusUnauthorized {
		t.Errorf("token invalid: %d", rec.Code)
	}
	// token valid cu jti, dar Redis-ul e inaccesibil → tratat ca revocat
	withJTI, _ := jwtService.GenerateAccess("1", 2, "admin", "Admin")
	if rec := call("Bearer " + withJTI); rec.Code != http.StatusUnauthorized {
		t.Errorf("verificarea blacklist-ului eșuată: %d", rec.Code)
	}
	// token fără jti → nu se verifică blacklist-ul, cererea trece
	noJTI, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "1", "role_id": 2, "role_code": "admin", "role_name": "Admin", "exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("secret"))
	rec := call("Bearer " + noJTI)
	if rec.Code != http.StatusOK || rec.Body.String() != `{"role_id":2,"user_id":"1"}` {
		t.Errorf("token valid: %d %s", rec.Code, rec.Body.String())
	}
}

func TestRequirePermission(t *testing.T) {
	repo := &permissionRepo{}
	router := gin.New()
	router.GET("/x", func(c *gin.Context) {
		if v := c.Query("role"); v != "" {
			switch v {
			case "str":
				c.Set("role_id", "2")
			case "zero":
				c.Set("role_id", float64(0))
			default:
				c.Set("role_id", float64(2))
			}
		}
	}, middleware.RequirePermission(repo, "machines:read"), func(c *gin.Context) { c.Status(http.StatusOK) })
	call := func(query string) int {
		req := httptest.NewRequest(http.MethodGet, "/x"+query, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := call(""); code != http.StatusUnauthorized {
		t.Errorf("fără role_id: %d", code)
	}
	if code := call("?role=str"); code != http.StatusForbidden {
		t.Errorf("role_id de tip greșit: %d", code)
	}
	if code := call("?role=zero"); code != http.StatusForbidden {
		t.Errorf("role_id zero: %d", code)
	}
	if code := call("?role=ok"); code != http.StatusForbidden {
		t.Errorf("fără permisiune: %d", code)
	}
	repo.err = errors.New("db down")
	if code := call("?role=ok"); code != http.StatusInternalServerError {
		t.Errorf("eroare la verificare: %d", code)
	}
	repo.err = nil
	repo.allowed = true
	if code := call("?role=ok"); code != http.StatusOK {
		t.Errorf("cu permisiune: %d", code)
	}
}

// Cu parolă provizorie (revendicarea must_change_password), trec doar rutele necesare schimbării
// parolei; restul primesc 403 cu codul password_change_required. Fără revendicare, totul trece.
func TestRequirePasswordChanged(t *testing.T) {
	jwtService := auth.NewJWTService("secret", 15*time.Minute, time.Hour)
	router := gin.New()
	// fără jti, ca blacklist-ul să nu fie consultat
	router.Use(func(c *gin.Context) {
		claims, err := jwtService.Parse(c.GetHeader("Authorization")[len("Bearer "):])
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if must, _ := claims[auth.ClaimMustChangePassword].(bool); must {
			c.Set(auth.ClaimMustChangePassword, true)
		}
		c.Next()
	}, middleware.RequirePasswordChanged())
	ok := func(c *gin.Context) { c.Status(http.StatusOK) }
	router.POST("/api/auth/change-password", ok)
	router.POST("/api/auth/logout", ok)
	router.GET("/api/profile", ok)
	router.GET("/api/fields", ok)
	router.POST("/api/field-operations", ok)

	sign := func(claims jwt.MapClaims) string {
		claims["exp"] = time.Now().Add(time.Hour).Unix()
		token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
		return "Bearer " + token
	}
	call := func(method, path, header string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", header)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	provisional := sign(jwt.MapClaims{"user_id": "5", auth.ClaimMustChangePassword: true})
	for _, route := range [][2]string{{http.MethodPost, "/api/auth/change-password"}, {http.MethodPost, "/api/auth/logout"}, {http.MethodGet, "/api/profile"}} {
		if rec := call(route[0], route[1], provisional); rec.Code != http.StatusOK {
			t.Errorf("%s %s trebuie permisă cu parolă provizorie: %d", route[0], route[1], rec.Code)
		}
	}
	for _, route := range [][2]string{{http.MethodGet, "/api/fields"}, {http.MethodPost, "/api/field-operations"}} {
		rec := call(route[0], route[1], provisional)
		if rec.Code != http.StatusForbidden || !strings.Contains(rec.Body.String(), middleware.ErrCodePasswordChangeRequired) {
			t.Errorf("%s %s trebuie blocată cu parolă provizorie: %d %s", route[0], route[1], rec.Code, rec.Body.String())
		}
	}
	normal := sign(jwt.MapClaims{"user_id": "5"})
	if rec := call(http.MethodGet, "/api/fields", normal); rec.Code != http.StatusOK {
		t.Errorf("fără revendicare totul trece: %d", rec.Code)
	}
}
