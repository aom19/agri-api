package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type MachineRepository interface {
	Create(machine *domain.Machine) error
	GetAll() ([]domain.Machine, error)
	GetByID(id int64) (*domain.Machine, error)
	Update(id int64, machine *domain.Machine) error
	Delete(id int64) error
	UpdateStatus(tx *sql.Tx, id int64, status domain.MachineStatus) error
	DB() *sql.DB
}
