package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"errors"
	"strings"
)

var (
	ErrFieldOperationNotFound      = errors.New("field operation not found")
	ErrFieldOperationInvalidStatus = errors.New("invalid field operation status")
	ErrFieldOperationFieldRequired = errors.New("field_id is required")
	ErrFieldOperationTypeRequired  = errors.New("operation_type_id is required")
	ErrFieldOperationBadDates      = errors.New("planned_end_at must be after or equal to planned_start_at")
)

type FieldOperationService struct {
	repo repository.FieldOperationRepository
}

func NewFieldOperationService(repo repository.FieldOperationRepository) *FieldOperationService {
	return &FieldOperationService{repo: repo}
}

func (s *FieldOperationService) GetAll(filter repository.FieldOperationFilter) ([]dto.FieldOperationResponse, error) {
	return s.repo.GetAll(filter)
}

func (s *FieldOperationService) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, ErrFieldOperationNotFound
	}
	return item, nil
}

func (s *FieldOperationService) Create(input *domain.FieldOperation) (*dto.FieldOperationResponse, error) {
	if err := s.validate(input); err != nil {
		return nil, err
	}
	if input.Status == "" {
		input.Status = domain.FieldOperationStatusPlanned
	}
	input.Notes = strings.TrimSpace(input.Notes)

	if err := s.repo.Create(input); err != nil {
		return nil, err
	}
	return s.repo.GetByID(input.ID)
}

func (s *FieldOperationService) Update(id int64, input *domain.FieldOperation) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := s.validate(input); err != nil {
		return nil, err
	}
	if input.Status == "" {
		input.Status = domain.FieldOperationStatus(existing.Status)
	}
	input.Notes = strings.TrimSpace(input.Notes)

	if err := s.repo.Update(id, input); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *FieldOperationService) Delete(id int64) error {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrFieldOperationNotFound
	}
	return s.repo.Delete(id)
}

func (s *FieldOperationService) validate(input *domain.FieldOperation) error {
	if input == nil {
		return errors.New("payload is required")
	}
	if strings.TrimSpace(input.FieldID) == "" {
		return ErrFieldOperationFieldRequired
	}
	if input.OperationTypeID <= 0 {
		return ErrFieldOperationTypeRequired
	}
	if input.Status != "" && !isValidFieldOperationStatus(input.Status) {
		return ErrFieldOperationInvalidStatus
	}
	if input.AreaPlannedHa != nil && *input.AreaPlannedHa < 0 {
		return errors.New("area_planned_ha must be greater than or equal to 0")
	}
	if input.PlannedStartAt != nil && input.PlannedEndAt != nil &&
		input.PlannedEndAt.Before(*input.PlannedStartAt) {
		return ErrFieldOperationBadDates
	}
	return nil
}

func isValidFieldOperationStatus(status domain.FieldOperationStatus) bool {
	switch status {
	case domain.FieldOperationStatusPlanned,
		domain.FieldOperationStatusInProgress,
		domain.FieldOperationStatusCompleted,
		domain.FieldOperationStatusCanceled:
		return true
	default:
		return false
	}
}
