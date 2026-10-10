package http_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	httpdelivery "agri-api/internal/delivery/http"
	"agri-api/internal/logger"

	"github.com/gin-gonic/gin"
)

// SetupRoutes doar înregistrează rutele; dependențele sunt folosite abia la cereri,
// deci putem verifica tabela de rutare fără DB, Redis sau SMTP.
func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	httpdelivery.SetupRoutes(router, httpdelivery.AppDeps{Log: logger.NewLogger("development")})

	expected := map[string]bool{
		"POST /api/auth/login":                     false,
		"GET /api/weather/current":                 false,
		"GET /api/machines":                        false,
		"PATCH /api/field-operations/:id/complete": false,
		"PUT /api/reports/subscription":            false,
		"GET /ws/notifications":                    false,
		"GET /api/audit-log":                       false,
		"POST /api/stock-movements":                false,
		"GET /swagger/*any":                        false,
	}
	routes := router.Routes()
	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := expected[key]; ok {
			expected[key] = true
		}
	}
	for key, found := range expected {
		if !found {
			t.Errorf("ruta %s nu este înregistrată", key)
		}
	}
	// rute șterse: assignments (alocările sunt operațiunile pe teren active), checklistul de start
	// (T10), compatibilitățile mașină ↔ echipament (T11, regula stă doar pe template) și stocurile
	// separate de resurse (T12, stocul e pe resursă și se schimbă prin /stock-movements)
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/api/assignments") ||
			route.Path == "/api/field-operations/:id/checklist" ||
			strings.HasPrefix(route.Path, "/api/implement-compatibilities") ||
			strings.HasPrefix(route.Path, "/api/stocks") {
			t.Errorf("ruta ștearsă %s %s e încă înregistrată", route.Method, route.Path)
		}
	}
	if len(routes) < 100 {
		t.Errorf("mă așteptam la peste 100 de rute, am %d", len(routes))
	}

	// rutele din /api sunt protejate de middleware-ul de autentificare
	for _, path := range []string{"/api/machines", "/api/weather/current?lat=47&lng=28"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s fără token: mă așteptam la 401, am %d", path, rec.Code)
		}
	}
}
