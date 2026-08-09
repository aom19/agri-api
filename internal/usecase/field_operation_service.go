package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"errors"
	"strings"
)

var (
	ErrFieldOperationNotFound             = errors.New("field operation not found")
	ErrFieldOperationInvalidStatus        = errors.New("invalid field operation status")
	ErrFieldOperationFieldRequired        = errors.New("field_id is required")
	ErrFieldOperationTypeRequired         = errors.New("operation_type_id is required")
	ErrFieldOperationBadDates             = errors.New("planned_end_at must be after or equal to planned_start_at")
	ErrFieldOperationChecklistIncomplete  = errors.New("checklist must be completed before starting")
	ErrFieldOperationResourcesUnavailable = errors.New("assigned machine and implement must be active before starting")
	ErrFieldOperationCannotStart          = errors.New("field operation cannot be started from current status")
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

func (s *FieldOperationService) GetByIDForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error) {
	item, err := s.repo.GetByIDForAssignedUser(id, userID)
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

func (s *FieldOperationService) UpdateChecklist(id int64, checklist domain.FieldOperationChecklist) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := s.repo.UpdateChecklist(id, checklist); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *FieldOperationService) UpdateChecklistForAssignedUser(id int64, userID int64, checklist domain.FieldOperationChecklist) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := s.repo.UpdateChecklist(id, checklist); err != nil {
		return nil, err
	}
	return s.repo.GetByIDForAssignedUser(id, userID)
}

func (s *FieldOperationService) Start(id int64) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := validateStartReadiness(existing); err != nil {
		return nil, err
	}
	if existing.Status == string(domain.FieldOperationStatusInProgress) {
		return existing, nil
	}
	if err := s.repo.UpdateStatus(id, domain.FieldOperationStatusInProgress); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}

func (s *FieldOperationService) StartForAssignedUser(id int64, userID int64) (*dto.FieldOperationResponse, error) {
	existing, err := s.repo.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if err := validateStartReadiness(existing); err != nil {
		return nil, err
	}
	if existing.Status == string(domain.FieldOperationStatusInProgress) {
		return existing, nil
	}
	if err := s.repo.UpdateStatus(id, domain.FieldOperationStatusInProgress); err != nil {
		return nil, err
	}
	return s.repo.GetByIDForAssignedUser(id, userID)
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

func validateStartReadiness(operation *dto.FieldOperationResponse) error {
	if operation.Status == string(domain.FieldOperationStatusInProgress) {
		return nil
	}
	if operation.Status != string(domain.FieldOperationStatusPlanned) {
		return ErrFieldOperationCannotStart
	}
	if !operation.Checklist.MachineStatus || !operation.Checklist.ImplementStatus ||
		!operation.Checklist.FieldArea || !operation.Checklist.NotesConfirmed {
		return ErrFieldOperationChecklistIncomplete
	}
	if operation.MachineID != nil && !isActiveAssetStatus(operation.MachineStatus) {
		return ErrFieldOperationResourcesUnavailable
	}
	if operation.ImplementID != nil && !isActiveAssetStatus(operation.ImplementStatus) {
		return ErrFieldOperationResourcesUnavailable
	}
	return nil
}

func isActiveAssetStatus(status *string) bool {
	return status != nil && *status == string(domain.AssetStatusActive)
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
