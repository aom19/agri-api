package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"encoding/json"
	"errors"
	"math"
)


type PermissionService struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionService(permissionRepo repository.PermissionRepository) *PermissionService {
	return &PermissionService{permissionRepo: permissionRepo}
}

func (s *PermissionService) GetPermissions() ([]domain.Permission, error) {
	return s.permissionRepo.GetAll()
}

func (s *PermissionService) GetPermissionByID(id string) (*domain.Permission, error) {
	if id == "" {
		return nil, errors.New("id-ul permisiunii este necesar")
	}
	return s.permissionRepo.GetByID(id)
}