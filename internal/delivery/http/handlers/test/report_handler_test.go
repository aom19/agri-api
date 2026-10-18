package handlers_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// reportRepoFake întoarce rapoarte goale; err, când e setat, e întors de metrici.
type reportRepoFake struct {
	err        error
	lastFilter domain.ReportFilter
}

func (r *reportRepoFake) GetOperationsMetrics(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
	r.lastFilter = f
	if r.err != nil {
		return nil, r.err
	}
	return &domain.ReportOperationsMetrics{OperationsTotal: 3, EstimatedCost: 50}, nil
}
func (r *reportRepoFake) GetInventorySnapshot() (*domain.ReportInventorySnapshot, error) {
	return &domain.ReportInventorySnapshot{}, nil
}
func (r *reportRepoFake) GetOperationsTimeline(domain.ReportFilter, string) ([]domain.ReportTimeBucket, error) {
	return nil, nil
}
func (r *reportRepoFake) GetOperationsByType(domain.ReportFilter) ([]domain.ReportOperationTypeStat, error) {
	return nil, nil
}
func (r *reportRepoFake) GetOperationRows(domain.ReportFilter, int) ([]domain.ReportOperationRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetFieldRows(domain.ReportFilter) ([]domain.ReportFieldRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetMachineStatusCounts() ([]domain.ReportNamedCount, error) { return nil, nil }
func (r *reportRepoFake) GetImplementStatusCounts() ([]domain.ReportNamedCount, error) {
	return nil, nil
}
func (r *reportRepoFake) GetMachinesByType() ([]domain.ReportNamedCount, error) { return nil, nil }
func (r *reportRepoFake) GetMachinesByFuel() ([]domain.ReportNamedCount, error) { return nil, nil }
func (r *reportRepoFake) GetMachinesByYear() ([]domain.ReportNamedCount, error) { return nil, nil }
func (r *reportRepoFake) GetMachineRows(domain.ReportFilter) ([]domain.ReportMachineRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetImplementRows(domain.ReportFilter) ([]domain.ReportImplementRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetOperatorRows(domain.ReportFilter) ([]domain.ReportOperatorRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetStockRows() ([]domain.ReportStockRow, error) { return nil, nil }
func (r *reportRepoFake) GetEstimatedConsumption(domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	return nil, nil
}
func (r *reportRepoFake) GetRealConsumption(domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	return nil, nil
}
func (r *reportRepoFake) GetMovementTotals(domain.ReportFilter) (*domain.ReportMovementTotals, error) {
	return &domain.ReportMovementTotals{}, nil
}
func (r *reportRepoFake) GetFieldCropRows(int64, string) ([]domain.ReportFieldCropRow, error) {
	return nil, nil
}
func (r *reportRepoFake) GetCompareName(kind domain.ReportCompareKind, id int64) (string, error) {
	if id == 404 {
		return "", nil
	}
	return string(kind), nil
}

type seasonRepoStub struct {
	repository.CropRepository
}

func (seasonRepoStub) GetSeasons() ([]domain.Season, error) {
	return []domain.Season{{ID: 1, Name: "Sezon 1"}}, nil
}
func (seasonRepoStub) GetSeasonByID(id int64) (*domain.Season, error) {
	return &domain.Season{ID: id, Name: "Sezon"}, nil
}
func (seasonRepoStub) GetActiveOrLatestSeason() (*domain.Season, error) {
	return &domain.Season{ID: 1, Name: "Sezon 1"}, nil
}

func reportRouter(repo *reportRepoFake, mw gin.HandlerFunc) *gin.Engine {
	h := handlers.NewReportHandler(usecase.NewReportService(repo).WithCrops(seasonRepoStub{}))
	r := gin.New()
	r.Use(mw)
	r.GET("/summary", h.GetSummary)
	r.GET("/operations", h.GetOperations)
	r.GET("/my-operations", h.GetMyOperations)
	r.GET("/compare", h.Compare)
	r.GET("/fields", h.GetFields)
	r.GET("/fleet", h.GetFleet)
	r.GET("/operators", h.GetOperators)
	r.GET("/stocks", h.GetStocks)
	r.GET("/crops", h.GetCrops)
	r.GET("/weather", h.GetWeather)
	r.POST("/summary/email", h.SendSummaryEmail)
	return r
}

func TestReportHandler_Endpoints(t *testing.T) {
	repo := &reportRepoFake{}
	r := reportRouter(repo, asUser("1", "admin"))

	period := "?from=2026-01-01&to=2026-01-31"
	for _, path := range []string{
		"/summary", "/operations", "/my-operations", "/fields", "/fleet", "/operators", "/stocks",
		"/compare?kind=operators&a=1&b=2", "/compare?kind=seasons&a=1&b=2",
	} {
		sep := period
		if strings.Contains(path, "?") {
			sep = "&" + period[1:]
		}
		if rec := do(t, r, http.MethodGet, path+sep, ""); rec.Code != http.StatusOK {
			t.Errorf("GET %s: %d %s", path, rec.Code, rec.Body)
		}
	}
	if rec := do(t, r, http.MethodGet, "/crops?season_id=1", ""); rec.Code != http.StatusOK {
		t.Errorf("culturi: %d %s", rec.Code, rec.Body)
	}

	for path, want := range map[string]int{
		"/summary?from=ieri":                       http.StatusBadRequest,
		"/operations?operation_type=x":             http.StatusBadRequest,
		"/fields?machine_id=x":                     http.StatusBadRequest,
		"/fleet?operator_id=x":                     http.StatusBadRequest,
		"/operators?from=2026-02-01&to=2026-01-01": http.StatusBadRequest,
		"/stocks?operator_id=x":                    http.StatusBadRequest,
		"/my-operations?machine_id=x":              http.StatusBadRequest,
		"/compare?kind=operators&a=1&b=1":          http.StatusBadRequest,
		"/compare?kind=operators&a=x&b=1":          http.StatusBadRequest,
		"/compare?kind=operators&a=1&b=x":          http.StatusBadRequest,
		"/compare?kind=operators&from=x&a=1&b=2":   http.StatusBadRequest,
		"/compare?kind=machines&a=1&b=404":         http.StatusNotFound,
		"/crops?season_id=x":                       http.StatusBadRequest,
		"/weather?operation_type=x":                http.StatusBadRequest,
		"/weather":                                 http.StatusInternalServerError,
	} {
		if rec := do(t, r, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("GET %s: vreau %d, am %d", path, want, rec.Code)
		}
	}

	repo.err = errors.New("db down")
	if rec := do(t, r, http.MethodGet, "/summary", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroarea repository-ului: %d", rec.Code)
	}
}

func TestReportHandler_MyOperationsUsesCurrentUser(t *testing.T) {
	repo := &reportRepoFake{}
	rec := do(t, reportRouter(repo, asUser("7", "operator")), http.MethodGet, "/my-operations?operator_id=5", "")
	if rec.Code != http.StatusOK || repo.lastFilter.OperatorID != 7 {
		t.Fatalf("raportul propriu: %d, filtru %+v", rec.Code, repo.lastFilter)
	}
	if !strings.Contains(rec.Body.String(), `"estimated_cost":0`) {
		t.Errorf("costurile nu sunt trimise operatorului: %s", rec.Body)
	}
	if rec := do(t, reportRouter(repo, asUser("x", "operator")), http.MethodGet, "/my-operations", ""); rec.Code != http.StatusUnauthorized {
		t.Errorf("fără cont valid: %d", rec.Code)
	}
}

func TestReportHandler_EmailNotConfigured(t *testing.T) {
	rec := do(t, reportRouter(&reportRepoFake{}, asUser("1", "admin")), http.MethodPost, "/summary/email", "")
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("fără serviciul de e-mail: %d", rec.Code)
	}
}
