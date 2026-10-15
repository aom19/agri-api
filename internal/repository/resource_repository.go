package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ResourceTypeRepository interface {
	GetAll() ([]domain.ResourceType, error)
	GetByID(id int64) (*domain.ResourceType, error)
	Create(rt *domain.ResourceType) error
	Update(id int64, rt *domain.ResourceType) error
	Delete(id int64) error
}

type ResourceRepository interface {
	GetAll() ([]domain.Resource, error)
	// List aplică filtrul în SQL; CountByCategory numără toate resursele pe categorie.
	List(filter domain.ResourceFilter) ([]domain.Resource, error)
	CountByCategory() (*domain.CategoryCounts, error)
	GetByID(id int64) (*domain.Resource, error)
	// Create inserează resursa cu stocul 0; cantitatea inițială se adaugă printr-o mișcare.
	Create(tx *sql.Tx, r *domain.Resource) error
	// Update nu schimbă cantitatea: ea se modifică doar prin mișcări.
	Update(id int64, r *domain.Resource) error
	Delete(id int64) error
}
