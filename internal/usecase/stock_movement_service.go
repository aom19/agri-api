package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

var ErrInvalidStockMovement = errors.New("mișcare de stoc invalidă")

type StockMovementService struct {
	db   *sql.DB
	repo repository.StockMovementRepository
}

func NewStockMovementService(db *sql.DB, repo repository.StockMovementRepository) *StockMovementService {
	return &StockMovementService{db: db, repo: repo}
}

func (service *StockMovementService) List(filter domain.StockMovementFilter) ([]domain.StockMovement, error) {
	return service.repo.List(filter)
}

// StockMovementResult conține mișcarea înregistrată și starea stocului față de pragul minim.
type StockMovementResult struct {
	Movement     *domain.StockMovement
	Minimum      float64
	BelowMinimum bool
}

// Create înregistrează o mișcare manuală (recepție, consum sau ajustare de inventar).
func (service *StockMovementService) Create(input domain.StockMovementInput) (*StockMovementResult, error) {
	if input.StockID <= 0 {
		return nil, fmt.Errorf("%w: stocul este obligatoriu", ErrInvalidStockMovement)
	}
	if input.MovementType != domain.StockMovementAdjustment && input.Quantity <= 0 {
		return nil, fmt.Errorf("%w: cantitatea trebuie să fie pozitivă", ErrInvalidStockMovement)
	}
	if input.MovementType == domain.StockMovementAdjustment && input.Quantity < 0 {
		return nil, fmt.Errorf("%w: nivelul ajustat nu poate fi negativ", ErrInvalidStockMovement)
	}
	if input.MovementType != domain.StockMovementIn &&
		input.MovementType != domain.StockMovementOut &&
		input.MovementType != domain.StockMovementAdjustment {
		return nil, fmt.Errorf("%w: tipul trebuie să fie in, out sau adjustment", ErrInvalidStockMovement)
	}

	tx, err := service.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	lock, err := service.repo.LockStockByID(tx, input.StockID)
	if err != nil {
		return nil, err
	}
	if lock == nil {
		return nil, ErrStockNotFound
	}

	movement, err := buildMovement(lock, input.MovementType, input.Quantity, input.UnitCost, input.FieldOperationID, input.Notes, input.ActorID)
	if err != nil {
		return nil, err
	}
	if err := service.repo.ApplyMovement(tx, movement); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &StockMovementResult{
		Movement:     movement,
		Minimum:      lock.Minimum,
		BelowMinimum: lock.Minimum > 0 && movement.ResultingQuantity <= lock.Minimum,
	}, nil
}

// buildMovement calculează variația și cantitatea rezultată pentru un stoc blocat.
func buildMovement(
	lock *repository.StockLock,
	movementType domain.StockMovementType,
	quantity float64,
	unitCost *float64,
	fieldOperationID *int64,
	notes string,
	actorID *int64,
) (*domain.StockMovement, error) {
	var delta float64
	switch movementType {
	case domain.StockMovementIn:
		delta = quantity
	case domain.StockMovementOut:
		delta = -quantity
	case domain.StockMovementAdjustment:
		delta = quantity - lock.Quantity
	default:
		return nil, ErrInvalidStockMovement
	}

	resulting := lock.Quantity + delta
	if resulting < 0 {
		return nil, fmt.Errorf("%w: stoc insuficient (disponibil %.3f, necesar %.3f)", ErrInvalidStockMovement, lock.Quantity, quantity)
	}

	cost := lock.PriceUnit
	if unitCost != nil && *unitCost >= 0 {
		cost = *unitCost
	}
	total := math.Abs(delta) * cost

	return &domain.StockMovement{
		StockID:           lock.StockID,
		ResourceID:        lock.ResourceID,
		FieldOperationID:  fieldOperationID,
		MovementType:      movementType,
		QuantityDelta:     delta,
		ResultingQuantity: resulting,
		UnitCost:          &cost,
		TotalCost:         &total,
		Notes:             notes,
		ActorID:           actorID,
	}, nil
}
