package handlers_test

import (
	"net/http"
	"testing"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type resourceRepoStub struct {
	repository.ResourceRepository
	resource domain.Resource
	updated  *domain.Resource
}

func (m *resourceRepoStub) GetByID(int64) (*domain.Resource, error) {
	resource := m.resource
	return &resource, nil
}
func (m *resourceRepoStub) Update(_ int64, r *domain.Resource) error {
	saved := *r
	m.updated = &saved
	m.resource = saved
	return nil
}

type resourceTypeRepoStub struct {
	repository.ResourceTypeRepository
}

func (resourceTypeRepoStub) GetByID(id int64) (*domain.ResourceType, error) {
	return &domain.ResourceType{ID: id}, nil
}

// Cantitatea se modifică doar prin mișcări: PATCH pe resursă schimbă pragul minim, nu cantitatea.
func TestResourceHandler_Update_QuantityOnlyThroughMovements(t *testing.T) {
	repo := &resourceRepoStub{resource: domain.Resource{ID: 1, Name: "Uree", ResourceTypeID: 2, PricePerUnit: 3, Quantity: 10, MinimumQuantity: 2}}
	notif := &notificationRepoMock{}
	r := gin.New()
	r.PATCH("/resources/:id", handlers.NewResourceHandler(
		usecase.NewResourceService(nil, resourceTypeRepoStub{}, repo, nil),
		handlers.WithResourceNotif(usecase.NewNotificationService(notif)),
	).UpdateResource)

	rejected := map[string]string{
		"cantitate":          `{"name":"Uree","resource_type_id":2,"price_per_unit":3,"quantity":50,"minimum_quantity":3}`,
		"fără prag minim":    `{"name":"Uree","resource_type_id":2,"price_per_unit":3}`,
		"prag minim negativ": `{"name":"Uree","resource_type_id":2,"price_per_unit":3,"minimum_quantity":-1}`,
		"fără preț":          `{"name":"Uree","resource_type_id":2,"minimum_quantity":3}`,
		"preț negativ":       `{"name":"Uree","resource_type_id":2,"price_per_unit":-1,"minimum_quantity":3}`,
	}
	for name, body := range rejected {
		if rec := do(t, r, http.MethodPatch, "/resources/1", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: mă așteptam la 400, nu %d: %s", name, rec.Code, rec.Body)
		}
	}
	if repo.updated != nil {
		t.Fatalf("cererile respinse nu trebuie să modifice resursa: %+v", repo.updated)
	}

	rec := do(t, r, http.MethodPatch, "/resources/1", `{"name":"Uree 46%","resource_type_id":2,"price_per_unit":3,"minimum_quantity":2}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mă așteptam la 200, nu %d: %s", rec.Code, rec.Body)
	}
	if repo.updated == nil || repo.updated.Name != "Uree 46%" || repo.updated.Quantity != 10 {
		t.Errorf("cantitatea trebuie să rămână neschimbată: %+v", repo.updated)
	}
	waitUntil(t, func() bool { return notif.createdCount() == 1 })

	// un prag minim ridicat peste stocul curent anunță stocul scăzut; prețul 0 e acceptat
	// (resursele de recoltă sunt create automat cu preț 0)
	rec = do(t, r, http.MethodPatch, "/resources/1", `{"name":"Uree 46%","resource_type_id":2,"price_per_unit":0,"minimum_quantity":12}`)
	if rec.Code != http.StatusOK || repo.updated.MinimumQuantity != 12 || repo.updated.PricePerUnit != 0 {
		t.Fatalf("prag minim cu preț 0: %d %s", rec.Code, rec.Body)
	}
	waitUntil(t, func() bool { return notif.createdCount() == 3 })
}
