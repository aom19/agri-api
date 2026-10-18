package usecase

import (
	"errors"
	"fmt"

	"agri-api/internal/domain"
)

var (
	ErrInvalidCompare  = errors.New("comparația cere un tip valid (operators, machines, seasons) și două entități diferite")
	ErrCompareNotFound = errors.New("una dintre entitățile comparate nu există")
)

// ReportCompareQuery: tipul, cele două entități și filtrele comune (perioada nu se aplică sezoanelor).
type ReportCompareQuery struct {
	ReportQuery
	Kind domain.ReportCompareKind
	A, B int64
}

// Compare pune față în față doi operatori, două mașini sau două sezoane, cu aceiași indicatori.
func (service *ReportService) Compare(query ReportCompareQuery) (*domain.ReportCompare, error) {
	if !query.Kind.IsValid() || query.A <= 0 || query.B <= 0 || query.A == query.B {
		return nil, ErrInvalidCompare
	}
	result := &domain.ReportCompare{Kind: query.Kind}

	var period domain.ReportPeriod
	var filter domain.ReportFilter
	if query.Kind != domain.ReportCompareSeasons {
		var err error
		if filter, period, err = service.resolvePeriod(query.ReportQuery); err != nil {
			return nil, err
		}
		result.Period = &period
	}

	for _, side := range []struct {
		id     int64
		target *domain.ReportCompareSide
	}{{query.A, &result.A}, {query.B, &result.B}} {
		var err error
		switch query.Kind {
		case domain.ReportCompareSeasons:
			err = service.compareSeason(side.id, query.ReportQuery, side.target)
		default:
			err = service.compareInPeriod(query.Kind, side.id, filter, period.Granularity, side.target)
		}
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (service *ReportService) compareInPeriod(kind domain.ReportCompareKind, id int64, filter domain.ReportFilter, granularity string, side *domain.ReportCompareSide) error {
	name, err := service.repo.GetCompareName(kind, id)
	if err != nil {
		return err
	}
	if name == "" {
		return ErrCompareNotFound
	}
	if kind == domain.ReportCompareOperators {
		filter.OperatorID = id
	} else {
		filter.MachineID = id
	}
	*side = domain.ReportCompareSide{ID: id, Name: name}
	if err := service.fillCompareOperations(filter, side); err != nil {
		return err
	}
	timeline, err := service.repo.GetOperationsTimeline(filter, granularity)
	if err != nil {
		return fmt.Errorf("serie temporală: %w", err)
	}
	side.Timeline = fillTimeline(timeline, filter.From, filter.To, granularity)
	return nil
}

// compareSeason: tot sezonul (perioada nu se aplică); filtrele de teren și tip, da.
func (service *ReportService) compareSeason(id int64, query ReportQuery, side *domain.ReportCompareSide) error {
	if service.crops == nil {
		return errors.New("raportul pe culturi nu este configurat")
	}
	season, err := service.crops.GetSeasonByID(id)
	if err != nil {
		return err
	}
	if season == nil {
		return ErrCompareNotFound
	}
	*side = domain.ReportCompareSide{ID: id, Name: season.Name, Season: season}
	filter := domain.ReportFilter{SeasonID: id, FieldID: query.FieldID, OperationType: domain.OperationType(query.OperationType)}
	if err := service.fillCompareOperations(filter, side); err != nil {
		return err
	}
	crops, err := service.GetCrops(id, query.FieldID)
	if err != nil {
		return fmt.Errorf("culturi: %w", err)
	}
	side.CropTotals = &crops.Totals
	side.ByCrop = crops.ByCrop
	return nil
}

func (service *ReportService) fillCompareOperations(filter domain.ReportFilter, side *domain.ReportCompareSide) error {
	metrics, err := service.repo.GetOperationsMetrics(filter)
	if err != nil {
		return fmt.Errorf("metrici: %w", err)
	}
	byType, err := service.repo.GetOperationsByType(filter)
	if err != nil {
		return fmt.Errorf("pe tip: %w", err)
	}
	side.Metrics = *metrics
	side.ByType = byType
	return nil
}
