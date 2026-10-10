package repository

import "agri-api/internal/domain"

type RoleRepository interface {
	GetAll() ([]domain.Role, error)
	GetByID(id int64) (*domain.Role, error)
	GetByCode(code string) (*domain.Role, error)
	GetPermissions(roleID int64) ([]domain.Permission, error)
}

type PermissionRepository interface {
	GetAllPermissions() ([]domain.Permission, error)
	GetPermissionByID(id int64) (*domain.Permission, error)
	// HasPermission este apelat de RequirePermission middleware la fiecare request
	HasPermission(roleID int64, permissionName string) (bool, error)
}
