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
	Create(s *domain.Stock) error
	Update(id int64, s *domain.Stock) error
	Delete(id int64) error
	DecrementQuantity(tx *sql.Tx, resourceID int64, qty float64) error
}
