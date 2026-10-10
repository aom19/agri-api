package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"database/sql"
	"errors"
)

var (
	ErrResourceTypeNotFound = errors.New("resource type not found")
	ErrResourceNotFound     = errors.New("resource not found")
)

// ResourceService gestionează resursele și stocul lor. Cantitatea se modifică doar prin mișcări,
// astfel încât suma mișcărilor unei resurse să fie egală cu stocul ei curent.
type ResourceService struct {
	db               *sql.DB
	resourceTypeRepo repository.ResourceTypeRepository
	resourceRepo     repository.ResourceRepository
	movements        repository.StockMovementRepository
}

func NewResourceService(
	db *sql.DB,
	resourceTypeRepo repository.ResourceTypeRepository,
	resourceRepo repository.ResourceRepository,
	movements repository.StockMovementRepository,
) *ResourceService {
	return &ResourceService{
		db:               db,
		resourceTypeRepo: resourceTypeRepo,
		resourceRepo:     resourceRepo,
		movements:        movements,
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

// CreateResource creează resursa cu stocul ei. Cantitatea inițială, dacă există, se înregistrează
// ca ajustare de inventar, în aceeași tranzacție.
func (s *ResourceService) CreateResource(input *domain.Resource, actorID *int64) (*domain.Resource, error) {
	if err := s.validateResourceInput(input); err != nil {
		return nil, err
	}
	if input.Quantity < 0 {
		return nil, errors.New("quantity must be greater than or equal to 0")
	}

	resourceType, err := s.resourceTypeRepo.GetByID(input.ResourceTypeID)
	if err != nil {
		return nil, err
	}
	if resourceType == nil {
		return nil, errors.New("resource type not found")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	res := &domain.Resource{
		Name:            input.Name,
		ResourceTypeID:  input.ResourceTypeID,
		PricePerUnit:    input.PricePerUnit,
		MinimumQuantity: input.MinimumQuantity,
		Notes:           input.Notes,
	}
	if err := s.resourceRepo.Create(tx, res); err != nil {
		return nil, err
	}
	if input.Quantity > 0 {
		lock, err := s.movements.LockStock(tx, res.ID)
		if err != nil {
			return nil, err
		}
		if lock == nil {
			return nil, ErrResourceNotFound
		}
		movement, err := buildMovement(lock, domain.StockMovementAdjustment, input.Quantity, nil, nil, "Stoc inițial", actorID)
		if err != nil {
			return nil, err
		}
		if err := s.movements.ApplyMovement(tx, movement); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return s.resourceRepo.GetByID(res.ID)
}

// UpdateResource schimbă datele resursei și pragul minim. Cantitatea nu se editează direct:
// corecțiile se fac prin mișcări (ajustare de inventar).
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
	existing.MinimumQuantity = input.MinimumQuantity
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
	if input.MinimumQuantity < 0 {
		return errors.New("minimum_quantity must be greater than or equal to 0")
	}
	return nil
}
