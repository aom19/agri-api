package usecase

import (
	"database/sql"
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

func TestBuildMovement(t *testing.T) {
	lock := &repository.StockLock{StockID: 1, ResourceID: 2, Quantity: 10, Minimum: 3, PriceUnit: 5}

	in, err := buildMovement(lock, domain.StockMovementIn, 4, nil, nil, "recepție", ptr(int64(9)))
	if err != nil || in.QuantityDelta != 4 || in.ResultingQuantity != 14 || *in.TotalCost != 20 || *in.UnitCost != 5 {
		t.Fatalf("intrare: %v, %+v", err, in)
	}
	if in.StockID != 1 || in.ResourceID != 2 || in.Notes != "recepție" || *in.ActorID != 9 {
		t.Errorf("câmpurile mișcării nu sunt copiate din stoc: %+v", in)
	}

	out, err := buildMovement(lock, domain.StockMovementOut, 4, ptr(2.0), ptr(int64(3)), "", nil)
	if err != nil || out.QuantityDelta != -4 || out.ResultingQuantity != 6 || *out.TotalCost != 8 || *out.FieldOperationID != 3 {
		t.Fatalf("ieșire cu cost explicit: %v, %+v", err, out)
	}

	adj, err := buildMovement(lock, domain.StockMovementAdjustment, 7, ptr(-1.0), nil, "", nil)
	if err != nil || adj.QuantityDelta != -3 || adj.ResultingQuantity != 7 || *adj.UnitCost != 5 {
		t.Fatalf("ajustare (costul negativ trebuie ignorat): %v, %+v", err, adj)
	}

	if _, err := buildMovement(lock, "teleport", 1, nil, nil, "", nil); !errors.Is(err, ErrInvalidStockMovement) {
		t.Errorf("tip invalid: %v", err)
	}
	if _, err := buildMovement(lock, domain.StockMovementOut, 11, nil, nil, "", nil); !errors.Is(err, ErrInvalidStockMovement) {
		t.Errorf("stoc insuficient: %v", err)
	}
}

func TestStockMovementService_Create(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &stockMovementRepoMock{}
	svc := NewStockMovementService(db, repo)

	invalid := map[string]domain.StockMovementInput{
		"stoc lipsă":        {MovementType: domain.StockMovementIn, Quantity: 1},
		"cantitate zero":    {StockID: 1, MovementType: domain.StockMovementIn},
		"ajustare negativă": {StockID: 1, MovementType: domain.StockMovementAdjustment, Quantity: -1},
		"tip necunoscut":    {StockID: 1, MovementType: "x", Quantity: 1},
	}
	for name, in := range invalid {
		if _, err := svc.Create(in); !errors.Is(err, ErrInvalidStockMovement) {
			t.Errorf("%s: %v", name, err)
		}
	}

	valid := domain.StockMovementInput{StockID: 1, MovementType: domain.StockMovementOut, Quantity: 8}

	mock.ExpectBegin()
	if _, err := svc.Create(valid); !errors.Is(err, ErrStockNotFound) {
		t.Errorf("stoc inexistent: %v", err)
	}

	repo.lockStockByID = func(_ *sql.Tx, id int64) (*repository.StockLock, error) {
		return &repository.StockLock{StockID: id, ResourceID: 2, Quantity: 10, Minimum: 3, PriceUnit: 5}, nil
	}
	var applied *domain.StockMovement
	repo.applyMovement = func(_ *sql.Tx, mv *domain.StockMovement) error { applied = mv; return nil }

	mock.ExpectBegin()
	mock.ExpectCommit()
	result, err := svc.Create(valid)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if applied == nil || result.Movement.ResultingQuantity != 2 || !result.BelowMinimum || result.Minimum != 3 {
		t.Errorf("rezultat greșit: %+v", result)
	}

	mock.ExpectBegin()
	if _, err := svc.Create(domain.StockMovementInput{StockID: 1, MovementType: domain.StockMovementOut, Quantity: 50}); !errors.Is(err, ErrInvalidStockMovement) {
		t.Errorf("stoc insuficient: %v", err)
	}

	boom := errors.New("db down")
	repo.applyMovement = func(*sql.Tx, *domain.StockMovement) error { return boom }
	mock.ExpectBegin()
	if _, err := svc.Create(valid); !errors.Is(err, boom) {
		t.Errorf("eroarea la aplicare trebuie propagată: %v", err)
	}
	repo.lockStockByID = func(*sql.Tx, int64) (*repository.StockLock, error) { return nil, boom }
	mock.ExpectBegin()
	if _, err := svc.Create(valid); !errors.Is(err, boom) {
		t.Errorf("eroarea la blocare trebuie propagată: %v", err)
	}
	mock.ExpectBegin().WillReturnError(boom)
	if _, err := svc.Create(valid); !errors.Is(err, boom) {
		t.Errorf("eroarea la begin trebuie propagată: %v", err)
	}

	repo.list = func(domain.StockMovementFilter) ([]domain.StockMovement, error) {
		return []domain.StockMovement{{ID: 1}}, nil
	}
	if items, err := svc.List(domain.StockMovementFilter{}); err != nil || len(items) != 1 {
		t.Errorf("List: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}
