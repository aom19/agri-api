package repository

import "agri-api/internal/domain"

type RoleRepository interface {
	GetAll() ([]domain.Role, error)
	GetByID(id int64) (*domain.Role, error)
	GetByName(name string) (*domain.Role, error)
	Create(role *domain.Role) error
	Update(role *domain.Role) error
	Delete(id int64) error
	GetPermissions(roleID int64) ([]domain.Permission, error)
	SetPermissions(roleID int64, permissionIDs []int64) error
}

type PermissionRepository interface {
	GetAllPermissions() ([]domain.Permission, error)
	GetPermissionByID(id int64) (*domain.Permission, error)
	// HasPermission este apelat de RequirePermission middleware la fiecare request
	HasPermission(roleID int64, permissionName string) (bool, error)
}
