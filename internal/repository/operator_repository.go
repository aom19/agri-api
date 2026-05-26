package repository

import (
	"agri-api/internal/domain"
	"database/sql"
)

type OperatorRepository interface {
	Create(operator *domain.Operator) error
	GetAll() ([]domain.Operator, error)
	GetByID(id int64) (*domain.Operator, error)
	Update(id int64, operator *domain.Operator) error
	Delete(id int64) error
	UpdateStatus(tx *sql.Tx, id int64, status domain.OperatorStatus) error
}
