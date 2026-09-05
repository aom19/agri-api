package repository

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"time"
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
	// GetOverdueInProgress returnează operațiunile în lucru al căror sfârșit planificat
	// este înainte de `now` și pentru care nu s-a trimis încă notificarea de depășire.
	GetOverdueInProgress(now time.Time) ([]dto.OverdueFieldOperation, error)
	// MarkOverdueNotified marchează că notificarea de depășire a fost trimisă.
	MarkOverdueNotified(id int64, at time.Time) error
}
