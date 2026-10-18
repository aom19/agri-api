package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type MachineRepository interface {
	Create(machine *domain.Machine) error
	GetAll() ([]domain.Machine, error)
	// List: căutare (nume, cod, tip, marcă, model, înmatriculare) și paginare opțională.
	List(filter domain.ListFilter, page *domain.PageRequest) ([]domain.Machine, int, error)
	GetByID(id int64) (*domain.Machine, error)
	Update(id int64, machine *domain.Machine) error
	Delete(id int64) error
	UpdateStatus(tx *sql.Tx, id int64, status domain.AssetStatus) error
	DB() *sql.DB
}
