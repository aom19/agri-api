package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func TestReportService_CompareOperators(t *testing.T) {
	var operators []int64
	repo := &reportRepoMock{
		getOperationsMetrics: func(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
			operators = append(operators, f.OperatorID)
			return &domain.ReportOperationsMetrics{OperationsTotal: int(f.OperatorID)}, nil
		},
	}
	svc := usecase.NewReportService(repo)

	report, err := svc.Compare(usecase.ReportCompareQuery{
		ReportQuery: usecase.ReportQuery{From: "2026-01-01", To: "2026-01-03"},
		Kind:        domain.ReportCompareOperators, A: 2, B: 5,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.A.Metrics.OperationsTotal != 2 || report.B.Metrics.OperationsTotal != 5 || len(operators) != 2 {
		t.Errorf("fiecare parte e filtrată pe operatorul ei: %v, %+v", operators, report)
	}
	if report.Period == nil || len(report.A.Timeline) != 3 || report.A.Name != "operators #2" {
		t.Errorf("perioada, seria și numele: %+v", report)
	}
}

func TestReportService_CompareValidation(t *testing.T) {
	repo := &reportRepoMock{getCompareName: func(domain.ReportCompareKind, int64) (string, error) { return "", nil }}
	svc := usecase.NewReportService(repo)

	for _, q := range []usecase.ReportCompareQuery{
		{Kind: "fields", A: 1, B: 2},
		{Kind: domain.ReportCompareMachines, A: 1, B: 1},
		{Kind: domain.ReportCompareMachines, A: 0, B: 2},
	} {
		if _, err := svc.Compare(q); !errors.Is(err, usecase.ErrInvalidCompare) {
			t.Errorf("%+v: %v", q, err)
		}
	}
	if _, err := svc.Compare(usecase.ReportCompareQuery{Kind: domain.ReportCompareMachines, A: 1, B: 2}); !errors.Is(err, usecase.ErrCompareNotFound) {
		t.Errorf("mașina inexistentă: %v", err)
	}
}

func TestReportService_CompareSeasons(t *testing.T) {
	var filters []domain.ReportFilter
	repo := &reportRepoMock{
		getOperationsMetrics: func(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
			filters = append(filters, f)
			return &domain.ReportOperationsMetrics{}, nil
		},
		getFieldCropRows: func(seasonID int64, _ string) ([]domain.ReportFieldCropRow, error) {
			area := float64(seasonID * 10)
			return []domain.ReportFieldCropRow{{FieldCrop: domain.FieldCrop{CropID: 1, CropName: "Grâu", PlantedAreaHa: &area}}}, nil
		},
	}
	crops := &cropRepoMock{getSeasonByID: func(id int64) (*domain.Season, error) {
		if id == 9 {
			return nil, nil
		}
		return &domain.Season{ID: id, Name: "Sezon"}, nil
	}}
	svc := usecase.NewReportService(repo).WithCrops(crops)

	report, err := svc.Compare(usecase.ReportCompareQuery{Kind: domain.ReportCompareSeasons, A: 1, B: 2})
	if err != nil {
		t.Fatal(err)
	}
	if report.Period != nil || report.A.Season == nil || report.B.CropTotals.PlantedAreaHa != 20 {
		t.Errorf("sezonul: fără perioadă, cu culturile lui: %+v", report)
	}
	for _, f := range filters {
		if !f.From.IsZero() || f.SeasonID == 0 {
			t.Errorf("lucrările sezonului, fără filtru de dată: %+v", f)
		}
	}
	if _, err := svc.Compare(usecase.ReportCompareQuery{Kind: domain.ReportCompareSeasons, A: 1, B: 9}); !errors.Is(err, usecase.ErrCompareNotFound) {
		t.Errorf("sezonul inexistent: %v", err)
	}
}
