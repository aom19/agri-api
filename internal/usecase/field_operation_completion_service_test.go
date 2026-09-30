package usecase

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
)

func inProgressOperation() *dto.FieldOperationResponse {
	start := time.Now().Add(-2 * time.Hour)
	return &dto.FieldOperationResponse{
		ID: 1, Status: string(domain.FieldOperationStatusInProgress),
		MachineID: ptr(int64(4)), OperationTemplateID: ptr(int64(3)),
		AreaPlannedHa: ptr(10.0), ActualStartAt: &start,
	}
}

func TestValidateCompletion(t *testing.T) {
	cases := map[string]domain.FieldOperationCompletion{
		"suprafață negativă":  {AreaCompletedHa: ptr(-1.0)},
		"combustibil negativ": {FuelUsedL: ptr(-1.0)},
		"ore negative":        {MachineHours: ptr(-1.0)},
		"resursă invalidă":    {Resources: []domain.FieldOperationResourceUsage{{ResourceID: 0, Quantity: 1}}},
		"cantitate negativă":  {Resources: []domain.FieldOperationResourceUsage{{ResourceID: 1, Quantity: -1}}},
	}
	for name, c := range cases {
		if err := validateCompletion(c); !errors.Is(err, ErrFieldOperationNotCompletable) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := validateCompletion(domain.FieldOperationCompletion{AreaCompletedHa: ptr(1.0)}); err != nil {
		t.Errorf("completare validă: %v", err)
	}
}

func TestFieldOperationCompletionService_Complete(t *testing.T) {
	db, mock := newSQLMock(t)
	current := inProgressOperation()
	ops := &fieldOperationRepoMock{
		getByID: func(id int64) (*dto.FieldOperationResponse, error) {
			if id == 1 {
				return current, nil
			}
			return nil, nil
		},
		getTemplateResources: func(int64) ([]domain.TemplateResourceUsage, error) {
			return []domain.TemplateResourceUsage{{ResourceID: 1, QuantityPerUnit: 2, PricePerUnit: 5}}, nil
		},
	}
	stocks := map[int64]*repository.StockLock{1: {StockID: 10, ResourceID: 1, Quantity: 100, Minimum: 90, PriceUnit: 5}}
	movements := &stockMovementRepoMock{
		lockStockByResource: func(_ *sql.Tx, id int64) (*repository.StockLock, error) { return stocks[id], nil },
	}
	svc := NewFieldOperationCompletionService(db, ops, movements)

	if _, err := svc.Complete(9, domain.FieldOperationCompletion{}, nil); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("operațiune inexistentă: %v", err)
	}

	current.Status = string(domain.FieldOperationStatusCompleted)
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{}, nil); !errors.Is(err, ErrFieldOperationNotCompletable) {
		t.Errorf("deja finalizată: %v", err)
	}
	current.Status = string(domain.FieldOperationStatusCanceled)
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{}, nil); !errors.Is(err, ErrFieldOperationNotCompletable) {
		t.Errorf("anulată: %v", err)
	}
	current.Status = string(domain.FieldOperationStatusInProgress)

	if _, err := svc.Complete(1, domain.FieldOperationCompletion{FuelUsedL: ptr(-1.0)}, nil); !errors.Is(err, ErrFieldOperationNotCompletable) {
		t.Errorf("validare: %v", err)
	}
	tooEarly := current.ActualStartAt.Add(-time.Hour)
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{ActualEndAt: &tooEarly}, nil); !errors.Is(err, ErrFieldOperationNotCompletable) {
		t.Errorf("sfârșit înaintea pornirii: %v", err)
	}

	// succes cu consum explicit, ore mașină și stoc care ajunge sub prag
	var hoursAdded float64
	ops.addMachineHours = func(_ *sql.Tx, _ int64, h float64) error { hoursAdded = h; return nil }
	mock.ExpectBegin()
	mock.ExpectCommit()
	result, err := svc.Complete(1, domain.FieldOperationCompletion{
		MachineHours: ptr(2.5),
		Resources:    []domain.FieldOperationResourceUsage{{ResourceID: 1, Quantity: 15}, {ResourceID: 2, Quantity: 0}},
	}, ptr(int64(7)))
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if hoursAdded != 2.5 || len(result.Movements) != 1 || result.Movements[0].ResultingQuantity != 85 || len(result.LowStocks) != 1 || result.Operation == nil {
		t.Errorf("rezultat greșit: %+v", result)
	}

	// consum din șablon: 2 × 10 ha planificate = 20
	mock.ExpectBegin()
	mock.ExpectCommit()
	result, err = svc.Complete(1, domain.FieldOperationCompletion{ConsumeFromTemplate: true}, nil)
	if err != nil || len(result.Movements) != 1 || result.Movements[0].QuantityDelta != -20 {
		t.Fatalf("consum din șablon: %v, %+v", err, result)
	}

	// suprafață realizată explicită are prioritate față de cea planificată
	mock.ExpectBegin()
	mock.ExpectCommit()
	result, err = svc.Complete(1, domain.FieldOperationCompletion{ConsumeFromTemplate: true, AreaCompletedHa: ptr(5.0)}, nil)
	if err != nil || result.Movements[0].QuantityDelta != -10 {
		t.Fatalf("consum din șablon cu suprafață realizată: %v, %+v", err, result)
	}

	// fără suprafață → fără consum
	current.AreaPlannedHa = nil
	mock.ExpectBegin()
	mock.ExpectCommit()
	if result, err = svc.Complete(1, domain.FieldOperationCompletion{ConsumeFromTemplate: true}, nil); err != nil || len(result.Movements) != 0 {
		t.Fatalf("fără suprafață: %v, %+v", err, result)
	}
	current.AreaPlannedHa = ptr(10.0)

	// resursă fără stoc
	mock.ExpectBegin()
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{Resources: []domain.FieldOperationResourceUsage{{ResourceID: 2, Quantity: 1}}}, nil); !errors.Is(err, ErrFieldOperationNotCompletable) {
		t.Errorf("resursă fără stoc: %v", err)
	}
	// stoc insuficient
	mock.ExpectBegin()
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{Resources: []domain.FieldOperationResourceUsage{{ResourceID: 1, Quantity: 1000}}}, nil); err == nil {
		t.Error("stocul insuficient trebuie să dea eroare")
	}

	boom := errors.New("db down")
	mock.ExpectBegin()
	ops.complete = func(*sql.Tx, int64, domain.FieldOperationCompletion, time.Time) error { return boom }
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{}, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la finalizare trebuie propagată: %v", err)
	}
	ops.complete = nil
	ops.getTemplateResources = func(int64) ([]domain.TemplateResourceUsage, error) { return nil, boom }
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{ConsumeFromTemplate: true}, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la citirea șablonului trebuie propagată: %v", err)
	}
	mock.ExpectBegin().WillReturnError(boom)
	if _, err := svc.Complete(1, domain.FieldOperationCompletion{}, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la begin trebuie propagată: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestFieldOperationCompletionService_CompleteForAssignedUser(t *testing.T) {
	db, mock := newSQLMock(t)
	current := inProgressOperation()
	ops := &fieldOperationRepoMock{
		getByIDForAssignedUser: func(id, userID int64) (*dto.FieldOperationResponse, error) {
			if id == 1 && userID == 7 {
				return current, nil
			}
			return nil, nil
		},
	}
	svc := NewFieldOperationCompletionService(db, ops, &stockMovementRepoMock{})

	if _, err := svc.CompleteForAssignedUser(1, 8, domain.FieldOperationCompletion{}, nil); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("alt utilizator: %v", err)
	}
	mock.ExpectBegin()
	mock.ExpectCommit()
	result, err := svc.CompleteForAssignedUser(1, 7, domain.FieldOperationCompletion{Notes: "gata"}, ptr(int64(7)))
	if err != nil || result.Operation != current {
		t.Fatalf("CompleteForAssignedUser: %v", err)
	}

	boom := errors.New("db down")
	ops.getByIDForAssignedUser = func(int64, int64) (*dto.FieldOperationResponse, error) { return nil, boom }
	if _, err := svc.CompleteForAssignedUser(1, 7, domain.FieldOperationCompletion{}, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea de citire trebuie propagată: %v", err)
	}
}
