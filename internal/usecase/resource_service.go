package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
)

var (
	ErrResourceTypeNotFound = errors.New("resource type not found")
	ErrResourceNotFound     = errors.New("resource not found")
)

type ResourceService struct {
	resourceTypeRepo repository.ResourceTypeRepository
	resourceRepo     repository.ResourceRepository
}

func NewResourceService(
	resourceTypeRepo repository.ResourceTypeRepository,
	resourceRepo repository.ResourceRepository,
) *ResourceService {
	return &ResourceService{
		resourceTypeRepo: resourceTypeRepo,
		resourceRepo:     resourceRepo,
	}
}

func (s *ResourceService) GetResourceTypes() ([]domain.ResourceType, error) {
	return s.resourceTypeRepo.GetAll()
}

func (s *ResourceService) GetResourceTypeByID(id int64) (*domain.ResourceType, error) {
	if id <= 0 {
		return nil, errors.New("invalid resource type id")
	}
	return s.resourceTypeRepo.GetByID(id)
}

func (s *ResourceService) CreateResourceType(input *domain.ResourceType) (*domain.ResourceType, error) {
	if input == nil {
		return nil, errors.New("resource type payload is required")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if !input.Category.IsValid() {
		return nil, errors.New("invalid resource category")
	}
	if input.DefaultUnit == "" {
		return nil, errors.New("default_unit is required")
	}

	rt := &domain.ResourceType{
		Name:        input.Name,
		Category:    input.Category,
		DefaultUnit: input.DefaultUnit,
	}

	if err := s.resourceTypeRepo.Create(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func (s *ResourceService) UpdateResourceType(id int64, input *domain.ResourceType) (*domain.ResourceType, error) {
	if id <= 0 {
		return nil, errors.New("invalid resource type id")
	}
	if input == nil {
		return nil, errors.New("resource type payload is required")
	}
	if input.Name == "" {
		return nil, errors.New("name is required")
	}
	if !input.Category.IsValid() {
		return nil, errors.New("invalid resource category")
	}
	if input.DefaultUnit == "" {
		return nil, errors.New("default_unit is required")
	}

	existing, err := s.resourceTypeRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrResourceTypeNotFound
	}

	existing.Name = input.Name
	existing.Category = input.Category
	existing.DefaultUnit = input.DefaultUnit

	if err := s.resourceTypeRepo.Update(id, existing); err != nil {
		return nil, err
	}

	return s.resourceTypeRepo.GetByID(id)
}

func (s *ResourceService) DeleteResourceType(id int64) error {
	if id <= 0 {
		return errors.New("invalid resource type id")
	}

	existing, err := s.resourceTypeRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrResourceTypeNotFound
	}

	return s.resourceTypeRepo.Delete(id)
}

func (s *ResourceService) GetResources() ([]domain.Resource, error) {
	return s.resourceRepo.GetAll()
}

func (s *ResourceService) GetResourceByID(id int64) (*domain.Resource, error) {
	if id <= 0 {
		return nil, errors.New("invalid resource id")
	}
	return s.resourceRepo.GetByID(id)
}

func (s *ResourceService) CreateResource(input *domain.Resource) (*domain.Resource, error) {
	if err := s.validateResourceInput(input); err != nil {
		return nil, err
	}

	resourceType, err := s.resourceTypeRepo.GetByID(input.ResourceTypeID)
	if err != nil {
		return nil, err
	}
	if resourceType == nil {
		return nil, errors.New("resource type not found")
	}

	res := &domain.Resource{
		Name:           input.Name,
		ResourceTypeID: input.ResourceTypeID,
		PricePerUnit:   input.PricePerUnit,
		Notes:          input.Notes,
	}

	if err := s.resourceRepo.Create(res); err != nil {
		return nil, err
	}

	return s.resourceRepo.GetByID(res.ID)
}

func (s *ResourceService) UpdateResource(id int64, input *domain.Resource) (*domain.Resource, error) {
	if id <= 0 {
		return nil, errors.New("invalid resource id")
	}
	if err := s.validateResourceInput(input); err != nil {
		return nil, err
	}

	existing, err := s.resourceRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrResourceNotFound
	}

	resourceType, err := s.resourceTypeRepo.GetByID(input.ResourceTypeID)
	if err != nil {
		return nil, err
	}
	if resourceType == nil {
		return nil, errors.New("resource type not found")
	}

	existing.Name = input.Name
	existing.ResourceTypeID = input.ResourceTypeID
	existing.PricePerUnit = input.PricePerUnit
	existing.Notes = input.Notes

	if err := s.resourceRepo.Update(id, existing); err != nil {
		return nil, err
	}

	return s.resourceRepo.GetByID(id)
}

func (s *ResourceService) DeleteResource(id int64) error {
	if id <= 0 {
		return errors.New("invalid resource id")
	}

	existing, err := s.resourceRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrResourceNotFound
	}

	return s.resourceRepo.Delete(id)
}

func (s *ResourceService) validateResourceInput(input *domain.Resource) error {
	if input == nil {
		return errors.New("resource payload is required")
	}
	if input.Name == "" {
		return errors.New("name is required")
	}
	if input.ResourceTypeID <= 0 {
		return errors.New("resource_type_id is required")
	}
	if input.PricePerUnit < 0 {
		return errors.New("price_per_unit must be greater than or equal to 0")
	}
	return nil
}
