package usecase

import (
	"errors"
	"fmt"
	"strings"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

// adminRoleCode este rolul contului de administrator creat la pornire.
const adminRoleCode = "admin"

// BootstrapAdminResult spune ce a făcut BootstrapAdmin, pentru jurnal.
type BootstrapAdminResult struct {
	Created  bool
	Promoted bool
	Email    string
}

// BootstrapAdmin garantează că instalarea are cel puțin un administrator. Rulează la pornire:
//   - dacă există deja un cont activ cu rolul admin, nu face nimic;
//   - dacă există un cont cu ADMIN_EMAIL, îl promovează la admin (parola rămâne a lui);
//   - altfel creează contul cu parola din ADMIN_PASSWORD, confirmat, cu schimbare obligatorie a
//     parolei la prima autentificare.
//
// Fără ADMIN_EMAIL / ADMIN_PASSWORD, întoarce nil și nu face nimic: pe o bază nouă, singura altă
// cale de a obține un admin e direct în baza de date.
func BootstrapAdmin(users repository.UserRepository, roles repository.RoleRepository, email, password string) (*BootstrapAdminResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, nil
	}
	admins, err := users.CountByRole(adminRoleCode)
	if err != nil {
		return nil, err
	}
	if admins > 0 {
		return nil, nil
	}
	role, err := roles.GetByCode(adminRoleCode)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("rolul admin nu există; rulează seed-ul RBAC înainte")
	}

	existing, err := users.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := users.UpdateRole(existing.ID, role.ID); err != nil {
			return nil, err
		}
		if !existing.EmailConfirmed {
			if err := users.MarkEmailConfirmed(existing.ID); err != nil {
				return nil, err
			}
		}
		return &BootstrapAdminResult{Promoted: true, Email: email}, nil
	}

	if len(password) < 8 {
		return nil, fmt.Errorf("ADMIN_PASSWORD trebuie să aibă cel puțin 8 caractere")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &domain.User{
		Email:              email,
		PasswordHash:       hash,
		RoleID:             role.ID,
		RoleCode:           role.Code,
		RoleName:           role.Name,
		EmailConfirmed:     true,
		MustChangePassword: true,
	}
	if err := users.Create(user); err != nil {
		return nil, err
	}
	return &BootstrapAdminResult{Created: true, Email: email}, nil
}
