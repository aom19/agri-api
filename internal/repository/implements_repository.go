package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ImplementRepository interface {
	Create(implement *domain.Implement) error
	GetAll() ([]domain.Implement, error)
	GetByID(id int64) (*domain.Implement, error)
	Update(id int64, implement *domain.Implement) error
	Delete(id int64) error
	Activate(id int64) error
	Deactivate(id int64) error
	DB() *sql.DB
}
