package usecase

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
)

var ErrFieldOperationNotCompletable = errors.New("operațiunea nu poate fi finalizată")

// FieldOperationCompletionService finalizează o operațiune pe teren înregistrând datele
// reale (timp, suprafață, combustibil, ore mașină) și consumul de resurse din stoc,
// totul într-o singură tranzacție.
type FieldOperationCompletionService struct {
	db        *sql.DB
	ops       repository.FieldOperationRepository
	movements repository.StockMovementRepository
}

func NewFieldOperationCompletionService(
	db *sql.DB,
	ops repository.FieldOperationRepository,
	movements repository.StockMovementRepository,
) *FieldOperationCompletionService {
	return &FieldOperationCompletionService{db: db, ops: ops, movements: movements}
}

// CompletionResult conține operațiunea actualizată și mișcările de stoc generate.
type CompletionResult struct {
	Operation *dto.FieldOperationResponse
	Movements []domain.StockMovement
	// Stocurile care au ajuns sub pragul minim după consum.
	LowStocks []repository.StockLock
}

func (service *FieldOperationCompletionService) Complete(id int64, completion domain.FieldOperationCompletion, actorID *int64) (*CompletionResult, error) {
	existing, err := service.ops.GetByID(id)
	if err != nil {
		return nil, err
	}
	return service.complete(existing, completion, actorID, func() (*dto.FieldOperationResponse, error) {
		return service.ops.GetByID(id)
	})
}

func (service *FieldOperationCompletionService) CompleteForAssignedUser(id int64, userID int64, completion domain.FieldOperationCompletion, actorID *int64) (*CompletionResult, error) {
	existing, err := service.ops.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	return service.complete(existing, completion, actorID, func() (*dto.FieldOperationResponse, error) {
		return service.ops.GetByIDForAssignedUser(id, userID)
	})
}

func (service *FieldOperationCompletionService) complete(
	existing *dto.FieldOperationResponse,
	completion domain.FieldOperationCompletion,
	actorID *int64,
	reload func() (*dto.FieldOperationResponse, error),
) (*CompletionResult, error) {
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if existing.Status == string(domain.FieldOperationStatusCompleted) {
		return nil, fmt.Errorf("%w: este deja finalizată", ErrFieldOperationNotCompletable)
	}
	if existing.Status == string(domain.FieldOperationStatusCanceled) {
		return nil, fmt.Errorf("%w: este anulată", ErrFieldOperationNotCompletable)
	}
	if err := validateCompletion(completion); err != nil {
		return nil, err
	}

	endAt := time.Now()
	if completion.ActualEndAt != nil {
		endAt = *completion.ActualEndAt
	}
	if existing.ActualStartAt != nil && endAt.Before(*existing.ActualStartAt) {
		return nil, fmt.Errorf("%w: sfârșitul real este înaintea pornirii", ErrFieldOperationNotCompletable)
	}

	usages, err := service.resolveResourceUsage(existing, completion)
	if err != nil {
		return nil, err
	}

	tx, err := service.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	if err := service.ops.Complete(tx, existing.ID, completion, endAt); err != nil {
		return nil, err
	}
	if completion.MachineHours != nil && *completion.MachineHours > 0 && existing.MachineID != nil {
		if err := service.ops.AddMachineHours(tx, *existing.MachineID, *completion.MachineHours); err != nil {
			return nil, err
		}
	}

	result := &CompletionResult{Movements: []domain.StockMovement{}, LowStocks: []repository.StockLock{}}
	operationID := existing.ID
	for _, usage := range usages {
		if usage.Quantity <= 0 {
			continue
		}
		lock, err := service.movements.LockStockByResource(tx, usage.ResourceID)
		if err != nil {
			return nil, err
		}
		if lock == nil {
			return nil, fmt.Errorf("%w: nu există stoc pentru resursa #%d", ErrFieldOperationNotCompletable, usage.ResourceID)
		}
		movement, err := buildMovement(lock, domain.StockMovementOut, usage.Quantity, nil, &operationID, "Consum la finalizarea operațiunii", actorID)
		if err != nil {
			return nil, fmt.Errorf("resursa #%d: %w", usage.ResourceID, err)
		}
		if err := service.movements.ApplyMovement(tx, movement); err != nil {
			return nil, err
		}
		result.Movements = append(result.Movements, *movement)
		if lock.Minimum > 0 && movement.ResultingQuantity <= lock.Minimum {
			lock.Quantity = movement.ResultingQuantity
			result.LowStocks = append(result.LowStocks, *lock)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	result.Operation, err = reload()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func validateCompletion(completion domain.FieldOperationCompletion) error {
	if completion.AreaCompletedHa != nil && *completion.AreaCompletedHa < 0 {
		return fmt.Errorf("%w: suprafața realizată nu poate fi negativă", ErrFieldOperationNotCompletable)
	}
	if completion.FuelUsedL != nil && *completion.FuelUsedL < 0 {
		return fmt.Errorf("%w: combustibilul consumat nu poate fi negativ", ErrFieldOperationNotCompletable)
	}
	if completion.MachineHours != nil && *completion.MachineHours < 0 {
		return fmt.Errorf("%w: orele de mașină nu pot fi negative", ErrFieldOperationNotCompletable)
	}
	for _, usage := range completion.Resources {
		if usage.ResourceID <= 0 || usage.Quantity < 0 {
			return fmt.Errorf("%w: consum de resurse invalid", ErrFieldOperationNotCompletable)
		}
	}
	return nil
}

// resolveResourceUsage stabilește consumul: cel declarat explicit sau, la cerere, cel din
// șablon (normă pe hectar × suprafața realizată, cu fallback pe cea planificată).
func (service *FieldOperationCompletionService) resolveResourceUsage(existing *dto.FieldOperationResponse, completion domain.FieldOperationCompletion) ([]domain.FieldOperationResourceUsage, error) {
	if len(completion.Resources) > 0 {
		return completion.Resources, nil
	}
	if !completion.ConsumeFromTemplate || existing.OperationTemplateID == nil {
		return nil, nil
	}

	area := 0.0
	if completion.AreaCompletedHa != nil {
		area = *completion.AreaCompletedHa
	} else if existing.AreaPlannedHa != nil {
		area = *existing.AreaPlannedHa
	}
	if area <= 0 {
		return nil, nil
	}

	norms, err := service.ops.GetTemplateResources(*existing.OperationTemplateID)
	if err != nil {
		return nil, err
	}
	usages := make([]domain.FieldOperationResourceUsage, 0, len(norms))
	for _, norm := range norms {
		usages = append(usages, domain.FieldOperationResourceUsage{
			ResourceID: norm.ResourceID,
			Quantity:   norm.QuantityPerUnit * area,
		})
	}
	return usages, nil
}
