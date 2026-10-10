package usecase_test

import (
	"database/sql"
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	"github.com/DATA-DOG/go-sqlmock"
)

func newResourceService(t *testing.T) (*usecase.ResourceService, *resourceTypeRepoMock, *resourceRepoMock, sqlmock.Sqlmock) {
	db, mock := newSQLMock(t)
	types := &resourceTypeRepoMock{
		getByID: func(id int64) (*domain.ResourceType, error) {
			if id == 1 {
				return &domain.ResourceType{ID: 1, Name: "Motorină", Category: domain.ResourceCategoryFuel, DefaultUnit: "l"}, nil
			}
			return nil, nil
		},
		create: func(rt *domain.ResourceType) error { rt.ID = 10; return nil },
	}
	resources := &resourceRepoMock{
		getByID: func(id int64) (*domain.Resource, error) {
			if id == 1 {
				return &domain.Resource{ID: 1, Name: "Diesel", ResourceTypeID: 1, PricePerUnit: 7}, nil
			}
			return nil, nil
		},
		create: func(r *domain.Resource) error { r.ID = 1; return nil },
	}
	return usecase.NewResourceService(db, types, resources, &stockMovementRepoMock{}), types, resources, mock
}

func TestResourceService_ResourceTypes(t *testing.T) {
	svc, types, _, _ := newResourceService(t)

	if _, err := svc.GetResourceTypeByID(0); err == nil {
		t.Error("id invalid trebuie să dea eroare")
	}
	if rt, err := svc.GetResourceTypeByID(1); err != nil || rt.Name != "Motorină" {
		t.Errorf("GetResourceTypeByID: %v", err)
	}

	invalid := map[string]*domain.ResourceType{
		"nil":                nil,
		"nume lipsă":         {Category: domain.ResourceCategoryFuel, DefaultUnit: "l"},
		"categorie invalidă": {Name: "x", Category: "gaz", DefaultUnit: "l"},
		"unitate lipsă":      {Name: "x", Category: domain.ResourceCategoryFuel},
	}
	for name, in := range invalid {
		if _, err := svc.CreateResourceType(in); err == nil {
			t.Errorf("create %s: mă așteptam la eroare", name)
		}
		if _, err := svc.UpdateResourceType(1, in); err == nil {
			t.Errorf("update %s: mă așteptam la eroare", name)
		}
	}

	valid := &domain.ResourceType{Name: "Semințe", Category: domain.ResourceCategorySeed, DefaultUnit: "kg"}
	created, err := svc.CreateResourceType(valid)
	if err != nil || created.ID != 10 {
		t.Fatalf("CreateResourceType: %v, %+v", err, created)
	}

	if _, err := svc.UpdateResourceType(0, valid); err == nil {
		t.Error("update cu id invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateResourceType(2, valid); !errors.Is(err, usecase.ErrResourceTypeNotFound) {
		t.Errorf("update pe tip inexistent: %v", err)
	}
	updated, err := svc.UpdateResourceType(1, valid)
	if err != nil || updated.ID != 1 {
		t.Fatalf("UpdateResourceType: %v", err)
	}

	if err := svc.DeleteResourceType(0); err == nil {
		t.Error("delete cu id invalid trebuie să dea eroare")
	}
	if err := svc.DeleteResourceType(2); !errors.Is(err, usecase.ErrResourceTypeNotFound) {
		t.Errorf("delete pe tip inexistent: %v", err)
	}
	if err := svc.DeleteResourceType(1); err != nil {
		t.Errorf("DeleteResourceType: %v", err)
	}
	if _, err := svc.GetResourceTypes(); err != nil {
		t.Errorf("GetResourceTypes: %v", err)
	}

	types.getByID = func(int64) (*domain.ResourceType, error) { return nil, errors.New("db down") }
	if _, err := svc.UpdateResourceType(1, valid); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteResourceType(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
}

func TestResourceService_Resources(t *testing.T) {
	svc, types, resources, mock := newResourceService(t)

	if _, err := svc.GetResourceByID(0); err == nil {
		t.Error("id invalid trebuie să dea eroare")
	}
	if r, err := svc.GetResourceByID(1); err != nil || r.Name != "Diesel" {
		t.Errorf("GetResourceByID: %v", err)
	}
	if _, err := svc.GetResources(); err != nil {
		t.Errorf("GetResources: %v", err)
	}

	invalid := map[string]*domain.Resource{
		"nil":          nil,
		"nume lipsă":   {ResourceTypeID: 1},
		"tip lipsă":    {Name: "x"},
		"preț negativ": {Name: "x", ResourceTypeID: 1, PricePerUnit: -1},
		"minim <0":     {Name: "x", ResourceTypeID: 1, MinimumQuantity: -1},
	}
	for name, in := range invalid {
		if _, err := svc.CreateResource(in, nil); err == nil {
			t.Errorf("create %s: mă așteptam la eroare", name)
		}
		if _, err := svc.UpdateResource(1, in); err == nil {
			t.Errorf("update %s: mă așteptam la eroare", name)
		}
	}

	if _, err := svc.CreateResource(&domain.Resource{Name: "x", ResourceTypeID: 1, Quantity: -1}, nil); err == nil {
		t.Error("cantitatea inițială negativă trebuie să dea eroare")
	}
	if _, err := svc.CreateResource(&domain.Resource{Name: "x", ResourceTypeID: 99}, nil); err == nil {
		t.Error("tipul inexistent trebuie să dea eroare")
	}
	mock.ExpectBegin()
	mock.ExpectCommit()
	created, err := svc.CreateResource(&domain.Resource{Name: "Diesel", ResourceTypeID: 1, PricePerUnit: 7}, nil)
	if err != nil || created.ID != 1 {
		t.Fatalf("CreateResource: %v, %+v", err, created)
	}

	if _, err := svc.UpdateResource(0, created); err == nil {
		t.Error("update cu id invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateResource(2, created); !errors.Is(err, usecase.ErrResourceNotFound) {
		t.Errorf("update pe resursă inexistentă: %v", err)
	}
	if _, err := svc.UpdateResource(1, &domain.Resource{Name: "x", ResourceTypeID: 99}); err == nil {
		t.Error("update cu tip inexistent trebuie să dea eroare")
	}
	if updated, err := svc.UpdateResource(1, &domain.Resource{Name: "Diesel Premium", ResourceTypeID: 1, PricePerUnit: 8}); err != nil || updated == nil {
		t.Fatalf("UpdateResource: %v", err)
	}

	if err := svc.DeleteResource(0); err == nil {
		t.Error("delete cu id invalid trebuie să dea eroare")
	}
	if err := svc.DeleteResource(2); !errors.Is(err, usecase.ErrResourceNotFound) {
		t.Errorf("delete pe resursă inexistentă: %v", err)
	}
	if err := svc.DeleteResource(1); err != nil {
		t.Errorf("DeleteResource: %v", err)
	}

	types.getByID = func(int64) (*domain.ResourceType, error) { return nil, errors.New("db down") }
	if _, err := svc.CreateResource(&domain.Resource{Name: "x", ResourceTypeID: 1}, nil); err == nil {
		t.Error("eroarea de citire a tipului trebuie propagată")
	}
	resources.getByID = func(int64) (*domain.Resource, error) { return nil, errors.New("db down") }
	if _, err := svc.UpdateResource(1, &domain.Resource{Name: "x", ResourceTypeID: 1}); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteResource(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// T12: resursa se creează cu stocul ei. Cantitatea inițială devine o ajustare de inventar, în
// aceeași tranzacție cu resursa; fără cantitate nu se generează nicio mișcare.
func TestResourceService_CreateWithInitialStock(t *testing.T) {
	db, mock := newSQLMock(t)
	stored := map[int64]*domain.Resource{}
	types := &resourceTypeRepoMock{getByID: func(id int64) (*domain.ResourceType, error) {
		return &domain.ResourceType{ID: id}, nil
	}}
	resources := &resourceRepoMock{
		create: func(r *domain.Resource) error {
			r.ID = int64(len(stored) + 1)
			saved := *r
			stored[r.ID] = &saved
			return nil
		},
		getByID: func(id int64) (*domain.Resource, error) { return stored[id], nil },
	}
	var applied *domain.StockMovement
	movements := &stockMovementRepoMock{
		lockStock: func(_ *sql.Tx, id int64) (*repository.StockLock, error) {
			return &repository.StockLock{ResourceID: id, Quantity: stored[id].Quantity, PriceUnit: 3}, nil
		},
		applyMovement: func(_ *sql.Tx, mv *domain.StockMovement) error {
			applied = mv
			stored[mv.ResourceID].Quantity = mv.ResultingQuantity
			return nil
		},
	}
	svc := usecase.NewResourceService(db, types, resources, movements)

	mock.ExpectBegin()
	mock.ExpectCommit()
	actor := int64(7)
	created, err := svc.CreateResource(&domain.Resource{Name: "Uree", ResourceTypeID: 1, Quantity: 5, MinimumQuantity: 1}, &actor)
	if err != nil {
		t.Fatalf("CreateResource: %v", err)
	}
	if applied == nil || applied.MovementType != domain.StockMovementAdjustment || applied.QuantityDelta != 5 ||
		applied.ResultingQuantity != 5 || applied.Notes != "Stoc inițial" || *applied.ActorID != 7 {
		t.Fatalf("mișcarea de stoc inițial: %+v", applied)
	}
	if created.Quantity != 5 || created.MinimumQuantity != 1 {
		t.Errorf("resursa creată trebuie să aibă cantitatea din mișcare: %+v", created)
	}

	applied = nil
	mock.ExpectBegin()
	mock.ExpectCommit()
	if _, err := svc.CreateResource(&domain.Resource{Name: "NPK", ResourceTypeID: 1}, nil); err != nil || applied != nil {
		t.Errorf("resursă fără stoc inițial: %v, mișcare %+v", err, applied)
	}

	// o mișcare eșuată anulează și crearea resursei
	boom := errors.New("db down")
	movements.applyMovement = func(*sql.Tx, *domain.StockMovement) error { return boom }
	mock.ExpectBegin()
	mock.ExpectRollback()
	if _, err := svc.CreateResource(&domain.Resource{Name: "Motorină", ResourceTypeID: 1, Quantity: 10}, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la mișcare trebuie propagată: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
