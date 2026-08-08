package repository

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
)

type FieldOperationFilter struct {
	Status          string
	FieldID         string
	OperationTypeID string
	MachineID       string
	OperatorID      string
}

type FieldOperationRepository interface {
	GetAll(filter FieldOperationFilter) ([]dto.FieldOperationResponse, error)
	GetByID(id int64) (*dto.FieldOperationResponse, error)
	Create(op *domain.FieldOperation) error
	Update(id int64, op *domain.FieldOperation) error
	Delete(id int64) error
}
