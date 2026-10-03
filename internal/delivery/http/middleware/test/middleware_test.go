package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
