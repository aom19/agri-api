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
	GetByID(id int64) (*domain.Resource, error)
	Create(r *domain.Resource) error
	Update(id int64, r *domain.Resource) error
	Delete(id int64) error
}

type StockRepository interface {
	GetAll() ([]domain.Stock, error)
	GetByID(id int64) (*domain.Stock, error)
	GetByResourceID(resourceID int64) (*domain.Stock, error)
	// Create inserează stocul cu cantitatea 0; cantitatea inițială se adaugă printr-o mișcare.
	Create(tx *sql.Tx, s *domain.Stock) error
	// UpdateMinimum schimbă doar pragul minim. Cantitatea se modifică doar prin mișcări.
	UpdateMinimum(id int64, minimum float64) error
	Delete(id int64) error
}
