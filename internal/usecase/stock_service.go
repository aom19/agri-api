package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"errors"
)

var (
	ErrStockNotFound         = errors.New("stock not found")
	ErrStockAlreadyExists    = errors.New("stock already exists for this resource")
	ErrStockResourceNotFound = errors.New("resource not found")
)

type StockService struct {
	stockRepo    repository.StockRepository
	resourceRepo repository.ResourceRepository
}

func NewStockService(
	stockRepo repository.StockRepository,
	resourceRepo repository.ResourceRepository,
) *StockService {
	return &StockService{stockRepo: stockRepo, resourceRepo: resourceRepo}
}

func (s *StockService) GetStocks() ([]domain.Stock, error) {
	return s.stockRepo.GetAll()
}

func (s *StockService) GetStockByID(id int64) (*domain.Stock, error) {
	if id <= 0 {
		return nil, errors.New("invalid stock id")
	}
	return s.stockRepo.GetByID(id)
}

func (s *StockService) CreateStock(input *domain.Stock) (*domain.Stock, error) {
	if err := s.validateStockInput(input); err != nil {
		return nil, err
	}
	if err := s.ensureResourceExists(input.ResourceID); err != nil {
		return nil, err
	}

	existing, err := s.stockRepo.GetByResourceID(input.ResourceID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrStockAlreadyExists
	}

	stock := &domain.Stock{
		ResourceID:      input.ResourceID,
		Quantity:        input.Quantity,
		MinimumQuantity: input.MinimumQuantity,
	}
	if err := s.stockRepo.Create(stock); err != nil {
		return nil, err
	}
	return s.stockRepo.GetByID(stock.ID)
}

func (s *StockService) UpdateStock(id int64, input *domain.Stock) (*domain.Stock, error) {
	if id <= 0 {
		return nil, errors.New("invalid stock id")
	}
	if err := s.validateStockInput(input); err != nil {
		return nil, err
	}

	existing, err := s.stockRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrStockNotFound
	}
	if err := s.ensureResourceExists(input.ResourceID); err != nil {
		return nil, err
	}

	if existing.ResourceID != input.ResourceID {
		stockForResource, err := s.stockRepo.GetByResourceID(input.ResourceID)
		if err != nil {
			return nil, err
		}
		if stockForResource != nil {
			return nil, ErrStockAlreadyExists
		}
	}

	existing.ResourceID = input.ResourceID
	existing.Quantity = input.Quantity
	existing.MinimumQuantity = input.MinimumQuantity
	if err := s.stockRepo.Update(id, existing); err != nil {
		return nil, err
	}
	return s.stockRepo.GetByID(id)
}

func (s *StockService) DeleteStock(id int64) error {
	if id <= 0 {
		return errors.New("invalid stock id")
	}

	existing, err := s.stockRepo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrStockNotFound
	}

	return s.stockRepo.Delete(id)
}

func (s *StockService) validateStockInput(input *domain.Stock) error {
	if input == nil {
		return errors.New("stock payload is required")
	}
	if input.ResourceID <= 0 {
		return errors.New("resource_id is required")
	}
	if input.Quantity < 0 {
		return errors.New("quantity must be greater than or equal to 0")
	}
	if input.MinimumQuantity < 0 {
		return errors.New("minimum_quantity must be greater than or equal to 0")
	}
	return nil
}

func (s *StockService) ensureResourceExists(resourceID int64) error {
	resource, err := s.resourceRepo.GetByID(resourceID)
	if err != nil {
		return err
	}
	if resource == nil {
		return ErrStockResourceNotFound
	}
	return nil
}
