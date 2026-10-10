package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
)

var (
	ErrOperationTemplateNotFound = errors.New("operation template not found")
	ErrTemplateNameRequired      = errors.New("template name is required")
	ErrTemplateUnitRequired      = errors.New("template unit is required")
	ErrInvalidOperationType      = errors.New("invalid operation_type")
)

type OperationService struct {
	templateRepo repository.OperationTemplateRepository
}

func NewOperationService(templateRepo repository.OperationTemplateRepository) *OperationService {
	return &OperationService{templateRepo: templateRepo}
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

func (s *OperationService) CreateTemplate(input *domain.OperationTemplate) (*domain.OperationTemplate, error) {
	if input.Name == "" {
		return nil, ErrTemplateNameRequired
	}
	if input.Unit == "" {
		return nil, ErrTemplateUnitRequired
	}
	if !input.OperationType.IsValid() {
		return nil, ErrInvalidOperationType
	}

	t := &domain.OperationTemplate{
		OperationType: input.OperationType,
		Name:          input.Name,
		Description:   input.Description,
		Unit:          input.Unit,
		CropID:        input.CropID,
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
	if !input.OperationType.IsValid() {
		return nil, ErrInvalidOperationType
	}

	t := &domain.OperationTemplate{
		OperationType: input.OperationType,
		Name:          input.Name,
		Description:   input.Description,
		Unit:          input.Unit,
		CropID:        input.CropID,
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
