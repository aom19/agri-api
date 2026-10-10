package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
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

type rbacRepos struct {
	roles *roleRepoMock
	perms *permissionRepoMock
	users *userRepoMock
}

func newRBACRepos() rbacRepos {
	roles := &roleRepoMock{
		getByID: func(id int64) (*domain.Role, error) {
			switch id {
			case 1:
				return &domain.Role{ID: 1, Code: "admin", Name: "Administrator"}, nil
			case 2:
				return &domain.Role{ID: 2, Code: "viewer", Name: "Vizualizator"}, nil
			}
			return nil, nil
		},
	}
	perms := &permissionRepoMock{getPermissionByID: func(id int64) (*domain.Permission, error) {
		if id == 1 {
			return &domain.Permission{ID: 1, Name: "machines:read"}, nil
		}
		return nil, nil
	}}
	return rbacRepos{roles: roles, perms: perms, users: &userRepoMock{}}
}

func (r rbacRepos) service(jwt *auth.JWTService, refresh *auth.Repo, blacklist *auth.Blacklist) *usecase.RBACService {
	return usecase.NewRBACService(r.roles, r.perms, r.users, jwt, refresh, blacklist)
}

// Rolurile sunt fixe (T18): serviciul le citește, dar nu le creează, editează sau șterge.
func TestRBACService_Roles(t *testing.T) {
	repos := newRBACRepos()
	svc := repos.service(nil, nil, nil)

	if _, err := svc.GetAllRoles(); err != nil {
		t.Errorf("GetAllRoles: %v", err)
	}
	if _, err := svc.GetRoleByID(9); err == nil {
		t.Error("rol inexistent trebuie să dea eroare")
	}
	if r, err := svc.GetRoleByID(1); err != nil || r.Code != "admin" {
		t.Errorf("GetRoleByID: %v", err)
	}
	if _, err := svc.GetRolePermissions(9); err == nil {
		t.Error("permisiunile unui rol inexistent trebuie să dea eroare")
	}
	if _, err := svc.GetRolePermissions(1); err != nil {
		t.Errorf("GetRolePermissions: %v", err)
	}

	boom := errors.New("db down")
	repos.roles.getByID = func(int64) (*domain.Role, error) { return nil, boom }
	if _, err := svc.GetRoleByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if _, err := svc.GetRolePermissions(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la permisiuni")
	}
}

func TestRBACService_Permissions(t *testing.T) {
	repos := newRBACRepos()
	perms := repos.perms
	svc := repos.service(nil, nil, nil)

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
	repos := newRBACRepos()
	roles, users := repos.roles, repos.users
	db, mock := newSQLMock(t)
	jwt := auth.NewJWTService("secret", time.Minute, time.Hour)
	svc := repos.service(jwt, auth.NewRepo(db), deadBlacklist())

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
