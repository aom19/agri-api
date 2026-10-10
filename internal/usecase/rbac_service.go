package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"context"
	"errors"
)

type RBACService struct {
	roles       repository.RoleRepository
	permissions repository.PermissionRepository
	users       repository.UserRepository
	jwtService  *auth.JWTService
	refreshRepo *auth.Repo
	blacklist   *auth.Blacklist
}

func NewRBACService(
	roles repository.RoleRepository,
	permissions repository.PermissionRepository,
	users repository.UserRepository,
	jwt *auth.JWTService,
	refreshRepo *auth.Repo,
	blacklist *auth.Blacklist,
) *RBACService {
	return &RBACService{
		roles:       roles,
		permissions: permissions,
		users:       users,
		jwtService:  jwt,
		refreshRepo: refreshRepo,
		blacklist:   blacklist,
	}
}

// ─── Roles ────────────────────────────────────────────────────────────────────

// Rolurile sunt fixe, definite în migrația 000054_fixed_roles_and_permissions; aici doar se citesc.

func (s *RBACService) GetAllRoles() ([]domain.Role, error) {
	return s.roles.GetAll()
}

func (s *RBACService) GetRoleByID(id int64) (*domain.Role, error) {
	role, err := s.roles.GetByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return role, nil
}

// ─── Permissions ──────────────────────────────────────────────────────────────

func (s *RBACService) GetAllPermissions() ([]domain.Permission, error) {
	return s.permissions.GetAllPermissions()
}

func (s *RBACService) GetPermissionByID(id int64) (*domain.Permission, error) {
	permission, err := s.permissions.GetPermissionByID(id)
	if err != nil {
		return nil, err
	}
	if permission == nil {
		return nil, errors.New("permission not found")
	}
	return permission, nil
}

func (s *RBACService) GetRolePermissions(roleID int64) ([]domain.Permission, error) {
	role, err := s.roles.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return s.roles.GetPermissions(roleID)
}

// ─── User role assignment ─────────────────────────────────────────────────────

// AssignRoleToUser schimbă rolul unui user și revocă toate sesiunile active ale
// acestuia (forțând re-login cu noul role_id în JWT).
func (s *RBACService) AssignRoleToUser(userID int64, roleID int64) error {
	role, err := s.roles.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	if err := s.users.UpdateRole(userID, roleID); err != nil {
		return err
	}

	// Revocă toate sesiunile active — forțează re-login cu noul role_id
	oldJTIs, _ := s.refreshRepo.RevokeAllUserSessions(userID)
	for _, jti := range oldJTIs {
		_ = s.blacklist.Add(context.Background(), jti, s.jwtService.AccessTokenTTL())
	}

	return nil
}
