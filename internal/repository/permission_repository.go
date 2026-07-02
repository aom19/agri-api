package repository

import "agri-api/internal/domain"

type PermissionRepository interface {
	GetAll() ([]domain.Permission, error)
	GetByID(id string) (*domain.Permission, error)
}