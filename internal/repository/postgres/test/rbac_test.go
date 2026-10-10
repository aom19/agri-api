package postgres_test

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"agri-api/internal/auth"
	httpdelivery "agri-api/internal/delivery/http"
	"agri-api/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// T18: rolurile și permisiunile sunt fixe, definite doar în migrația 054.
const fixedRolesMigration = "000054_fixed_roles_and_permissions"

// requireRBAC întoarce permisiunile fiecărui rol de după migrații (vezi rbacAfterMigrations).
func requireRBAC(t *testing.T) map[string][]string {
	t.Helper()
	requireDB(t)
	return rbacAfterMigrations
}

// permissionRecorder notează permisiunea cerută de fiecare rută și refuză cererea, ca
// handler-ul să nu mai ruleze.
type permissionRecorder struct {
	repository.PermissionRepository
	checked []string
}

func (r *permissionRecorder) HasPermission(_ int64, name string) (bool, error) {
	r.checked = append(r.checked, name)
	return false, nil
}

// routePermissions întoarce, pentru fiecare rută din /api, permisiunea pe care o verifică
// (sau "" pentru rutele care cer doar autentificare).
func routePermissions(t *testing.T) map[string]string {
	t.Helper()
	gin.SetMode(gin.TestMode)
	const secret = "test-secret"
	recorder := &permissionRecorder{}
	router := gin.New()
	// rutele fără permisiune ajung la handler, care nu are dependențe; panica devine 500
	router.Use(gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		c.AbortWithStatus(http.StatusInternalServerError)
	}))
	httpdelivery.SetupRoutes(router, httpdelivery.AppDeps{
		PermissionRepo: recorder,
		JWTService:     auth.NewJWTService(secret, time.Minute, time.Minute),
	})

	// fără jti, middleware-ul de autentificare nu consultă blacklist-ul din Redis
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "1", "role_id": 1, "exp": time.Now().Add(time.Minute).Unix(),
	}).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}

	result := map[string]string{}
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/") {
			continue
		}
		recorder.checked = nil
		req := httptest.NewRequest(route.Method, strings.ReplaceAll(route.Path, ":id", "1"), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(httptest.NewRecorder(), req)

		key := route.Method + " " + route.Path
		switch len(recorder.checked) {
		case 0:
			result[key] = ""
		case 1:
			result[key] = recorder.checked[0]
		default:
			t.Errorf("%s verifică mai multe permisiuni: %v", key, recorder.checked)
		}
	}
	return result
}

func queryStrings(t *testing.T, db *sql.DB, query string, args ...interface{}) []string {
	t.Helper()
	rows, err := db.Query(query, args...)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			t.Fatal(err)
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return values
}

// Fiecare permisiune din baza de date e verificată de cel puțin o rută, iar fiecare permisiune
// cerută de o rută există în baza de date (altfel ruta n-ar putea fi folosită nici de admin).
func TestRBAC_EveryPermissionIsCheckedByARoute(t *testing.T) {
	inDB := map[string]bool{}
	for _, name := range requireRBAC(t)[""] {
		inDB[name] = true
	}
	checked := map[string]bool{}
	for route, name := range routePermissions(t) {
		if name == "" {
			continue
		}
		checked[name] = true
		if !inDB[name] {
			t.Errorf("%s cere %q, care nu există în baza de date", route, name)
		}
	}
	for name := range inDB {
		if !checked[name] {
			t.Errorf("permisiunea %q nu e verificată de nicio rută", name)
		}
	}
	// un singur format: modul:read, modul:write și excepția field_operations:execute
	for name := range inDB {
		if !strings.HasSuffix(name, ":read") && !strings.HasSuffix(name, ":write") && name != "field_operations:execute" {
			t.Errorf("permisiunea %q nu respectă formatul modul:read / modul:write", name)
		}
	}
	if len(inDB) != 24 {
		t.Errorf("mă așteptam la 24 de permisiuni, am %d", len(inDB))
	}
}

// Rutele fără permisiune sunt doar cele ale contului propriu și cele publice de autentificare.
func TestRBAC_RoutesWithoutPermission(t *testing.T) {
	var open []string
	for route, name := range routePermissions(t) {
		if name == "" {
			open = append(open, route)
		}
	}
	sort.Strings(open)
	want := []string{
		"GET /api/auth/me/permissions",
		"GET /api/health",
		"GET /api/profile",
		"GET /api/weather/current",
		"PATCH /api/profile",
		"POST /api/auth/change-password",
		"POST /api/auth/confirm-email",
		"POST /api/auth/forgot-password",
		"POST /api/auth/login",
		"POST /api/auth/logout",
		"POST /api/auth/refresh",
		"POST /api/auth/register",
		"POST /api/auth/resend-confirmation",
		"POST /api/auth/reset-password/:token",
		"POST /api/profile/photo",
	}
	if strings.Join(open, "\n") != strings.Join(want, "\n") {
		t.Errorf("rute fără permisiune:\n%s\nașteptate:\n%s", strings.Join(open, "\n"), strings.Join(want, "\n"))
	}
}

func rolePermissions(t *testing.T, db *sql.DB, code string) []string {
	t.Helper()
	rbac, err := readRBAC(db)
	if err != nil {
		t.Fatal(err)
	}
	return rbac[code]
}

// Cele 4 roluri fixe și permisiunile lor, așa cum le definește migrația 054.
func TestRBAC_FixedRoles(t *testing.T) {
	rbac := requireRBAC(t)
	if roles := queryStrings(t, testDB, `SELECT code FROM roles ORDER BY code`); strings.Join(roles, ",") != "admin,manager,operator,viewer" {
		t.Fatalf("roluri: %v", roles)
	}
	if got := rbac["admin"]; strings.Join(got, ",") != strings.Join(rbac[""], ",") {
		t.Errorf("admin trebuie să aibă toate permisiunile, are %v", got)
	}

	want := map[string][]string{
		"manager": {
			"crops:read", "crops:write", "dashboard:read",
			"field_operations:execute", "field_operations:read", "field_operations:write",
			"fields:read", "fields:write", "implements:read", "implements:write",
			"machines:read", "machines:write", "notifications:read",
			"operations:read", "operations:write", "operators:read", "operators:write",
			"reports:read", "resources:read", "resources:write",
		},
		"operator": {
			"crops:read", "dashboard:read", "field_operations:execute", "field_operations:read",
			"notifications:read",
		},
		"viewer": {
			"crops:read", "dashboard:read", "field_operations:read", "fields:read",
			"implements:read", "machines:read", "notifications:read", "operations:read",
			"operators:read", "reports:read", "resources:read",
		},
	}
	for code, perms := range want {
		if got := rbac[code]; strings.Join(got, ",") != strings.Join(perms, ",") {
			t.Errorf("%s:\n are      %v\n așteptat %v", code, got, perms)
		}
	}
}

// Migrația 054 șterge rolurile create din aplicație și mută conturile lor pe viewer; down
// readuce cele 50 de permisiuni vechi.
func TestRBAC_MigrationFixesRoles(t *testing.T) {
	db := requireDB(t)
	runMigrationFile(t, db, fixedRolesMigration+".down.sql")
	applied := false
	migrateUp := func() {
		t.Helper()
		applied = true
		runMigrationFile(t, db, fixedRolesMigration+".up.sql")
	}
	t.Cleanup(func() {
		if !applied {
			migrateUp()
		}
	})

	if n := len(queryStrings(t, db, `SELECT name FROM permissions`)); n != 50 {
		t.Errorf("down trebuie să readucă cele 50 de permisiuni vechi, am %d", n)
	}
	if got := rolePermissions(t, db, "operator"); strings.Join(got, ",") != "crops:read,dashboard:read,field_operations:complete,field_operations:read,field_operations:start,notifications:read" {
		t.Errorf("operator după down: %v", got)
	}

	// rol creat din aplicație, cu numele unui rol fix și cu un cont
	testC := insertID(t, db, `INSERT INTO roles (code, name) VALUES ('testc', 'TestC') RETURNING id`)
	if _, err := db.Exec(`UPDATE roles SET name = 'Vizualizator (vechi)' WHERE code = 'viewer'`); err != nil {
		t.Fatal(err)
	}
	clash := insertID(t, db, `INSERT INTO roles (code, name) VALUES ('viewer_copy', 'Vizualizator') RETURNING id`)
	user := insertID(t, db, `INSERT INTO users (email, password_hash, role_id) VALUES ('testc@test.ro', '!', $1) RETURNING id`, testC)
	if _, err := db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT $1, id FROM permissions WHERE name IN ('machines:delete', 'resource.view')`, testC); err != nil {
		t.Fatal(err)
	}

	migrateUp()

	var roleCode string
	if err := db.QueryRow(`SELECT r.code FROM users u JOIN roles r ON r.id = u.role_id WHERE u.id = $1`, user).Scan(&roleCode); err != nil || roleCode != "viewer" {
		t.Errorf("contul cu rolul TestC trebuie mutat pe viewer: %q, %v", roleCode, err)
	}
	var left int
	if err := db.QueryRow(`SELECT COUNT(*) FROM roles WHERE id IN ($1, $2)`, testC, clash).Scan(&left); err != nil || left != 0 {
		t.Errorf("rolurile create din aplicație trebuie șterse: %d, %v", left, err)
	}
	if name := queryStrings(t, db, `SELECT name FROM roles WHERE code = 'viewer'`); len(name) != 1 || name[0] != "Vizualizator" {
		t.Errorf("rolul viewer își recapătă numele: %v", name)
	}
	if old := queryStrings(t, db, `
		SELECT name FROM permissions
		WHERE name IN ('resource.view', 'permissions:read', 'machines:delete', 'users:enable', 'field_operations:start')`); len(old) != 0 {
		t.Errorf("permisiunile vechi trebuie șterse: %v", old)
	}
	// reaplicată, migrația duce baza exact în starea de după migrații
	after, err := readRBAC(db)
	if err != nil {
		t.Fatal(err)
	}
	for role, perms := range rbacAfterMigrations {
		if strings.Join(after[role], ",") != strings.Join(perms, ",") {
			t.Errorf("%q după down + up:\n are      %v\n așteptat %v", role, after[role], perms)
		}
	}
}
