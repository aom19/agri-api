package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

type DashboardService struct {
	dashboardRepo repository.DashboardRepository
	auditRepo     repository.AuditRepository
}

func NewDashboardService(dashboardRepo repository.DashboardRepository, auditRepo repository.AuditRepository) *DashboardService {
	return &DashboardService{dashboardRepo: dashboardRepo, auditRepo: auditRepo}
}

// GetQuickStats returnează indicatorii rapizi calculați din datele reale.
func (service *DashboardService) GetQuickStats() (*domain.DashboardQuickStats, error) {
	return service.dashboardRepo.GetQuickStats()
}

// GetRecentActivity returnează ultimele intrări din jurnalul de audit, simplificate pentru dashboard.
func (service *DashboardService) GetRecentActivity(limit int) ([]domain.DashboardActivityItem, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 50 {
		limit = 50
	}
	if service.auditRepo == nil {
		return []domain.DashboardActivityItem{}, nil
	}

	entries, err := service.auditRepo.GetAll(limit, "", "")
	if err != nil {
		return nil, err
	}

	items := make([]domain.DashboardActivityItem, 0, len(entries))
	for _, entry := range entries {
		items = append(items, domain.DashboardActivityItem{
			ID:         entry.ID,
			EntityType: entry.EntityType,
			EntityID:   entry.EntityID,
			EntityName: entry.EntityName,
			Action:     entry.Action,
			ActorName:  entry.ActorName,
			Status:     changeString(entry.Changes, "status"),
			OldStatus:  changeString(entry.Changes, "old_status"),
			CreatedAt:  entry.CreatedAt,
		})
	}
	return items, nil
}

// changeString extrage o valoare text din harta de modificări a unei intrări de audit.
func changeString(changes map[string]interface{}, key string) *string {
	if changes == nil {
		return nil
	}
	value, ok := changes[key].(string)
	if !ok || value == "" {
		return nil
	}
	return &value
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
