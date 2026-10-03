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

type stockRepoStub struct {
	repository.StockRepository
	stock   domain.Stock
	minimum *float64
}

func (m *stockRepoStub) GetByID(int64) (*domain.Stock, error) {
	stock := m.stock
	return &stock, nil
}
func (m *stockRepoStub) UpdateMinimum(_ int64, minimum float64) error {
	m.minimum = &minimum
	m.stock.MinimumQuantity = minimum
	return nil
}

// Cantitatea se modifică doar prin mișcări: PATCH acceptă doar pragul minim.
func TestStockHandler_Update_OnlyMinimum(t *testing.T) {
	repo := &stockRepoStub{stock: domain.Stock{ID: 1, ResourceID: 4, Quantity: 10, MinimumQuantity: 2}}
	r := gin.New()
	r.PATCH("/stocks/:id", handlers.NewStockHandler(usecase.NewStockService(nil, repo, nil, nil)).Update)

	rejected := map[string]string{
		"cantitate":          `{"quantity":50,"minimum_quantity":3}`,
		"altă resursă":       `{"resource_id":5,"minimum_quantity":3}`,
		"fără prag minim":    `{}`,
		"prag minim negativ": `{"minimum_quantity":-1}`,
	}
	for name, body := range rejected {
		if rec := do(t, r, http.MethodPatch, "/stocks/1", body); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: mă așteptam la 400, nu %d: %s", name, rec.Code, rec.Body)
		}
	}
	if repo.minimum != nil {
		t.Fatalf("cererile respinse nu trebuie să modifice stocul (minim %v)", *repo.minimum)
	}

	rec := do(t, r, http.MethodPatch, "/stocks/1", `{"resource_id":4,"minimum_quantity":3}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("mă așteptam la 200, nu %d: %s", rec.Code, rec.Body)
	}
	if repo.minimum == nil || *repo.minimum != 3 || repo.stock.Quantity != 10 {
		t.Errorf("trebuia schimbat doar minimul: %+v", repo.stock)
	}
}
