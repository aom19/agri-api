package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/store"
	"agri-api/internal/usecase"
)

func newUserStore() (*store.Store, *userRepoMock, *roleRepoMock) {
	users := &userRepoMock{
		getByID: func(id int64) (*domain.User, error) {
			if id == 1 {
				return &domain.User{ID: 1, Email: "ana@x.ro", RoleID: 1}, nil
			}
			return nil, nil
		},
		getByEmail: func(email string) (*domain.User, error) {
			if email == "ana@x.ro" {
				return &domain.User{ID: 1, Email: email}, nil
			}
			return nil, nil
		},
		create: func(u *domain.User) error { u.ID = 1; return nil },
		getAll: func() ([]domain.User, error) {
			return []domain.User{{ID: 1, Email: "ana@x.ro"}, {ID: 2, Email: "b@x.ro", Disabled: true}}, nil
		},
	}
	roles := &roleRepoMock{getByID: func(id int64) (*domain.Role, error) {
		if id == 1 {
			return &domain.Role{ID: 1, Code: "admin", Name: "Administrator"}, nil
		}
		return nil, nil
	}}
	return &store.Store{UserRepo: users, RoleRepo: roles}, users, roles
}

func TestUserService_Read(t *testing.T) {
	st, users, _ := newUserStore()
	svc := usecase.NewUserService(st)

	if _, err := svc.GetUserByID(9); err == nil {
		t.Error("utilizator inexistent trebuie să dea eroare")
	}
	if u, err := svc.GetUserByID(1); err != nil || u.Email != "ana@x.ro" {
		t.Errorf("GetUserByID: %v", err)
	}
	if all, err := svc.GetAllUsers(); err != nil || len(all) != 2 {
		t.Errorf("GetAllUsers: %v", err)
	}

	// nume din profil
	users.getProfile = func(int64) (*domain.UserProfile, error) {
		return &domain.UserProfile{FirstName: "Ana", LastName: "Pop"}, nil
	}
	if name, err := svc.GetUserDisplayName(1); err != nil || name != "Ana Pop" {
		t.Errorf("nume din profil: %v, %q", err, name)
	}
	// profil fără nume → email
	users.getProfile = func(int64) (*domain.UserProfile, error) { return &domain.UserProfile{}, nil }
	if name, _ := svc.GetUserDisplayName(1); name != "ana@x.ro" {
		t.Errorf("fallback pe email: %q", name)
	}
	// utilizator găsit doar prin GetAll
	if name, err := svc.GetUserDisplayName(2); err != nil || name != "b@x.ro" {
		t.Errorf("fallback prin GetAll: %v, %q", err, name)
	}
	if _, err := svc.GetUserDisplayName(9); err == nil {
		t.Error("utilizator inexistent trebuie să dea eroare")
	}
	users.getByID = func(int64) (*domain.User, error) { return nil, errors.New("db down") }
	if _, err := svc.GetUserDisplayName(1); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}

func TestUserService_CreateUpdate(t *testing.T) {
	st, users, _ := newUserStore()
	svc := usecase.NewUserService(st)

	if _, err := svc.CreateUser("ana@x.ro", "x", 1, false); err == nil {
		t.Error("emailul folosit trebuie să dea eroare")
	}
	if _, err := svc.CreateUser("nou@x.ro", "x", 9, false); err == nil {
		t.Error("rolul inexistent trebuie să dea eroare")
	}

	var created *domain.User
	confirmed := false
	users.create = func(u *domain.User) error { u.ID = 1; created = u; return nil }
	users.markEmailConfirmed = func(int64) error { confirmed = true; return nil }
	user, err := svc.CreateUser("nou@x.ro", "", 1, true)
	if err != nil || user == nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if created.PasswordHash == "" || created.RoleCode != "admin" || !confirmed {
		t.Errorf("utilizatorul creat greșit: %+v (confirmat=%v)", created, confirmed)
	}

	if _, err := svc.UpdateUser(9, "x@x.ro", 1, true, ""); err == nil {
		t.Error("update pe utilizator inexistent trebuie să dea eroare")
	}
	users.getByEmail = func(email string) (*domain.User, error) { return &domain.User{ID: 2, Email: email}, nil }
	if _, err := svc.UpdateUser(1, "b@x.ro", 1, true, ""); err == nil {
		t.Error("emailul altui utilizator trebuie să dea eroare")
	}
	users.getByEmail = func(string) (*domain.User, error) { return nil, nil }
	if _, err := svc.UpdateUser(1, "ana2@x.ro", 9, true, ""); err == nil {
		t.Error("rolul inexistent trebuie să dea eroare")
	}
	passwordUpdated := false
	users.updatePassword = func(int64, string) error { passwordUpdated = true; return nil }
	if _, err := svc.UpdateUser(1, "ana2@x.ro", 1, true, "Parola1!"); err != nil || !passwordUpdated {
		t.Errorf("UpdateUser cu parolă: %v (parola actualizată=%v)", err, passwordUpdated)
	}

	boom := errors.New("db down")
	users.update = func(*domain.User) error { return boom }
	if _, err := svc.UpdateUser(1, "ana2@x.ro", 1, true, ""); !errors.Is(err, boom) {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	users.create = func(*domain.User) error { return boom }
	if _, err := svc.CreateUser("alt@x.ro", "x", 1, false); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
}

func TestUserService_DeleteDisableEnable(t *testing.T) {
	st, users, _ := newUserStore()
	svc := usecase.NewUserService(st)

	if err := svc.DeleteUser(9); err == nil {
		t.Error("ștergerea unui utilizator inexistent trebuie să dea eroare")
	}
	if err := svc.DeleteUser(1); err != nil {
		t.Errorf("DeleteUser: %v", err)
	}
	if err := svc.DisableUsers(9); err == nil {
		t.Error("dezactivarea unui utilizator inexistent trebuie să dea eroare")
	}
	if err := svc.DisableUsers(1); err != nil {
		t.Errorf("DisableUsers: %v", err)
	}
	if err := svc.EnableUsers(1); err == nil {
		t.Error("utilizatorul deja activ trebuie să dea eroare")
	}
	if err := svc.EnableUsers(2); err != nil {
		t.Errorf("EnableUsers: %v", err)
	}
	if err := svc.EnableUsers(9); err == nil {
		t.Error("activarea unui utilizator inexistent trebuie să dea eroare")
	}

	boom := errors.New("db down")
	users.getAll = func() ([]domain.User, error) { return nil, boom }
	if err := svc.EnableUsers(2); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	users.getByID = func(int64) (*domain.User, error) { return nil, boom }
	if err := svc.DeleteUser(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
	if err := svc.DisableUsers(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la dezactivare")
	}
}
