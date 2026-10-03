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
	fuel := fuelUsage(completion)

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
	consume := func(usage domain.FieldOperationResourceUsage, notes string, fuelOnly bool) error {
		if usage.Quantity <= 0 {
			return nil
		}
		lock, err := service.movements.LockStockByResource(tx, usage.ResourceID)
		if err != nil {
			return err
		}
		if lock == nil {
			return fmt.Errorf("%w: nu există stoc pentru resursa #%d", ErrFieldOperationNotCompletable, usage.ResourceID)
		}
		if fuelOnly && lock.Category != fuelCategory {
			return fmt.Errorf("%w: resursa #%d nu este combustibil", ErrFieldOperationNotCompletable, usage.ResourceID)
		}
		operationID := existing.ID
		movement, err := buildMovement(lock, domain.StockMovementOut, usage.Quantity, nil, &operationID, notes, actorID)
		if err != nil {
			return fmt.Errorf("resursa #%d: %w", usage.ResourceID, err)
		}
		if err := service.movements.ApplyMovement(tx, movement); err != nil {
			return err
		}
		result.Movements = append(result.Movements, *movement)
		if lock.Minimum > 0 && movement.ResultingQuantity <= lock.Minimum {
			lock.Quantity = movement.ResultingQuantity
			result.LowStocks = append(result.LowStocks, *lock)
		}
		return nil
	}
	for _, usage := range usages {
		if err := consume(usage, "Consum la finalizarea operațiunii", false); err != nil {
			return nil, err
		}
	}
	if fuel != nil {
		if err := consume(*fuel, "Combustibil raportat la finalizare", true); err != nil {
			return nil, err
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

// fuelCategory este categoria resurselor de combustibil.
const fuelCategory = "fuel"

func validateCompletion(completion domain.FieldOperationCompletion) error {
	if completion.AreaCompletedHa != nil && *completion.AreaCompletedHa < 0 {
		return fmt.Errorf("%w: suprafața realizată nu poate fi negativă", ErrFieldOperationNotCompletable)
	}
	if completion.FuelUsedL != nil && *completion.FuelUsedL < 0 {
		return fmt.Errorf("%w: combustibilul consumat nu poate fi negativ", ErrFieldOperationNotCompletable)
	}
	if completion.FuelUsedL != nil && *completion.FuelUsedL > 0 && (completion.FuelResourceID == nil || *completion.FuelResourceID <= 0) {
		return fmt.Errorf("%w: alege resursa de combustibil din care se scade consumul", ErrFieldOperationNotCompletable)
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

// fuelUsage este ieșirea de combustibil generată de câmpul „combustibil consumat”, sau nil.
func fuelUsage(completion domain.FieldOperationCompletion) *domain.FieldOperationResourceUsage {
	if completion.FuelUsedL == nil || completion.FuelResourceID == nil {
		return nil
	}
	return &domain.FieldOperationResourceUsage{ResourceID: *completion.FuelResourceID, Quantity: *completion.FuelUsedL}
}

// resolveResourceUsage stabilește consumul de resurse, fără combustibilul raportat:
//   - cu consume_from_template, pornește de la normele șablonului (normă × suprafață), iar
//     resursele trimise explicit înlocuiesc norma aceleiași resurse sau se adaugă;
//   - altfel, doar resursele trimise explicit.
//
// Resursa de combustibil aleasă e scoasă din listă: consumul ei vine doar din fuel_used_l.
func (service *FieldOperationCompletionService) resolveResourceUsage(existing *dto.FieldOperationResponse, completion domain.FieldOperationCompletion) ([]domain.FieldOperationResourceUsage, error) {
	usages := []domain.FieldOperationResourceUsage{}
	if completion.ConsumeFromTemplate {
		estimate, err := service.estimateUsage(existing, completion.AreaCompletedHa)
		if err != nil {
			return nil, err
		}
		for _, item := range estimate {
			usages = append(usages, domain.FieldOperationResourceUsage{ResourceID: item.ResourceID, Quantity: item.Quantity})
		}
	}
	for _, explicit := range completion.Resources {
		replaced := false
		for i := range usages {
			if usages[i].ResourceID == explicit.ResourceID {
				usages[i].Quantity = explicit.Quantity
				replaced = true
			}
		}
		if !replaced {
			usages = append(usages, explicit)
		}
	}

	if fuel := fuelUsage(completion); fuel != nil {
		filtered := usages[:0]
		for _, usage := range usages {
			if usage.ResourceID != fuel.ResourceID {
				filtered = append(filtered, usage)
			}
		}
		usages = filtered
	}
	return usages, nil
}

// operationArea este suprafața pe care se aplică normele: cea realizată sau, în lipsă, cea planificată.
func operationArea(existing *dto.FieldOperationResponse, areaCompleted *float64) float64 {
	if areaCompleted != nil {
		return *areaCompleted
	}
	if existing.AreaPlannedHa != nil {
		return *existing.AreaPlannedHa
	}
	return 0
}

// estimateUsage calculează consumul din normele șablonului: normă pe hectar × suprafață.
// Este singurul loc din aplicație unde se face acest calcul pentru o operațiune.
func (service *FieldOperationCompletionService) estimateUsage(existing *dto.FieldOperationResponse, areaCompleted *float64) ([]domain.ConsumptionEstimateItem, error) {
	items := []domain.ConsumptionEstimateItem{}
	if existing.OperationTemplateID == nil {
		return items, nil
	}
	norms, err := service.ops.GetTemplateResources(*existing.OperationTemplateID)
	if err != nil {
		return nil, err
	}
	area := operationArea(existing, areaCompleted)
	if area < 0 {
		area = 0
	}
	for _, norm := range norms {
		items = append(items, domain.ConsumptionEstimateItem{
			ResourceID:      norm.ResourceID,
			ResourceName:    norm.ResourceName,
			Category:        norm.Category,
			Unit:            norm.Unit,
			QuantityPerUnit: norm.QuantityPerUnit,
			Quantity:        roundQuantity(norm.QuantityPerUnit * area),
		})
	}
	return items, nil
}

// Estimate întoarce previzualizarea consumului la finalizare, pentru suprafața dată
// (sau cea planificată), plus resursele de combustibil din care se poate scădea motorina.
func (service *FieldOperationCompletionService) Estimate(id int64, areaCompleted *float64) (*domain.ConsumptionEstimate, error) {
	existing, err := service.ops.GetByID(id)
	if err != nil {
		return nil, err
	}
	return service.estimate(existing, areaCompleted)
}

func (service *FieldOperationCompletionService) EstimateForAssignedUser(id, userID int64, areaCompleted *float64) (*domain.ConsumptionEstimate, error) {
	existing, err := service.ops.GetByIDForAssignedUser(id, userID)
	if err != nil {
		return nil, err
	}
	return service.estimate(existing, areaCompleted)
}

func (service *FieldOperationCompletionService) estimate(existing *dto.FieldOperationResponse, areaCompleted *float64) (*domain.ConsumptionEstimate, error) {
	if existing == nil {
		return nil, ErrFieldOperationNotFound
	}
	if areaCompleted != nil && *areaCompleted < 0 {
		return nil, fmt.Errorf("%w: suprafața realizată nu poate fi negativă", ErrFieldOperationNotCompletable)
	}
	items, err := service.estimateUsage(existing, areaCompleted)
	if err != nil {
		return nil, err
	}
	fuel, err := service.movements.ListFuelStocks()
	if err != nil {
		return nil, err
	}
	return &domain.ConsumptionEstimate{
		AreaHa:        operationArea(existing, areaCompleted),
		Items:         items,
		FuelResources: fuel,
	}, nil
}
