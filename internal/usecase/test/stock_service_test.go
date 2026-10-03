package usecase_test

import (
	"database/sql"
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
)

func TestStockService(t *testing.T) {
	db, mock := newSQLMock(t)
	stocks := map[int64]*domain.Stock{1: {ID: 1, ResourceID: 1, Quantity: 10, MinimumQuantity: 2}}
	stockRepo := &stockRepoMock{
		getByID: func(id int64) (*domain.Stock, error) { return stocks[id], nil },
		getByResourceID: func(resourceID int64) (*domain.Stock, error) {
			for _, s := range stocks {
				if s.ResourceID == resourceID {
					return s, nil
				}
			}
			return nil, nil
		},
		create: func(s *domain.Stock) error { s.ID = 2; stocks[2] = s; return nil },
		getAll: func() ([]domain.Stock, error) { return []domain.Stock{*stocks[1]}, nil },
		updateMinimum: func(id int64, minimum float64) error {
			stocks[id].MinimumQuantity = minimum
			return nil
		},
	}
	resourceRepo := &resourceRepoMock{getByID: func(id int64) (*domain.Resource, error) {
		if id == 1 || id == 2 {
			return &domain.Resource{ID: id}, nil
		}
		return nil, nil
	}}
	movements := &stockMovementRepoMock{
		lockStockByID: func(_ *sql.Tx, id int64) (*repository.StockLock, error) {
			return &repository.StockLock{StockID: id, ResourceID: stocks[id].ResourceID, PriceUnit: 3}, nil
		},
	}
	var applied *domain.StockMovement
	movements.applyMovement = func(_ *sql.Tx, mv *domain.StockMovement) error {
		applied = mv
		stocks[mv.StockID].Quantity = mv.ResultingQuantity
		return nil
	}
	svc := usecase.NewStockService(db, stockRepo, resourceRepo, movements)

	if _, err := svc.GetStockByID(0); err == nil {
		t.Error("id invalid trebuie să dea eroare")
	}
	if s, err := svc.GetStockByID(1); err != nil || s.Quantity != 10 {
		t.Errorf("GetStockByID: %v", err)
	}
	if all, err := svc.GetStocks(); err != nil || len(all) != 1 {
		t.Errorf("GetStocks: %v", err)
	}

	invalid := map[string]*domain.Stock{
		"nil":                 nil,
		"resursă lipsă":       {},
		"cantitate <0":        {ResourceID: 1, Quantity: -1},
		"minim <0":            {ResourceID: 1, MinimumQuantity: -1},
		"resursă inexistentă": {ResourceID: 99},
	}
	for name, in := range invalid {
		if _, err := svc.CreateStock(in, nil); err == nil {
			t.Errorf("create %s: mă așteptam la eroare", name)
		}
	}
	if _, err := svc.CreateStock(&domain.Stock{ResourceID: 1}, nil); !errors.Is(err, usecase.ErrStockAlreadyExists) {
		t.Errorf("stoc duplicat pentru resursă: %v", err)
	}

	// Cantitatea inițială devine o ajustare de inventar, în aceeași tranzacție cu fișa.
	mock.ExpectBegin()
	mock.ExpectCommit()
	actor := int64(7)
	created, err := svc.CreateStock(&domain.Stock{ResourceID: 2, Quantity: 5, MinimumQuantity: 1}, &actor)
	if err != nil || created.ID != 2 {
		t.Fatalf("CreateStock: %v, %+v", err, created)
	}
	if applied == nil || applied.MovementType != domain.StockMovementAdjustment || applied.QuantityDelta != 5 ||
		applied.ResultingQuantity != 5 || applied.Notes != "Stoc inițial" || *applied.ActorID != 7 {
		t.Fatalf("mișcarea de stoc inițial: %+v", applied)
	}
	if created.Quantity != 5 || created.MinimumQuantity != 1 {
		t.Errorf("stocul creat trebuie să aibă cantitatea din mișcare: %+v", created)
	}

	// Fără cantitate inițială nu se generează nicio mișcare.
	delete(stocks, 2)
	applied = nil
	mock.ExpectBegin()
	mock.ExpectCommit()
	if _, err := svc.CreateStock(&domain.Stock{ResourceID: 2}, nil); err != nil || applied != nil {
		t.Errorf("stoc gol: %v, mișcare %+v", err, applied)
	}

	if _, err := svc.UpdateMinimum(0, 1); err == nil {
		t.Error("update cu id invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateMinimum(1, -1); err == nil {
		t.Error("update cu minim negativ trebuie să dea eroare")
	}
	if _, err := svc.UpdateMinimum(9, 1); !errors.Is(err, usecase.ErrStockNotFound) {
		t.Errorf("update pe stoc inexistent: %v", err)
	}
	updated, err := svc.UpdateMinimum(1, 3)
	if err != nil || updated.Quantity != 10 || updated.MinimumQuantity != 3 {
		t.Fatalf("UpdateMinimum trebuie să schimbe doar minimul: %v, %+v", err, updated)
	}

	if err := svc.DeleteStock(0); err == nil {
		t.Error("delete cu id invalid trebuie să dea eroare")
	}
	if err := svc.DeleteStock(9); !errors.Is(err, usecase.ErrStockNotFound) {
		t.Errorf("delete pe stoc inexistent: %v", err)
	}
	if err := svc.DeleteStock(1); err != nil {
		t.Errorf("DeleteStock: %v", err)
	}

	boom := errors.New("db down")
	stockRepo.getByID = func(int64) (*domain.Stock, error) { return nil, boom }
	if _, err := svc.UpdateMinimum(1, 1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteStock(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
	resourceRepo.getByID = func(int64) (*domain.Resource, error) { return nil, boom }
	if _, err := svc.CreateStock(&domain.Stock{ResourceID: 1}, nil); !errors.Is(err, boom) {
		t.Error("eroarea de citire a resursei trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
