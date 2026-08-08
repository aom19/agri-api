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
	AssignedUserID  int64
}

type FieldOperationRepository interface {
	GetAll(filter FieldOperationFilter) ([]dto.FieldOperationResponse, error)
	GetByID(id int64) (*dto.FieldOperationResponse, error)
	GetByIDForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error)
	Create(op *domain.FieldOperation) error
	Update(id int64, op *domain.FieldOperation) error
	UpdateChecklist(id int64, checklist domain.FieldOperationChecklist) error
	UpdateStatus(id int64, status domain.FieldOperationStatus) error
	Delete(id int64) error
}
