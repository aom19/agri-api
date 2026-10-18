package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ImplementRepository interface {
	Create(implement *domain.Implement) error
	GetAll() ([]domain.Implement, error)
	// List: căutare (nume, cod, tip, marcă, model) și paginare opțională.
	List(filter domain.ListFilter, page *domain.PageRequest) ([]domain.Implement, int, error)
	GetByID(id int64) (*domain.Implement, error)
	Update(id int64, implement *domain.Implement) error
	Delete(id int64) error
	Activate(id int64) error
	Deactivate(id int64) error
	DB() *sql.DB
}
