package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type fieldOperationRepoStub struct {
	repository.FieldOperationRepository
}

func (fieldOperationRepoStub) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	if id != 1 {
		return nil, nil
	}
	template, area := int64(3), 10.0
	return &dto.FieldOperationResponse{ID: 1, OperationTemplateID: &template, AreaPlannedHa: &area}, nil
}
func (fieldOperationRepoStub) GetTemplateResources(int64) ([]domain.TemplateResourceUsage, error) {
	return []domain.TemplateResourceUsage{{ResourceID: 1, ResourceName: "Uree", QuantityPerUnit: 2}}, nil
}

type stockMovementRepoStub struct {
	repository.StockMovementRepository
}

func (stockMovementRepoStub) ListFuelStocks() ([]domain.FuelStock, error) {
	return []domain.FuelStock{}, nil
}

func TestFieldOperationHandler_ConsumptionEstimate(t *testing.T) {
	completion := usecase.NewFieldOperationCompletionService(nil, fieldOperationRepoStub{}, stockMovementRepoStub{})
	r := gin.New()
	r.GET("/field-operations/:id/consumption-estimate", handlers.NewFieldOperationHandler(nil, handlers.WithFieldOpCompletion(completion)).ConsumptionEstimate)

	rec := do(t, r, http.MethodGet, "/field-operations/1/consumption-estimate?area_ha=2,5", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("mă așteptam la 200, nu %d: %s", rec.Code, rec.Body)
	}
	var estimate domain.ConsumptionEstimate
	if err := json.Unmarshal(rec.Body.Bytes(), &estimate); err != nil {
		t.Fatal(err)
	}
	if estimate.AreaHa != 2.5 || len(estimate.Items) != 1 || estimate.Items[0].Quantity != 5 {
		t.Errorf("estimare: %+v", estimate)
	}

	for path, want := range map[string]int{
		"/field-operations/x/consumption-estimate":            http.StatusBadRequest,
		"/field-operations/1/consumption-estimate?area_ha=ab": http.StatusBadRequest,
		"/field-operations/1/consumption-estimate?area_ha=-1": http.StatusBadRequest,
		"/field-operations/9/consumption-estimate":            http.StatusNotFound,
	} {
		if rec := do(t, r, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("%s: mă așteptam la %d, nu %d", path, want, rec.Code)
		}
	}
}
