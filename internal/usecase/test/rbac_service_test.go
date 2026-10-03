package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/store"
	"agri-api/internal/usecase"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
)

type silentRedisLogger struct{}

func (silentRedisLogger) Printf(context.Context, string, ...interface{}) {}

// deadBlacklist folosește un Redis inaccesibil: apelurile eșuează rapid, iar serviciile
// ignoră eroarea (blacklist-ul e best-effort).
func deadBlacklist() *auth.Blacklist {
	redis.SetLogger(silentRedisLogger{})
	return auth.NewBlacklist(redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond,
	}))
}

func newRBACStore() (*store.Store, *roleRepoMock, *permissionRepoMock, *userRepoMock) {
	roles := &roleRepoMock{
		getByID: func(id int64) (*domain.Role, error) {
			switch id {
			case 1:
				return &domain.Role{ID: 1, Code: "admin", Name: "Administrator"}, nil
			case 2:
				return &domain.Role{ID: 2, Code: "viewer", Name: "Vizitator"}, nil
			}
			return nil, nil
		},
		getByCode: func(code string) (*domain.Role, error) {
			if code == "admin" {
				return &domain.Role{ID: 1, Code: "admin"}, nil
			}
			return nil, nil
		},
		create: func(r *domain.Role) error { r.ID = 3; return nil },
	}
	perms := &permissionRepoMock{getPermissionByID: func(id int64) (*domain.Permission, error) {
		if id == 1 {
			return &domain.Permission{ID: 1, Name: "machines:read"}, nil
		}
		return nil, nil
	}}
	users := &userRepoMock{}
	return &store.Store{RoleRepo: roles, PermissionRepo: perms, UserRepo: users}, roles, perms, users
}

func TestRBACService_Roles(t *testing.T) {
	st, roles, _, _ := newRBACStore()
	svc := usecase.NewRBACService(st, nil, nil, nil)

	if _, err := svc.GetAllRoles(); err != nil {
		t.Errorf("GetAllRoles: %v", err)
	}
	if _, err := svc.GetRoleByID(9); err == nil {
		t.Error("rol inexistent trebuie să dea eroare")
	}
	if r, err := svc.GetRoleByID(1); err != nil || r.Code != "admin" {
		t.Errorf("GetRoleByID: %v", err)
	}

	if _, err := svc.CreateRole("admin", "x", ""); err == nil {
		t.Error("codul duplicat trebuie să dea eroare")
	}
	created, err := svc.CreateRole("manager", "Manager", "desc")
	if err != nil || created.ID != 3 {
		t.Fatalf("CreateRole: %v", err)
	}

	if _, err := svc.UpdateRole(9, "x", "y", ""); err == nil {
		t.Error("update pe rol inexistent trebuie să dea eroare")
	}
	updated, err := svc.UpdateRole(2, "viewer2", "V2", "d")
	if err != nil || updated.Code != "viewer2" || updated.Name != "V2" {
		t.Fatalf("UpdateRole: %v, %+v", err, updated)
	}

	if err := svc.DeleteRole(9); err == nil {
		t.Error("ștergerea unui rol inexistent trebuie să dea eroare")
	}
	if err := svc.DeleteRole(1); err == nil {
		t.Error("rolul admin nu poate fi șters")
	}
	if err := svc.DeleteRole(2); err != nil {
		t.Errorf("DeleteRole: %v", err)
	}

	if _, err := svc.GetRolePermissions(9); err == nil {
		t.Error("permisiunile unui rol inexistent trebuie să dea eroare")
	}
	if _, err := svc.GetRolePermissions(1); err != nil {
		t.Errorf("GetRolePermissions: %v", err)
	}
	if err := svc.SetRolePermissions(9, nil); err == nil {
		t.Error("setarea permisiunilor pe rol inexistent trebuie să dea eroare")
	}
	if err := svc.SetRolePermissions(1, []int64{1}); err != nil {
		t.Errorf("SetRolePermissions: %v", err)
	}

	boom := errors.New("db down")
	roles.create = func(*domain.Role) error { return boom }
	if _, err := svc.CreateRole("x", "y", ""); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	roles.update = func(*domain.Role) error { return boom }
	if _, err := svc.UpdateRole(2, "x", "y", ""); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	roles.getByCode = func(string) (*domain.Role, error) { return nil, boom }
	if _, err := svc.CreateRole("x", "y", ""); !errors.Is(err, boom) {
		t.Error("eroarea de citire după cod trebuie propagată")
	}
	roles.getByID = func(int64) (*domain.Role, error) { return nil, boom }
	if _, err := svc.GetRoleByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if _, err := svc.UpdateRole(1, "x", "y", ""); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la update")
	}
	if err := svc.DeleteRole(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la delete")
	}
	if _, err := svc.GetRolePermissions(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la permisiuni")
	}
	if err := svc.SetRolePermissions(1, nil); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la setarea permisiunilor")
	}
}

func TestRBACService_Permissions(t *testing.T) {
	st, _, perms, _ := newRBACStore()
	svc := usecase.NewRBACService(st, nil, nil, nil)

	if _, err := svc.GetAllPermissions(); err != nil {
		t.Errorf("GetAllPermissions: %v", err)
	}
	if _, err := svc.GetPermissionByID(9); err == nil {
		t.Error("permisiune inexistentă trebuie să dea eroare")
	}
	if p, err := svc.GetPermissionByID(1); err != nil || p.Name != "machines:read" {
		t.Errorf("GetPermissionByID: %v", err)
	}
	perms.getPermissionByID = func(int64) (*domain.Permission, error) { return nil, errors.New("db down") }
	if _, err := svc.GetPermissionByID(1); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}

func TestRBACService_AssignRoleToUser(t *testing.T) {
	st, roles, _, users := newRBACStore()
	db, mock := newSQLMock(t)
	jwt := auth.NewJWTService("secret", time.Minute, time.Hour)
	svc := usecase.NewRBACService(st, jwt, auth.NewRepo(db), deadBlacklist())

	if err := svc.AssignRoleToUser(5, 9); err == nil {
		t.Error("rolul inexistent trebuie să dea eroare")
	}

	var assigned int64
	users.updateRole = func(_ int64, roleID int64) error { assigned = roleID; return nil }
	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"access_jti"}).AddRow("jti-1").AddRow(nil))
	if err := svc.AssignRoleToUser(5, 2); err != nil || assigned != 2 {
		t.Fatalf("AssignRoleToUser: %v (rol=%d)", err, assigned)
	}

	boom := errors.New("db down")
	users.updateRole = func(int64, int64) error { return boom }
	if err := svc.AssignRoleToUser(5, 2); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	roles.getByID = func(int64) (*domain.Role, error) { return nil, boom }
	if err := svc.AssignRoleToUser(5, 2); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}
