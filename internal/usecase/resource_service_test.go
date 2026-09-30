package usecase

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
)

func newResourceService() (*ResourceService, *resourceTypeRepoMock, *resourceRepoMock) {
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
	return NewResourceService(types, resources), types, resources
}

func TestResourceService_ResourceTypes(t *testing.T) {
	svc, types, _ := newResourceService()

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
	if _, err := svc.UpdateResourceType(2, valid); !errors.Is(err, ErrResourceTypeNotFound) {
		t.Errorf("update pe tip inexistent: %v", err)
	}
	updated, err := svc.UpdateResourceType(1, valid)
	if err != nil || updated.ID != 1 {
		t.Fatalf("UpdateResourceType: %v", err)
	}

	if err := svc.DeleteResourceType(0); err == nil {
		t.Error("delete cu id invalid trebuie să dea eroare")
	}
	if err := svc.DeleteResourceType(2); !errors.Is(err, ErrResourceTypeNotFound) {
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
	svc, types, resources := newResourceService()

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
	}
	for name, in := range invalid {
		if _, err := svc.CreateResource(in); err == nil {
			t.Errorf("create %s: mă așteptam la eroare", name)
		}
		if _, err := svc.UpdateResource(1, in); err == nil {
			t.Errorf("update %s: mă așteptam la eroare", name)
		}
	}

	if _, err := svc.CreateResource(&domain.Resource{Name: "x", ResourceTypeID: 99}); err == nil {
		t.Error("tipul inexistent trebuie să dea eroare")
	}
	created, err := svc.CreateResource(&domain.Resource{Name: "Diesel", ResourceTypeID: 1, PricePerUnit: 7})
	if err != nil || created.ID != 1 {
		t.Fatalf("CreateResource: %v, %+v", err, created)
	}

	if _, err := svc.UpdateResource(0, created); err == nil {
		t.Error("update cu id invalid trebuie să dea eroare")
	}
	if _, err := svc.UpdateResource(2, created); !errors.Is(err, ErrResourceNotFound) {
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
	if err := svc.DeleteResource(2); !errors.Is(err, ErrResourceNotFound) {
		t.Errorf("delete pe resursă inexistentă: %v", err)
	}
	if err := svc.DeleteResource(1); err != nil {
		t.Errorf("DeleteResource: %v", err)
	}

	types.getByID = func(int64) (*domain.ResourceType, error) { return nil, errors.New("db down") }
	if _, err := svc.CreateResource(&domain.Resource{Name: "x", ResourceTypeID: 1}); err == nil {
		t.Error("eroarea de citire a tipului trebuie propagată")
	}
	resources.getByID = func(int64) (*domain.Resource, error) { return nil, errors.New("db down") }
	if _, err := svc.UpdateResource(1, &domain.Resource{Name: "x", ResourceTypeID: 1}); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteResource(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
}
