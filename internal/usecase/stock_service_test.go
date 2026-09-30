package usecase

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
)

func TestStockService(t *testing.T) {
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
	}
	resourceRepo := &resourceRepoMock{getByID: func(id int64) (*domain.Resource, error) {
		if id == 1 || id == 2 {
			return &domain.Resource{ID: id}, nil
		}
		return nil, nil
	}}
	svc := NewStockService(stockRepo, resourceRepo)

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
		if _, err := svc.CreateStock(in); err == nil {
			t.Errorf("create %s: mă așteptam la eroare", name)
		}
	}
	if _, err := svc.CreateStock(&domain.Stock{ResourceID: 1}); !errors.Is(err, ErrStockAlreadyExists) {
		t.Errorf("stoc duplicat pentru resursă: %v", err)
	}
	created, err := svc.CreateStock(&domain.Stock{ResourceID: 2, Quantity: 5})
	if err != nil || created.ID != 2 {
		t.Fatalf("CreateStock: %v, %+v", err, created)
	}

	if _, err := svc.UpdateStock(0, created); err == nil {
		t.Error("update cu id invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateStock(1, nil); err == nil {
		t.Error("update cu payload nil trebuie să dea eroare")
	}
	if _, err := svc.UpdateStock(9, &domain.Stock{ResourceID: 1}); !errors.Is(err, ErrStockNotFound) {
		t.Errorf("update pe stoc inexistent: %v", err)
	}
	if _, err := svc.UpdateStock(1, &domain.Stock{ResourceID: 99}); !errors.Is(err, ErrStockResourceNotFound) {
		t.Errorf("update cu resursă inexistentă: %v", err)
	}
	// schimbarea resursei către una care are deja stoc
	if _, err := svc.UpdateStock(1, &domain.Stock{ResourceID: 2}); !errors.Is(err, ErrStockAlreadyExists) {
		t.Errorf("update către resursă cu stoc existent: %v", err)
	}
	updated, err := svc.UpdateStock(1, &domain.Stock{ResourceID: 1, Quantity: 20, MinimumQuantity: 3})
	if err != nil || updated.Quantity != 20 || updated.MinimumQuantity != 3 {
		t.Fatalf("UpdateStock: %v, %+v", err, updated)
	}

	if err := svc.DeleteStock(0); err == nil {
		t.Error("delete cu id invalid trebuie să dea eroare")
	}
	if err := svc.DeleteStock(9); !errors.Is(err, ErrStockNotFound) {
		t.Errorf("delete pe stoc inexistent: %v", err)
	}
	if err := svc.DeleteStock(1); err != nil {
		t.Errorf("DeleteStock: %v", err)
	}

	boom := errors.New("db down")
	stockRepo.getByID = func(int64) (*domain.Stock, error) { return nil, boom }
	if _, err := svc.UpdateStock(1, &domain.Stock{ResourceID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteStock(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
	resourceRepo.getByID = func(int64) (*domain.Resource, error) { return nil, boom }
	if _, err := svc.CreateStock(&domain.Stock{ResourceID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de citire a resursei trebuie propagată")
	}
}
