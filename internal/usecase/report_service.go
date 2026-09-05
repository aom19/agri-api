package usecase

import (
	"errors"
	"fmt"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

// ReportQuery sunt filtrele brute primite din query string.
type ReportQuery struct {
	From            string
	To              string
	FieldID         string
	OperationTypeID int64
	MachineID       int64
	OperatorID      int64
}

var ErrInvalidReportPeriod = errors.New("interval invalid: folosește formatul YYYY-MM-DD, cu data de început înaintea celei de sfârșit (maxim 3 ani)")

const (
	reportDateLayout      = "2006-01-02"
	reportDefaultSpanDays = 30
	reportMaxSpanDays     = 3 * 366
	reportOperationsLimit = 500
)

type ReportService struct {
	repo    repository.ReportRepository
	crops   repository.CropRepository
	weather repository.WeatherSnapshotRepository
}

func NewReportService(repo repository.ReportRepository) *ReportService {
	return &ReportService{repo: repo}
}

// resolvePeriod transformă query-ul în filtrul efectiv (interval semi-deschis) și în
// descrierea perioadei, inclusiv perioada anterioară de aceeași lungime.
func (service *ReportService) resolvePeriod(query ReportQuery) (domain.ReportFilter, domain.ReportPeriod, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)

	toDate := today
	if query.To != "" {
		parsed, err := time.ParseInLocation(reportDateLayout, query.To, time.UTC)
		if err != nil {
			return domain.ReportFilter{}, domain.ReportPeriod{}, ErrInvalidReportPeriod
		}
		toDate = parsed
	}

	fromDate := toDate.AddDate(0, 0, -(reportDefaultSpanDays - 1))
	if query.From != "" {
		parsed, err := time.ParseInLocation(reportDateLayout, query.From, time.UTC)
		if err != nil {
			return domain.ReportFilter{}, domain.ReportPeriod{}, ErrInvalidReportPeriod
		}
		fromDate = parsed
	}

	if fromDate.After(toDate) {
		return domain.ReportFilter{}, domain.ReportPeriod{}, ErrInvalidReportPeriod
	}
	spanDays := int(toDate.Sub(fromDate).Hours()/24) + 1
	if spanDays > reportMaxSpanDays {
		return domain.ReportFilter{}, domain.ReportPeriod{}, ErrInvalidReportPeriod
	}

	filter := domain.ReportFilter{
		From:            fromDate,
		To:              toDate.AddDate(0, 0, 1),
		FieldID:         query.FieldID,
		OperationTypeID: query.OperationTypeID,
		MachineID:       query.MachineID,
		OperatorID:      query.OperatorID,
	}
	previousFrom := fromDate.AddDate(0, 0, -spanDays)
	period := domain.ReportPeriod{
		From:         fromDate.Format(reportDateLayout),
		To:           toDate.Format(reportDateLayout),
		PreviousFrom: previousFrom.Format(reportDateLayout),
		PreviousTo:   fromDate.AddDate(0, 0, -1).Format(reportDateLayout),
		Granularity:  granularityFor(spanDays),
	}
	return filter, period, nil
}

func granularityFor(spanDays int) string {
	switch {
	case spanDays <= 31:
		return "day"
	case spanDays <= 183:
		return "week"
	default:
		return "month"
	}
}

func previousFilter(filter domain.ReportFilter) domain.ReportFilter {
	span := filter.To.Sub(filter.From)
	previous := filter
	previous.To = filter.From
	previous.From = filter.From.Add(-span)
	return previous
}

func (service *ReportService) GetSummary(query ReportQuery) (*domain.ReportSummary, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	current, err := service.repo.GetOperationsMetrics(filter)
	if err != nil {
		return nil, fmt.Errorf("metrici curente: %w", err)
	}
	previous, err := service.repo.GetOperationsMetrics(previousFilter(filter))
	if err != nil {
		return nil, fmt.Errorf("metrici anterioare: %w", err)
	}
	inventory, err := service.repo.GetInventorySnapshot()
	if err != nil {
		return nil, fmt.Errorf("inventar: %w", err)
	}

	return &domain.ReportSummary{
		Period:    period,
		Current:   *current,
		Previous:  *previous,
		Inventory: *inventory,
	}, nil
}

func (service *ReportService) GetOperations(query ReportQuery) (*domain.ReportOperations, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	metrics, err := service.repo.GetOperationsMetrics(filter)
	if err != nil {
		return nil, fmt.Errorf("metrici: %w", err)
	}
	timeline, err := service.repo.GetOperationsTimeline(filter, period.Granularity)
	if err != nil {
		return nil, fmt.Errorf("serie temporală: %w", err)
	}
	byType, err := service.repo.GetOperationsByType(filter)
	if err != nil {
		return nil, fmt.Errorf("pe tip: %w", err)
	}
	items, err := service.repo.GetOperationRows(filter, reportOperationsLimit)
	if err != nil {
		return nil, fmt.Errorf("lista operațiuni: %w", err)
	}

	return &domain.ReportOperations{
		Period:   period,
		Metrics:  *metrics,
		Timeline: fillTimeline(timeline, filter.From, filter.To, period.Granularity),
		ByType:   byType,
		ByStatus: []domain.ReportNamedCount{
			{Key: string(domain.FieldOperationStatusPlanned), Count: metrics.OperationsPlanned},
			{Key: string(domain.FieldOperationStatusInProgress), Count: metrics.OperationsInProgress},
			{Key: string(domain.FieldOperationStatusCompleted), Count: metrics.OperationsCompleted},
			{Key: string(domain.FieldOperationStatusCanceled), Count: metrics.OperationsCanceled},
		},
		Items: items,
	}, nil
}

// fillTimeline completează intervalele fără operațiuni cu valori zero, ca axa
// temporală să fie continuă.
func fillTimeline(buckets []domain.ReportTimeBucket, from, to time.Time, granularity string) []domain.ReportTimeBucket {
	byKey := make(map[string]domain.ReportTimeBucket, len(buckets))
	for _, bucket := range buckets {
		byKey[bucket.Bucket] = bucket
	}

	result := make([]domain.ReportTimeBucket, 0, len(buckets))
	for cursor := truncateTo(from, granularity); cursor.Before(to); cursor = advance(cursor, granularity) {
		key := cursor.Format(reportDateLayout)
		if bucket, ok := byKey[key]; ok {
			result = append(result, bucket)
			continue
		}
		result = append(result, domain.ReportTimeBucket{Bucket: key})
	}
	return result
}

func truncateTo(value time.Time, granularity string) time.Time {
	day := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	switch granularity {
	case "week":
		weekday := int(day.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		return day.AddDate(0, 0, -(weekday - 1))
	case "month":
		return time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return day
	}
}

func advance(value time.Time, granularity string) time.Time {
	switch granularity {
	case "week":
		return value.AddDate(0, 0, 7)
	case "month":
		return value.AddDate(0, 1, 0)
	default:
		return value.AddDate(0, 0, 1)
	}
}

func (service *ReportService) GetFields(query ReportQuery) (*domain.ReportFields, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	items, err := service.repo.GetFieldRows(filter)
	if err != nil {
		return nil, err
	}

	report := &domain.ReportFields{Period: period, Items: items, TotalFields: len(items)}
	for _, item := range items {
		area := 0.0
		if item.AreaHa != nil {
			area = *item.AreaHa
		}
		report.TotalAreaHa += area
		if item.OperationsCount > 0 {
			report.FieldsWithOperations++
			report.WorkedAreaHa += area
		}
	}
	return report, nil
}

func (service *ReportService) GetFleet(query ReportQuery) (*domain.ReportFleet, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	report := &domain.ReportFleet{Period: period}
	if report.MachineStatus, err = service.repo.GetMachineStatusCounts(); err != nil {
		return nil, fmt.Errorf("status mașini: %w", err)
	}
	if report.ImplementStatus, err = service.repo.GetImplementStatusCounts(); err != nil {
		return nil, fmt.Errorf("status echipamente: %w", err)
	}
	if report.MachinesByType, err = service.repo.GetMachinesByType(); err != nil {
		return nil, fmt.Errorf("mașini pe tip: %w", err)
	}
	if report.MachinesByFuel, err = service.repo.GetMachinesByFuel(); err != nil {
		return nil, fmt.Errorf("mașini pe combustibil: %w", err)
	}
	if report.MachinesByYear, err = service.repo.GetMachinesByYear(); err != nil {
		return nil, fmt.Errorf("mașini pe an: %w", err)
	}
	if report.Machines, err = service.repo.GetMachineRows(filter); err != nil {
		return nil, fmt.Errorf("lista mașini: %w", err)
	}
	if report.Implements, err = service.repo.GetImplementRows(filter); err != nil {
		return nil, fmt.Errorf("lista echipamente: %w", err)
	}
	return report, nil
}

func (service *ReportService) GetOperators(query ReportQuery) (*domain.ReportOperators, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	items, err := service.repo.GetOperatorRows(filter)
	if err != nil {
		return nil, err
	}

	report := &domain.ReportOperators{Period: period, Items: items, TotalOperators: len(items)}
	for _, item := range items {
		if item.Status == string(domain.OperatorStatusActive) {
			report.ActiveOperators++
		}
	}
	return report, nil
}

func (service *ReportService) GetStocks(query ReportQuery) (*domain.ReportStocks, error) {
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	items, err := service.repo.GetStockRows()
	if err != nil {
		return nil, fmt.Errorf("stocuri: %w", err)
	}
	consumption, err := service.repo.GetEstimatedConsumption(filter)
	if err != nil {
		return nil, fmt.Errorf("consum estimat: %w", err)
	}
	realConsumption, err := service.repo.GetRealConsumption(filter)
	if err != nil {
		return nil, fmt.Errorf("consum real: %w", err)
	}
	movementTotals, err := service.repo.GetMovementTotals(filter)
	if err != nil {
		return nil, fmt.Errorf("mișcări de stoc: %w", err)
	}

	report := &domain.ReportStocks{
		Period:               period,
		Items:                items,
		EstimatedConsumption: consumption,
		RealConsumption:      realConsumption,
		MovementTotals:       *movementTotals,
		TotalStocks:          len(items),
		ValueByCategory:      []domain.ReportNamedValue{},
	}
	for _, item := range realConsumption {
		report.RealConsumptionCost += item.Cost
	}
	valueByCategory := map[string]float64{}
	categoryOrder := []string{}
	for _, item := range items {
		report.TotalValue += item.Value
		if item.BelowMinimum {
			report.LowStocks++
		}
		if _, seen := valueByCategory[item.Category]; !seen {
			categoryOrder = append(categoryOrder, item.Category)
		}
		valueByCategory[item.Category] += item.Value
	}
	for _, category := range categoryOrder {
		report.ValueByCategory = append(report.ValueByCategory, domain.ReportNamedValue{
			Key:   category,
			Value: valueByCategory[category],
		})
	}
	return report, nil
}

// WithCrops atașează sursa de date pentru raportul pe culturi.
func (service *ReportService) WithCrops(repo repository.CropRepository) *ReportService {
	service.crops = repo
	return service
}

// WithWeather atașează sursa de date pentru raportul meteo.
func (service *ReportService) WithWeather(repo repository.WeatherSnapshotRepository) *ReportService {
	service.weather = repo
	return service
}

// GetCrops returnează raportul pe culturi pentru sezonul cerut (implicit cel activ sau cel mai recent).
func (service *ReportService) GetCrops(seasonID int64, fieldID string) (*domain.ReportCrops, error) {
	if service.crops == nil {
		return nil, errors.New("raportul pe culturi nu este configurat")
	}
	seasons, err := service.crops.GetSeasons()
	if err != nil {
		return nil, err
	}
	report := &domain.ReportCrops{Seasons: seasons, ByCrop: []domain.ReportCropStat{}, Items: []domain.ReportFieldCropRow{}}

	var season *domain.Season
	if seasonID > 0 {
		if season, err = service.crops.GetSeasonByID(seasonID); err != nil {
			return nil, err
		}
	} else if season, err = service.crops.GetActiveOrLatestSeason(); err != nil {
		return nil, err
	}
	if season == nil {
		return report, nil
	}
	report.Season = season

	items, err := service.repo.GetFieldCropRows(season.ID, fieldID)
	if err != nil {
		return nil, err
	}
	report.Items = items

	byCrop := map[int64]*domain.ReportCropStat{}
	order := []int64{}
	expectedSum := map[int64]float64{}
	expectedCount := map[int64]int{}
	for _, item := range items {
		stat, ok := byCrop[item.CropID]
		if !ok {
			stat = &domain.ReportCropStat{CropID: item.CropID, CropName: item.CropName, YieldUnit: item.YieldUnit}
			byCrop[item.CropID] = stat
			order = append(order, item.CropID)
		}
		stat.FieldsCount++
		report.Totals.FieldsCount++
		if item.PlantedAreaHa != nil {
			stat.PlantedAreaHa += *item.PlantedAreaHa
			report.Totals.PlantedAreaHa += *item.PlantedAreaHa
		}
		if item.ProductionTotal != nil {
			stat.ProductionTotal += *item.ProductionTotal
			report.Totals.ProductionTotal += *item.ProductionTotal
		}
		if item.ExpectedYieldPerHa != nil {
			expectedSum[item.CropID] += *item.ExpectedYieldPerHa
			expectedCount[item.CropID]++
		}
		stat.EstimatedCost += item.EstimatedCost
		stat.RealCost += item.RealCost
		report.Totals.EstimatedCost += item.EstimatedCost
		report.Totals.RealCost += item.RealCost
	}
	for _, cropID := range order {
		stat := byCrop[cropID]
		if stat.PlantedAreaHa > 0 {
			if stat.ProductionTotal > 0 {
				yield := stat.ProductionTotal / stat.PlantedAreaHa
				stat.YieldPerHa = &yield
			}
			cost := stat.RealCost
			if cost <= 0 {
				cost = stat.EstimatedCost
			}
			if cost > 0 {
				perHa := cost / stat.PlantedAreaHa
				stat.CostPerHa = &perHa
			}
		}
		if count := expectedCount[cropID]; count > 0 {
			expected := expectedSum[cropID] / float64(count)
			stat.ExpectedYieldPerHa = &expected
		}
		report.ByCrop = append(report.ByCrop, *stat)
	}
	if report.Totals.PlantedAreaHa > 0 && report.Totals.ProductionTotal > 0 {
		yield := report.Totals.ProductionTotal / report.Totals.PlantedAreaHa
		report.Totals.YieldPerHa = &yield
	}
	return report, nil
}

// GetWeather returnează istoricul meteo agregat pe zile pentru intervalul și terenul filtrat.
func (service *ReportService) GetWeather(query ReportQuery) (*domain.ReportWeather, error) {
	if service.weather == nil {
		return nil, errors.New("raportul meteo nu este configurat")
	}
	filter, period, err := service.resolvePeriod(query)
	if err != nil {
		return nil, err
	}

	series, err := service.weather.GetDailyAggregates(filter)
	if err != nil {
		return nil, fmt.Errorf("serie meteo: %w", err)
	}
	latest, err := service.weather.GetLatestPerField()
	if err != nil {
		return nil, fmt.Errorf("observații curente: %w", err)
	}
	if filter.FieldID != "" {
		filtered := []domain.WeatherSnapshot{}
		for _, item := range latest {
			if item.FieldID != nil && *item.FieldID == filter.FieldID {
				filtered = append(filtered, item)
			}
		}
		latest = filtered
	}
	samples, fields, err := service.weather.CountInPeriod(filter)
	if err != nil {
		return nil, fmt.Errorf("număr observații: %w", err)
	}

	report := &domain.ReportWeather{Period: period, Series: series, Latest: latest}
	report.Summary.Samples = samples
	report.Summary.FieldsCovered = fields
	if len(series) > 0 {
		report.Summary.MinTemperatureC = series[0].MinTemperatureC
		report.Summary.MaxTemperatureC = series[0].MaxTemperatureC
		sum := 0.0
		for _, day := range series {
			sum += day.AvgTemperatureC
			if day.MinTemperatureC < report.Summary.MinTemperatureC {
				report.Summary.MinTemperatureC = day.MinTemperatureC
			}
			if day.MaxTemperatureC > report.Summary.MaxTemperatureC {
				report.Summary.MaxTemperatureC = day.MaxTemperatureC
			}
			report.Summary.TotalPrecipitationMM += day.PrecipitationMM
			if day.PrecipitationMM >= 0.5 {
				report.Summary.RainyDays++
			}
		}
		report.Summary.AvgTemperatureC = sum / float64(len(series))
	}
	return report, nil
}
