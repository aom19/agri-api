package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/store"
	"context"
	"errors"
)

type RBACService struct {
	store       *store.Store
	jwtService  *auth.JWTService
	refreshRepo *auth.Repo
	blacklist   *auth.Blacklist
}

func NewRBACService(s *store.Store, jwt *auth.JWTService, refreshRepo *auth.Repo, blacklist *auth.Blacklist) *RBACService {
	return &RBACService{store: s, jwtService: jwt, refreshRepo: refreshRepo, blacklist: blacklist}
}

// ─── Roles CRUD ───────────────────────────────────────────────────────────────

func (s *RBACService) GetAllRoles() ([]domain.Role, error) {
	return s.store.RoleRepo.GetAll()
}

func (s *RBACService) GetRoleByID(id int64) (*domain.Role, error) {
	role, err := s.store.RoleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return role, nil
}

func (s *RBACService) CreateRole(code, name, description string) (*domain.Role, error) {
	existing, err := s.store.RoleRepo.GetByCode(code)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("role code already exists")
	}
	role := &domain.Role{Code: code, Name: name, Description: description}
	if err := s.store.RoleRepo.Create(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RBACService) UpdateRole(id int64, code, name, description string) (*domain.Role, error) {
	role, err := s.store.RoleRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	role.Code = code
	role.Name = name
	role.Description = description
	if err := s.store.RoleRepo.Update(role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *RBACService) DeleteRole(id int64) error {
	role, err := s.store.RoleRepo.GetByID(id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}
	return s.store.RoleRepo.Delete(id)
}

// ─── Permissions ──────────────────────────────────────────────────────────────

func (s *RBACService) GetAllPermissions() ([]domain.Permission, error) {
	return s.store.PermissionRepo.GetAllPermissions()
}

func (s *RBACService) GetRolePermissions(roleID int64) ([]domain.Permission, error) {
	role, err := s.store.RoleRepo.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}
	return s.store.RoleRepo.GetPermissions(roleID)
}

func (s *RBACService) SetRolePermissions(roleID int64, permissionIDs []int64) error {
	role, err := s.store.RoleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}
	return s.store.RoleRepo.SetPermissions(roleID, permissionIDs)
}

// ─── User role assignment ─────────────────────────────────────────────────────

// AssignRoleToUser schimbă rolul unui user și revocă toate sesiunile active ale
// acestuia (forțând re-login cu noul role_id în JWT).
func (s *RBACService) AssignRoleToUser(userID int64, roleID int64) error {
	role, err := s.store.RoleRepo.GetByID(roleID)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	if err := s.store.UserRepo.UpdateRole(userID, roleID); err != nil {
		return err
	}

	// Revocă toate sesiunile active — forțează re-login cu noul role_id
	oldJTIs, _ := s.refreshRepo.RevokeAllUserSessions(userID)
	for _, jti := range oldJTIs {
		_ = s.blacklist.Add(context.Background(), jti, s.jwtService.AccessTokenTTL())
	}

	return nil
}
