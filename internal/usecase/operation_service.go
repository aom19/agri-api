package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
)

var (
	ErrOperationTypeNotFound     = errors.New("operation type not found")
	ErrOperationTemplateNotFound = errors.New("operation template not found")
	ErrOperationTypeCodeRequired = errors.New("code is required")
	ErrOperationTypeNameRequired = errors.New("name is required")
	ErrTemplateNameRequired      = errors.New("template name is required")
	ErrTemplateUnitRequired      = errors.New("template unit is required")
	ErrInvalidOperationTypeID    = errors.New("invalid operation_type_id")
)

type OperationService struct {
	typeRepo     repository.OperationTypeRepository
	templateRepo repository.OperationTemplateRepository
}

func NewOperationService(
	typeRepo repository.OperationTypeRepository,
	templateRepo repository.OperationTemplateRepository,
) *OperationService {
	return &OperationService{
		typeRepo:     typeRepo,
		templateRepo: templateRepo,
	}
}

// ─── OperationType CRUD ──────────────────────────────────────────────────────

func (s *OperationService) GetAllTypes() ([]domain.OperationType, error) {
	return s.typeRepo.GetAll()
}

func (s *OperationService) GetTypeByID(id int64) (*domain.OperationType, error) {
	ot, err := s.typeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if ot == nil {
		return nil, ErrOperationTypeNotFound
	}
	return ot, nil
}

func (s *OperationService) CreateType(input *domain.OperationType) (*domain.OperationType, error) {
	if input.Code == "" {
		return nil, ErrOperationTypeCodeRequired
	}
	if input.Name == "" {
		return nil, ErrOperationTypeNameRequired
	}

	ot := &domain.OperationType{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
	}
	if err := s.typeRepo.Create(ot); err != nil {
		return nil, err
	}
	return ot, nil
}

func (s *OperationService) UpdateType(id int64, input *domain.OperationType) (*domain.OperationType, error) {
	existing, err := s.typeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrOperationTypeNotFound
	}

	if input.Code == "" {
		return nil, ErrOperationTypeCodeRequired
	}
	if input.Name == "" {
		return nil, ErrOperationTypeNameRequired
	}

	ot := &domain.OperationType{
		Code:        input.Code,
		Name:        input.Name,
		Description: input.Description,
	}
	if err := s.typeRepo.Update(id, ot); err != nil {
		return nil, err
	}
	ot.ID = id
	return ot, nil
}

func (s *OperationService) DeleteType(id int64) error {
	existing, err := s.typeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrOperationTypeNotFound
	}
	return s.typeRepo.Delete(id)
}

// ─── OperationTemplate CRUD ──────────────────────────────────────────────────

func (s *OperationService) GetAllTemplates() ([]domain.OperationTemplate, error) {
	return s.templateRepo.GetAll()
}

func (s *OperationService) GetTemplateByID(id int64) (*domain.OperationTemplate, error) {
	t, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrOperationTemplateNotFound
	}
	return t, nil
}

func (s *OperationService) GetTemplatesByOperationType(operationTypeID int64) ([]domain.OperationTemplate, error) {
	return s.templateRepo.GetByOperationType(operationTypeID)
}

func (s *OperationService) CreateTemplate(input *domain.OperationTemplate) (*domain.OperationTemplate, error) {
	if input.Name == "" {
		return nil, ErrTemplateNameRequired
	}
	if input.Unit == "" {
		return nil, ErrTemplateUnitRequired
	}
	if input.OperationTypeID <= 0 {
		return nil, ErrInvalidOperationTypeID
	}

	// Validate operation type exists
	ot, err := s.typeRepo.GetByID(input.OperationTypeID)
	if err != nil {
		return nil, err
	}
	if ot == nil {
		return nil, ErrOperationTypeNotFound
	}

	t := &domain.OperationTemplate{
		OperationTypeID: input.OperationTypeID,
		Name:            input.Name,
		Description:     input.Description,
		Unit:            input.Unit,
	}
	if err := s.templateRepo.Create(t); err != nil {
		return nil, err
	}

	// Set related data if provided
	if len(input.Resources) > 0 {
		if err := s.templateRepo.SetResources(t.ID, input.Resources); err != nil {
			return nil, err
		}
	}
	if len(input.MachineTypes) > 0 {
		if err := s.templateRepo.SetMachineTypes(t.ID, input.MachineTypes); err != nil {
			return nil, err
		}
	}
	if len(input.ImplementTypes) > 0 {
		if err := s.templateRepo.SetImplementTypes(t.ID, input.ImplementTypes); err != nil {
			return nil, err
		}
	}

	return s.templateRepo.GetByID(t.ID)
}

func (s *OperationService) UpdateTemplate(id int64, input *domain.OperationTemplate) (*domain.OperationTemplate, error) {
	existing, err := s.templateRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrOperationTemplateNotFound
	}

	if input.Name == "" {
		return nil, ErrTemplateNameRequired
	}
	if input.Unit == "" {
		return nil, ErrTemplateUnitRequired
	}
	if input.OperationTypeID <= 0 {
		return nil, ErrInvalidOperationTypeID
	}

	t := &domain.OperationTemplate{
		OperationTypeID: input.OperationTypeID,
		Name:            input.Name,
		Description:     input.Description,
		Unit:            input.Unit,
	}
	if err := s.templateRepo.Update(id, t); err != nil {
		return nil, err
	}

	// Update related data
	if input.Resources != nil {
		if err := s.templateRepo.SetResources(id, input.Resources); err != nil {
			return nil, err
		}
	}
	if input.MachineTypes != nil {
		if err := s.templateRepo.SetMachineTypes(id, input.MachineTypes); err != nil {
			return nil, err
		}
	}
	if input.ImplementTypes != nil {
		if err := s.templateRepo.SetImplementTypes(id, input.ImplementTypes); err != nil {
			return nil, err
		}
	}

	return s.templateRepo.GetByID(id)
}

func (s *OperationService) DeleteTemplate(id int64) error {
	existing, err := s.templateRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrOperationTemplateNotFound
	}
	return s.templateRepo.Delete(id)
}
