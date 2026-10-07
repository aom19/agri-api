package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"errors"
	"slices"
	"strings"
	"time"
)

var (
	ErrFieldOperationNotFound              = errors.New("operațiunea pe teren nu a fost găsită")
	ErrFieldOperationInvalidStatus         = errors.New("statusul operațiunii nu este valid")
	ErrFieldOperationFieldRequired         = errors.New("terenul este obligatoriu")
	ErrFieldOperationTypeRequired          = errors.New("tipul operațiunii este obligatoriu")
	ErrFieldOperationBadDates              = errors.New("sfârșitul planificat nu poate fi înaintea începutului planificat")
	ErrFieldOperationResourcesUnavailable  = errors.New("mașina și echipamentul trebuie să fie active înainte de pornire")
	ErrFieldOperationCannotStart           = errors.New("lucrarea nu poate fi pornită din statusul curent")
	ErrFieldOperationMachineIncompatible   = errors.New("tipul mașinii nu este acceptat de template-ul operațiunii")
	ErrFieldOperationImplementIncompatible = errors.New("tipul echipamentului nu este acceptat de template-ul operațiunii")
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
	if err := s.repo.MarkStarted(id, time.Now()); err != nil {
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
	if err := s.repo.MarkStarted(id, time.Now()); err != nil {
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
		return errors.New("datele operațiunii lipsesc")
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
		return errors.New("suprafața planificată nu poate fi negativă")
	}
	if input.PlannedStartAt != nil && input.PlannedEndAt != nil &&
		input.PlannedEndAt.Before(*input.PlannedStartAt) {
		return ErrFieldOperationBadDates
	}
	return s.validateAssetCompatibility(input)
}

// validateAssetCompatibility verifică mașina și echipamentul față de tipurile acceptate de
// template. Un template fără tipuri, sau o operațiune fără template, acceptă orice.
func (s *FieldOperationService) validateAssetCompatibility(input *domain.FieldOperation) error {
	if input.OperationTemplateID == nil || (input.MachineID == nil && input.ImplementID == nil) {
		return nil
	}
	c, err := s.repo.GetAssetCompatibility(*input.OperationTemplateID, input.MachineID, input.ImplementID)
	if err != nil {
		return err
	}
	if !typeAllowed(c.TemplateMachineTypes, c.MachineType) {
		return ErrFieldOperationMachineIncompatible
	}
	if !typeAllowed(c.TemplateImplementTypes, c.ImplementType) {
		return ErrFieldOperationImplementIncompatible
	}
	return nil
}

func typeAllowed(allowed []string, assetType string) bool {
	return assetType == "" || len(allowed) == 0 || slices.Contains(allowed, assetType)
}

func validateStartReadiness(operation *dto.FieldOperationResponse) error {
	if operation.Status == string(domain.FieldOperationStatusInProgress) {
		return nil
	}
	if operation.Status != string(domain.FieldOperationStatusPlanned) {
		return ErrFieldOperationCannotStart
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
