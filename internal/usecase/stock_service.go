package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"database/sql"
	"errors"
)

var (
	ErrStockNotFound         = errors.New("stock not found")
	ErrStockAlreadyExists    = errors.New("stock already exists for this resource")
	ErrStockResourceNotFound = errors.New("resource not found")
)

// StockService gestionează fișele de stoc. Cantitatea se modifică doar prin mișcări,
// astfel încât suma mișcărilor unei resurse să fie egală cu stocul ei curent.
type StockService struct {
	db           *sql.DB
	stockRepo    repository.StockRepository
	resourceRepo repository.ResourceRepository
	movements    repository.StockMovementRepository
}

func NewStockService(
	db *sql.DB,
	stockRepo repository.StockRepository,
	resourceRepo repository.ResourceRepository,
	movements repository.StockMovementRepository,
) *StockService {
	return &StockService{db: db, stockRepo: stockRepo, resourceRepo: resourceRepo, movements: movements}
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

// CreateStock creează fișa de stoc. Cantitatea inițială, dacă există, se înregistrează
// ca ajustare de inventar, în aceeași tranzacție.
func (s *StockService) CreateStock(input *domain.Stock, actorID *int64) (*domain.Stock, error) {
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

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	stock := &domain.Stock{
		ResourceID:      input.ResourceID,
		MinimumQuantity: input.MinimumQuantity,
	}
	if err := s.stockRepo.Create(tx, stock); err != nil {
		return nil, err
	}
	if input.Quantity > 0 {
		lock, err := s.movements.LockStockByID(tx, stock.ID)
		if err != nil {
			return nil, err
		}
		if lock == nil {
			return nil, ErrStockNotFound
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
	return s.stockRepo.GetByID(stock.ID)
}

// UpdateMinimum schimbă pragul minim al stocului. Cantitatea nu se editează direct:
// corecțiile se fac prin mișcări (ajustare de inventar).
func (s *StockService) UpdateMinimum(id int64, minimum float64) (*domain.Stock, error) {
	if id <= 0 {
		return nil, errors.New("invalid stock id")
	}
	if minimum < 0 {
		return nil, errors.New("minimum_quantity must be greater than or equal to 0")
	}

	existing, err := s.stockRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrStockNotFound
	}

	if err := s.stockRepo.UpdateMinimum(id, minimum); err != nil {
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
