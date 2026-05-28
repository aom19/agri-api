package repository

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"database/sql"
)

type AssigmentRepository interface {
	Create(tx *sql.Tx, assigment *domain.Assigment) error
	GetAll() ([]domain.Assigment, error)
	GetAllWithPagination(query dto.PaginationQuery) (*dto.PaginatedAssignmentsResponse, error)
	GetByID(id int64) (*domain.Assigment, error)
	Update(id int64, assigment *domain.Assigment) error
	Delete(id int64) error
	GetActiveByMachine(machineID int64) (*domain.Assigment, error)
	GetActiveByOperator(operatorID int64) (*domain.Assigment, error)
	IsAssigmentActive(id int64) (bool, error)
}
