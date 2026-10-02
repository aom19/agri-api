package handlers

import (
	"net/http"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/store"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// userRepoStub implementează doar metodele folosite la înregistrare; restul ar da panică.
type userRepoStub struct {
	repository.UserRepository
	created *domain.User
}

func (m *userRepoStub) GetByEmail(string) (*domain.User, error) { return nil, nil }
func (m *userRepoStub) Create(u *domain.User) error {
	u.ID = 1
	m.created = u
	return nil
}

type roleRepoStub struct {
	repository.RoleRepository
	requested []string
}

func (m *roleRepoStub) GetByCode(code string) (*domain.Role, error) {
	m.requested = append(m.requested, code)
	return &domain.Role{ID: 99, Code: code, Name: code}, nil
}

func TestAuthHandler_Register_IgnoresRole(t *testing.T) {
	users, roles := &userRepoStub{}, &roleRepoStub{}
	// fără serviciu de e-mail: contul se creează, apoi trimiterea confirmării eșuează
	svc := usecase.NewAuthService(&store.Store{UserRepo: users, RoleRepo: roles}, nil, nil, nil, nil, "")
	r := gin.New()
	r.POST("/register", NewAuthHandler(svc).Register)

	rec := do(t, r, http.MethodPost, "/register", `{"email":"oricine@exemplu.com","password":"parola123","role":"admin"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("mă așteptam la 400 (e-mail neconfigurat), nu %d: %s", rec.Code, rec.Body)
	}
	if users.created == nil {
		t.Fatal("contul nu a fost creat")
	}
	if users.created.RoleCode != "viewer" {
		t.Errorf("rolul din cerere nu trebuie folosit: contul a primit %q", users.created.RoleCode)
	}
	if len(roles.requested) != 1 || roles.requested[0] != "viewer" {
		t.Errorf("roluri cerute: %v", roles.requested)
	}
}
