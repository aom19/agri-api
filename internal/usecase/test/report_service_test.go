package usecase_test

import (
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func TestReportService_ResolvePeriod(t *testing.T) {
	svc := usecase.NewReportService(&reportRepoMock{})

	filter, period, err := svc.ResolvePeriod(usecase.ReportQuery{From: "2026-01-01", To: "2026-01-31", FieldID: "f1", MachineID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if period.Granularity != "day" || period.PreviousFrom != "2025-12-01" || period.PreviousTo != "2025-12-31" {
		t.Errorf("perioadă greșită: %+v", period)
	}
	if filter.To.Format(usecase.ReportDateLayout) != "2026-02-01" || filter.FieldID != "f1" || filter.MachineID != 2 {
		t.Errorf("filtru greșit: %+v", filter)
	}

	if _, period, _ = svc.ResolvePeriod(usecase.ReportQuery{From: "2026-01-01", To: "2026-04-30"}); period.Granularity != "week" {
		t.Errorf("120 de zile → week, am %s", period.Granularity)
	}
	if _, period, _ = svc.ResolvePeriod(usecase.ReportQuery{From: "2025-01-01", To: "2025-12-31"}); period.Granularity != "month" {
		t.Errorf("un an → month, am %s", period.Granularity)
	}

	_, period, err = svc.ResolvePeriod(usecase.ReportQuery{})
	today := time.Now().UTC().Format(usecase.ReportDateLayout)
	if err != nil || period.To != today {
		t.Errorf("perioada implicită se termină azi: %v, %+v", err, period)
	}

	invalid := []usecase.ReportQuery{
		{To: "azi"}, {From: "ieri"},
		{From: "2026-02-01", To: "2026-01-01"},
		{From: "2020-01-01", To: "2026-01-01"},
	}
	for _, q := range invalid {
		if _, _, err := svc.ResolvePeriod(q); !errors.Is(err, usecase.ErrInvalidReportPeriod) {
			t.Errorf("%+v: %v", q, err)
		}
	}
}

func TestFillTimeline(t *testing.T) {
	day := func(s string) time.Time {
		v, _ := time.ParseInLocation(usecase.ReportDateLayout, s, time.UTC)
		return v
	}
	result := usecase.FillTimeline([]domain.ReportTimeBucket{{Bucket: "2026-01-02", Planned: 2}}, day("2026-01-01"), day("2026-01-04"), "day")
	if len(result) != 3 || result[1].Planned != 2 || result[0].Bucket != "2026-01-01" || result[2].Planned != 0 {
		t.Errorf("zi: %+v", result)
	}
	// 2026-01-07 e miercuri → săptămâna începe luni 2026-01-05
	result = usecase.FillTimeline(nil, day("2026-01-07"), day("2026-01-21"), "week")
	if len(result) != 3 || result[0].Bucket != "2026-01-05" || result[2].Bucket != "2026-01-19" {
		t.Errorf("săptămână: %+v", result)
	}
	// duminică → luni anterioară
	if got := usecase.TruncateTo(day("2026-01-11"), "week").Format(usecase.ReportDateLayout); got != "2026-01-05" {
		t.Errorf("duminica trebuie trunchiată la luni: %s", got)
	}
	result = usecase.FillTimeline(nil, day("2026-01-15"), day("2026-03-01"), "month")
	if len(result) != 2 || result[0].Bucket != "2026-01-01" || result[1].Bucket != "2026-02-01" {
		t.Errorf("lună: %+v", result)
	}
}

func TestReportService_GetSummary(t *testing.T) {
	var filters []domain.ReportFilter
	repo := &reportRepoMock{getOperationsMetrics: func(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
		filters = append(filters, f)
		return &domain.ReportOperationsMetrics{OperationsTotal: len(filters)}, nil
	}}
	svc := usecase.NewReportService(repo)

	summary, err := svc.GetSummary(usecase.ReportQuery{From: "2026-01-01", To: "2026-01-10"})
	if err != nil || summary.Current.OperationsTotal != 1 || summary.Previous.OperationsTotal != 2 {
		t.Fatalf("GetSummary: %v, %+v", err, summary)
	}
	if !filters[1].To.Equal(filters[0].From) || filters[0].From.Sub(filters[1].From) != 10*24*time.Hour {
		t.Errorf("perioada anterioară greșită: %+v", filters[1])
	}

	if _, err := svc.GetSummary(usecase.ReportQuery{To: "x"}); err == nil {
		t.Error("perioada invalidă trebuie să dea eroare")
	}
	repo.getInventorySnapshot = func() (*domain.ReportInventorySnapshot, error) { return nil, errors.New("db down") }
	if _, err := svc.GetSummary(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea de inventar trebuie propagată")
	}
	calls := 0
	repo.getOperationsMetrics = func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
		calls++
		if calls == 2 {
			return nil, errors.New("db down")
		}
		return &domain.ReportOperationsMetrics{}, nil
	}
	if _, err := svc.GetSummary(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea metricilor anterioare trebuie propagată")
	}
	repo.getOperationsMetrics = func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) { return nil, errors.New("db down") }
	if _, err := svc.GetSummary(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea metricilor curente trebuie propagată")
	}
}

func TestReportService_GetOperations(t *testing.T) {
	repo := &reportRepoMock{
		getOperationsMetrics: func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
			return &domain.ReportOperationsMetrics{OperationsPlanned: 1, OperationsInProgress: 2, OperationsCompleted: 3, OperationsCanceled: 4}, nil
		},
		getOperationsTimeline: func(_ domain.ReportFilter, g string) ([]domain.ReportTimeBucket, error) {
			return []domain.ReportTimeBucket{{Bucket: "2026-01-02", Completed: 1}}, nil
		},
	}
	svc := usecase.NewReportService(repo)

	report, err := svc.GetOperations(usecase.ReportQuery{From: "2026-01-01", To: "2026-01-03"})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Timeline) != 3 || report.Timeline[1].Completed != 1 {
		t.Errorf("timeline: %+v", report.Timeline)
	}
	if len(report.ByStatus) != 4 || report.ByStatus[3].Key != "canceled" || report.ByStatus[3].Count != 4 {
		t.Errorf("by_status: %+v", report.ByStatus)
	}

	if _, err := svc.GetOperations(usecase.ReportQuery{To: "x"}); err == nil {
		t.Error("perioada invalidă trebuie să dea eroare")
	}
	boom := errors.New("db down")
	repo.getOperationRows = func(domain.ReportFilter, int) ([]domain.ReportOperationRow, error) { return nil, boom }
	if _, err := svc.GetOperations(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea listei trebuie propagată")
	}
	repo.getOperationsByType = func(domain.ReportFilter) ([]domain.ReportOperationTypeStat, error) { return nil, boom }
	if _, err := svc.GetOperations(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea pe tip trebuie propagată")
	}
	repo.getOperationsTimeline = func(domain.ReportFilter, string) ([]domain.ReportTimeBucket, error) { return nil, boom }
	if _, err := svc.GetOperations(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea seriei trebuie propagată")
	}
	repo.getOperationsMetrics = func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) { return nil, boom }
	if _, err := svc.GetOperations(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea metricilor trebuie propagată")
	}
}

func TestReportService_FieldsFleetOperators(t *testing.T) {
	repo := &reportRepoMock{
		getFieldRows: func(domain.ReportFilter) ([]domain.ReportFieldRow, error) {
			return []domain.ReportFieldRow{
				{ID: "a", AreaHa: ptr(10.0), OperationsCount: 2},
				{ID: "b", AreaHa: ptr(5.0)},
				{ID: "c", OperationsCount: 1},
			}, nil
		},
		getOperatorRows: func(domain.ReportFilter) ([]domain.ReportOperatorRow, error) {
			return []domain.ReportOperatorRow{{Status: "active"}, {Status: "inactive"}, {Status: "active"}}, nil
		},
	}
	svc := usecase.NewReportService(repo)

	fields, err := svc.GetFields(usecase.ReportQuery{})
	if err != nil || fields.TotalFields != 3 || fields.TotalAreaHa != 15 || fields.FieldsWithOperations != 2 || fields.WorkedAreaHa != 10 {
		t.Errorf("GetFields: %v, %+v", err, fields)
	}
	operators, err := svc.GetOperators(usecase.ReportQuery{})
	if err != nil || operators.TotalOperators != 3 || operators.ActiveOperators != 2 {
		t.Errorf("GetOperators: %v, %+v", err, operators)
	}
	fleet, err := svc.GetFleet(usecase.ReportQuery{})
	if err != nil || fleet == nil {
		t.Errorf("GetFleet: %v", err)
	}

	for _, q := range []usecase.ReportQuery{{To: "x"}} {
		if _, err := svc.GetFields(q); err == nil {
			t.Error("GetFields: perioada invalidă trebuie să dea eroare")
		}
		if _, err := svc.GetOperators(q); err == nil {
			t.Error("GetOperators: perioada invalidă trebuie să dea eroare")
		}
		if _, err := svc.GetFleet(q); err == nil {
			t.Error("GetFleet: perioada invalidă trebuie să dea eroare")
		}
	}

	boom := errors.New("db down")
	repo.getFieldRows = func(domain.ReportFilter) ([]domain.ReportFieldRow, error) { return nil, boom }
	if _, err := svc.GetFields(usecase.ReportQuery{}); !errors.Is(err, boom) {
		t.Error("GetFields: eroarea trebuie propagată")
	}
	repo.getOperatorRows = func(domain.ReportFilter) ([]domain.ReportOperatorRow, error) { return nil, boom }
	if _, err := svc.GetOperators(usecase.ReportQuery{}); !errors.Is(err, boom) {
		t.Error("GetOperators: eroarea trebuie propagată")
	}

	// fiecare sursă a raportului de flotă poate eșua
	counts := func() ([]domain.ReportNamedCount, error) { return nil, boom }
	failures := []func(){
		func() {
			repo.getImplementRows = func(domain.ReportFilter) ([]domain.ReportImplementRow, error) { return nil, boom }
		},
		func() {
			repo.getMachineRows = func(domain.ReportFilter) ([]domain.ReportMachineRow, error) { return nil, boom }
		},
		func() { repo.getMachinesByYear = counts },
		func() { repo.getMachinesByFuel = counts },
		func() { repo.getMachinesByType = counts },
		func() { repo.getImplementStatusCounts = counts },
		func() { repo.getMachineStatusCounts = counts },
	}
	for i, fail := range failures {
		fail()
		if _, err := svc.GetFleet(usecase.ReportQuery{}); !errors.Is(err, boom) {
			t.Errorf("GetFleet: eroarea #%d trebuie propagată: %v", i, err)
		}
	}
}

func TestReportService_GetStocks(t *testing.T) {
	repo := &reportRepoMock{
		getStockRows: func() ([]domain.ReportStockRow, error) {
			return []domain.ReportStockRow{
				{Category: "fuel", Value: 100, BelowMinimum: true},
				{Category: "seed", Value: 50},
				{Category: "fuel", Value: 25},
			}, nil
		},
		getRealConsumption: func(domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
			return []domain.ReportResourceConsumption{{Cost: 10}, {Cost: 5.5}}, nil
		},
	}
	svc := usecase.NewReportService(repo)

	report, err := svc.GetStocks(usecase.ReportQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if report.TotalStocks != 3 || report.LowStocks != 1 || report.TotalValue != 175 || report.RealConsumptionCost != 15.5 {
		t.Errorf("totaluri: %+v", report)
	}
	if len(report.ValueByCategory) != 2 || report.ValueByCategory[0].Key != "fuel" || report.ValueByCategory[0].Value != 125 {
		t.Errorf("valoare pe categorie: %+v", report.ValueByCategory)
	}

	if _, err := svc.GetStocks(usecase.ReportQuery{To: "x"}); err == nil {
		t.Error("perioada invalidă trebuie să dea eroare")
	}
	boom := errors.New("db down")
	repo.getMovementTotals = func(domain.ReportFilter) (*domain.ReportMovementTotals, error) { return nil, boom }
	if _, err := svc.GetStocks(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea mișcărilor trebuie propagată")
	}
	repo.getRealConsumption = func(domain.ReportFilter) ([]domain.ReportResourceConsumption, error) { return nil, boom }
	if _, err := svc.GetStocks(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea consumului real trebuie propagată")
	}
	repo.getEstimatedConsumption = func(domain.ReportFilter) ([]domain.ReportResourceConsumption, error) { return nil, boom }
	if _, err := svc.GetStocks(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea consumului estimat trebuie propagată")
	}
	repo.getStockRows = func() ([]domain.ReportStockRow, error) { return nil, boom }
	if _, err := svc.GetStocks(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea stocurilor trebuie propagată")
	}
}

func TestReportService_GetCrops(t *testing.T) {
	svc := usecase.NewReportService(&reportRepoMock{})
	if _, err := svc.GetCrops(0, ""); err == nil {
		t.Error("fără sursă de culturi trebuie să dea eroare")
	}

	crops := &cropRepoMock{
		getSeasons:              func() ([]domain.Season, error) { return []domain.Season{{ID: 1}, {ID: 2}}, nil },
		getSeasonByID:           func(id int64) (*domain.Season, error) { return &domain.Season{ID: id}, nil },
		getActiveOrLatestSeason: func() (*domain.Season, error) { return &domain.Season{ID: 2}, nil },
	}
	repo := &reportRepoMock{getFieldCropRows: func(seasonID int64, fieldID string) ([]domain.ReportFieldCropRow, error) {
		return []domain.ReportFieldCropRow{
			{FieldCrop: domain.FieldCrop{CropID: 1, CropName: "Grâu", PlantedAreaHa: ptr(10.0), ProductionTotal: ptr(40.0), ExpectedYieldPerHa: ptr(4.0)}, EstimatedCost: 100, RealCost: 80},
			{FieldCrop: domain.FieldCrop{CropID: 1, CropName: "Grâu", PlantedAreaHa: ptr(10.0), ProductionTotal: ptr(60.0), ExpectedYieldPerHa: ptr(6.0)}, EstimatedCost: 100},
			{FieldCrop: domain.FieldCrop{CropID: 2, CropName: "Porumb", PlantedAreaHa: ptr(5.0)}, EstimatedCost: 50},
			{FieldCrop: domain.FieldCrop{CropID: 3, CropName: "Fără suprafață"}},
		}, nil
	}}
	svc = usecase.NewReportService(repo).WithCrops(crops)

	report, err := svc.GetCrops(1, "f1")
	if err != nil || report.Season.ID != 1 || len(report.Seasons) != 2 {
		t.Fatalf("GetCrops: %v, %+v", err, report)
	}
	if len(report.ByCrop) != 3 || report.Totals.FieldsCount != 4 || report.Totals.PlantedAreaHa != 25 || report.Totals.ProductionTotal != 100 {
		t.Errorf("agregare: %+v", report.Totals)
	}
	wheat := report.ByCrop[0]
	if wheat.FieldsCount != 2 || *wheat.YieldPerHa != 5 || *wheat.ExpectedYieldPerHa != 5 || *wheat.CostPerHa != 4 {
		t.Errorf("grâu: %+v", wheat)
	}
	corn := report.ByCrop[1]
	if corn.YieldPerHa != nil || *corn.CostPerHa != 10 {
		t.Errorf("porumb (cost estimat când nu există cost real): %+v", corn)
	}
	if report.ByCrop[2].CostPerHa != nil || *report.Totals.YieldPerHa != 4 {
		t.Errorf("fără suprafață / total: %+v, %v", report.ByCrop[2], report.Totals.YieldPerHa)
	}

	if report, err = svc.GetCrops(0, ""); err != nil || report.Season.ID != 2 {
		t.Errorf("sezonul activ implicit: %v", err)
	}
	crops.getActiveOrLatestSeason = func() (*domain.Season, error) { return nil, nil }
	if report, err = svc.GetCrops(0, ""); err != nil || report.Season != nil || len(report.Items) != 0 {
		t.Errorf("fără sezon: %v, %+v", err, report)
	}

	boom := errors.New("db down")
	repo.getFieldCropRows = func(int64, string) ([]domain.ReportFieldCropRow, error) { return nil, boom }
	if _, err := svc.GetCrops(1, ""); !errors.Is(err, boom) {
		t.Error("eroarea rândurilor trebuie propagată")
	}
	crops.getSeasonByID = func(int64) (*domain.Season, error) { return nil, boom }
	if _, err := svc.GetCrops(1, ""); !errors.Is(err, boom) {
		t.Error("eroarea sezonului trebuie propagată")
	}
	crops.getActiveOrLatestSeason = func() (*domain.Season, error) { return nil, boom }
	if _, err := svc.GetCrops(0, ""); !errors.Is(err, boom) {
		t.Error("eroarea sezonului activ trebuie propagată")
	}
	crops.getSeasons = func() ([]domain.Season, error) { return nil, boom }
	if _, err := svc.GetCrops(0, ""); !errors.Is(err, boom) {
		t.Error("eroarea sezoanelor trebuie propagată")
	}
}

func TestReportService_GetWeather(t *testing.T) {
	svc := usecase.NewReportService(&reportRepoMock{})
	if _, err := svc.GetWeather(usecase.ReportQuery{}); err == nil {
		t.Error("fără sursă meteo trebuie să dea eroare")
	}

	weather := &weatherSnapshotRepoMock{
		getDailyAggregates: func(domain.ReportFilter) ([]domain.WeatherDailyAggregate, error) {
			return []domain.WeatherDailyAggregate{
				{Day: "2026-01-01", AvgTemperatureC: 10, MinTemperatureC: 5, MaxTemperatureC: 15, PrecipitationMM: 1},
				{Day: "2026-01-02", AvgTemperatureC: 20, MinTemperatureC: 2, MaxTemperatureC: 25, PrecipitationMM: 0.2},
			}, nil
		},
		getLatestPerField: func() ([]domain.WeatherSnapshot, error) {
			return []domain.WeatherSnapshot{{FieldID: ptr("f1")}, {FieldID: ptr("f2")}, {}}, nil
		},
		countInPeriod: func(domain.ReportFilter) (int, int, error) { return 48, 2, nil },
	}
	svc = usecase.NewReportService(&reportRepoMock{}).WithWeather(weather)

	report, err := svc.GetWeather(usecase.ReportQuery{FieldID: "f1"})
	if err != nil {
		t.Fatal(err)
	}
	s := report.Summary
	if s.AvgTemperatureC != 15 || s.MinTemperatureC != 2 || s.MaxTemperatureC != 25 || s.TotalPrecipitationMM != 1.2 || s.RainyDays != 1 || s.Samples != 48 || s.FieldsCovered != 2 {
		t.Errorf("sumar: %+v", s)
	}
	if len(report.Latest) != 1 {
		t.Errorf("observațiile curente trebuie filtrate pe teren: %+v", report.Latest)
	}
	if report, err = svc.GetWeather(usecase.ReportQuery{}); err != nil || len(report.Latest) != 3 {
		t.Errorf("fără filtru pe teren: %v", err)
	}

	if _, err := svc.GetWeather(usecase.ReportQuery{To: "x"}); err == nil {
		t.Error("perioada invalidă trebuie să dea eroare")
	}
	boom := errors.New("db down")
	weather.countInPeriod = func(domain.ReportFilter) (int, int, error) { return 0, 0, boom }
	if _, err := svc.GetWeather(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea numărului de observații trebuie propagată")
	}
	weather.getLatestPerField = func() ([]domain.WeatherSnapshot, error) { return nil, boom }
	if _, err := svc.GetWeather(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea observațiilor curente trebuie propagată")
	}
	weather.getDailyAggregates = func(domain.ReportFilter) ([]domain.WeatherDailyAggregate, error) { return nil, boom }
	if _, err := svc.GetWeather(usecase.ReportQuery{}); err == nil {
		t.Error("eroarea seriei trebuie propagată")
	}
}
