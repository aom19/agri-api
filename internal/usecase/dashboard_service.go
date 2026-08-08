package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

type DashboardService struct {
	dashboardRepo repository.DashboardRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo}
}

func (service *DashboardService) GetCards() ([]domain.DashboardCard, error) {
	stats, err := service.dashboardRepo.GetCardStats()
	if err != nil {
		return nil, err
	}

	capacity := minPositive(stats.TotalMachines, stats.TotalOperators)

	return []domain.DashboardCard{
		{
			Key:      domain.DashboardCardTotalMachines,
			Label:    "Total mașini",
			Value:    stats.TotalMachines,
			Trend:    stats.NewMachinesLast7Days,
			Progress: percentage(stats.ActiveAssignments, stats.TotalMachines),
		},
		{
			Key:      domain.DashboardCardActiveMachines,
			Label:    "Mașini active",
			Value:    stats.ActiveMachines,
			Trend:    stats.NewActiveMachinesLast7Days,
			Progress: percentage(stats.ActiveMachines, stats.TotalMachines),
		},
		{
			Key:      domain.DashboardCardTotalOperators,
			Label:    "Total operatori",
			Value:    stats.TotalOperators,
			Trend:    stats.NewOperatorsLast7Days,
			Progress: percentage(stats.ActiveOperators, stats.TotalOperators),
		},
		{
			Key:      domain.DashboardCardActiveAssignments,
			Label:    "Alocări active",
			Value:    stats.ActiveAssignments,
			Trend:    stats.NewActiveAssignmentsLast7Days,
			Progress: percentage(stats.ActiveAssignments, capacity),
		},
	}, nil
}

func percentage(value int, total int) int {
	if total <= 0 || value <= 0 {
		return 0
	}

	result := (value*100 + total/2) / total
	if result > 100 {
		return 100
	}

	return result
}

func minPositive(first int, second int) int {
	if first <= 0 {
		return second
	}
	if second <= 0 || first < second {
		return first
	}

	return second
}
