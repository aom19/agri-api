package handlers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func TestCurrentUserID(t *testing.T) {
	cases := []struct {
		value interface{}
		want  int64
		ok    bool
	}{
		{"7", 7, true}, {"abc", 0, false}, {"-1", 0, false},
		{float64(3), 3, true}, {float64(0), 0, false},
		{int64(4), 4, true}, {int(5), 5, true}, {int(0), 0, false},
		{true, 0, false}, {nil, 0, false},
	}
	for _, c := range cases {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		if c.value != nil {
			ctx.Set("user_id", c.value)
		}
		got, ok := handlers.CurrentUserID(ctx)
		if got != c.want || ok != c.ok {
			t.Errorf("currentUserID(%v) = %d,%v; vreau %d,%v", c.value, got, ok, c.want, c.ok)
		}
	}
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	if handlers.CurrentActorID(ctx) != nil {
		t.Error("fără utilizator, actorul trebuie să fie nil")
	}
	ctx.Set("user_id", "9")
	ctx.Set("role_code", "operator")
	if actor := handlers.CurrentActorID(ctx); actor == nil || *actor != 9 || !handlers.IsOperatorRequest(ctx) {
		t.Error("actorul și rolul de operator trebuie citite din context")
	}
	if handlers.AuditID(12) != "12" || handlers.StatusChange("a", "b")["old_status"] != "a" {
		t.Error("helper-ele de audit")
	}
}

func TestDashboardHandler(t *testing.T) {
	repo := &dashboardRepoMock{}
	audit := &auditRepoMock{getAll: func(limit int, _, _ string) ([]domain.AuditEntry, error) {
		return []domain.AuditEntry{{ID: int64(limit)}}, nil
	}}
	handler := handlers.NewDashboardHandler(usecase.NewDashboardService(repo, audit))
	router := gin.New()
	router.GET("/cards", handler.GetCards)
	router.GET("/quick-stats", handler.GetQuickStats)
	router.GET("/activity", handler.GetActivity)

	if rec := do(t, router, http.MethodGet, "/cards", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"cards"`) {
		t.Errorf("carduri: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodGet, "/quick-stats", ""); rec.Code != http.StatusOK {
		t.Errorf("statistici: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, "/activity?limit=3", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"id":3`) {
		t.Errorf("activitate cu limită: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodGet, "/activity?limit=x", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"id":5`) {
		t.Errorf("activitate cu limită invalidă: %d %s", rec.Code, rec.Body.String())
	}

	boom := errors.New("db down")
	repo.getCardStats = func() (*domain.DashboardCardStats, error) { return nil, boom }
	repo.getQuickStats = func() (*domain.DashboardQuickStats, error) { return nil, boom }
	audit.getAll = func(int, string, string) ([]domain.AuditEntry, error) { return nil, boom }
	for _, path := range []string{"/cards", "/quick-stats", "/activity"} {
		if rec := do(t, router, http.MethodGet, path, ""); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: eroarea trebuie să dea 500, am %d", path, rec.Code)
		}
	}
}

func TestNotificationHandler(t *testing.T) {
	repo := &notificationRepoMock{
		getByUser: func(userID int64, onlyUnread bool, limit int) ([]domain.UserNotification, error) {
			if onlyUnread {
				return []domain.UserNotification{{ID: 1, UserID: userID}}, nil
			}
			return nil, nil
		},
		countUnread: func(int64) (int64, error) { return 2, nil },
	}
	handler := handlers.NewNotificationHandler(usecase.NewNotificationService(repo))
	build := func(user interface{}) *gin.Engine {
		router := gin.New()
		router.Use(withUser(user))
		router.GET("/notifications", handler.GetAll)
		router.GET("/notifications/count", handler.CountUnread)
		router.PATCH("/notifications/:id/read", handler.MarkAsRead)
		router.PATCH("/notifications/read-all", handler.MarkAllAsRead)
		return router
	}

	anonymous := build(nil)
	for _, r := range [][2]string{{http.MethodGet, "/notifications"}, {http.MethodGet, "/notifications/count"}, {http.MethodPatch, "/notifications/1/read"}, {http.MethodPatch, "/notifications/read-all"}} {
		if rec := do(t, anonymous, r[0], r[1], ""); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s fără utilizator: %d", r[0], r[1], rec.Code)
		}
	}

	router := build("7")
	if rec := do(t, router, http.MethodGet, "/notifications?unread=true&limit=10", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"user_id":7`) {
		t.Errorf("listare necitite: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodGet, "/notifications", ""); rec.Code != http.StatusOK || rec.Body.String() != "[]" {
		t.Errorf("listă goală trebuie serializată ca []: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodGet, "/notifications/count", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"count":2`) {
		t.Errorf("număr necitite: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodPatch, "/notifications/abc/read", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("id invalid: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPatch, "/notifications/1/read", ""); rec.Code != http.StatusNoContent {
		t.Errorf("marcare citită: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPatch, "/notifications/read-all", ""); rec.Code != http.StatusNoContent {
		t.Errorf("marcare toate: %d", rec.Code)
	}

	boom := errors.New("db down")
	repo.getByUser = func(int64, bool, int) ([]domain.UserNotification, error) { return nil, boom }
	repo.countUnread = func(int64) (int64, error) { return 0, boom }
	repo.markAsRead = func(int64, int64) error { return boom }
	repo.markAllAsRead = func(int64) error { return boom }
	for _, r := range [][2]string{{http.MethodGet, "/notifications"}, {http.MethodGet, "/notifications/count"}, {http.MethodPatch, "/notifications/1/read"}, {http.MethodPatch, "/notifications/read-all"}} {
		if rec := do(t, router, r[0], r[1], ""); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s %s cu eroare: %d", r[0], r[1], rec.Code)
		}
	}
}

func TestAuditHandler(t *testing.T) {
	var gotLimit int
	repo := &auditRepoMock{getAll: func(limit int, entityType, entityID string) ([]domain.AuditEntry, error) {
		gotLimit = limit
		if entityType == "machine" {
			return []domain.AuditEntry{{ID: 1, EntityID: entityID}}, nil
		}
		return nil, nil
	}}
	router := gin.New()
	router.GET("/audit-log", handlers.NewAuditHandler(repo).GetAll)

	if rec := do(t, router, http.MethodGet, "/audit-log?entity_type=machine&entity_id=4&limit=20", ""); rec.Code != http.StatusOK || gotLimit != 20 || !strings.Contains(rec.Body.String(), `"entity_id":"4"`) {
		t.Errorf("filtrare: %d %s (limit=%d)", rec.Code, rec.Body.String(), gotLimit)
	}
	if rec := do(t, router, http.MethodGet, "/audit-log?limit=9999", ""); rec.Code != http.StatusOK || gotLimit != 100 || rec.Body.String() != "[]" {
		t.Errorf("limita implicită și lista goală: %d %s (limit=%d)", rec.Code, rec.Body.String(), gotLimit)
	}
	repo.getAll = func(int, string, string) ([]domain.AuditEntry, error) { return nil, errors.New("db down") }
	if rec := do(t, router, http.MethodGet, "/audit-log", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroare: %d", rec.Code)
	}
}

func TestWeatherHandler(t *testing.T) {
	// Serviciul meteo real ar apela internetul; folosim un context deja anulat, astfel încât
	// cererea HTTP către furnizor eșuează imediat, fără rețea.
	handler := handlers.NewWeatherHandler(usecase.NewWeatherService(""))
	router := gin.New()
	router.GET("/weather", handler.GetCurrent)
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	call := func(query string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/weather"+query, nil).WithContext(cancelled)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}

	if rec := call(""); rec.Code != http.StatusBadGateway {
		t.Errorf("locația implicită fără furnizor: %d", rec.Code)
	}
	if rec := call("?lat=46"); rec.Code != http.StatusBadRequest {
		t.Errorf("lat fără lng: %d", rec.Code)
	}
	if rec := call("?lat=abc&lng=28"); rec.Code != http.StatusBadRequest {
		t.Errorf("lat invalid: %d", rec.Code)
	}
	if rec := call("?lat=46&lng=200"); rec.Code != http.StatusBadRequest {
		t.Errorf("lng invalid: %d", rec.Code)
	}
	if rec := call("?lat=46&lng=28&location=Lot"); rec.Code != http.StatusBadGateway {
		t.Errorf("coordonate valide fără furnizor: %d", rec.Code)
	}
}
