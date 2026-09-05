package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type DashboardRepo struct {
	db *sql.DB
}

func NewDashboardRepo(db *sql.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

func (repo *DashboardRepo) GetCardStats() (*domain.DashboardCardStats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL) AS total_machines,
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL AND asset_status = 'active') AS active_machines,
			(SELECT COUNT(*) FROM operators WHERE deleted_at IS NULL) AS total_operators,
			(SELECT COUNT(*) FROM operators WHERE deleted_at IS NULL AND status = 'active') AS active_operators,
			(SELECT COUNT(*) FROM field_operations WHERE deleted_at IS NULL) AS total_assignments,
			(SELECT COUNT(*) FROM field_operations WHERE deleted_at IS NULL AND status IN ('planned', 'in_progress')) AS active_assignments,
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL AND created_at >= NOW() - INTERVAL '7 days') AS new_machines_last_7_days,
			(SELECT COUNT(*) FROM machines WHERE deleted_at IS NULL AND asset_status = 'active' AND created_at >= NOW() - INTERVAL '7 days') AS new_active_machines_last_7_days,
			(SELECT COUNT(*) FROM operators WHERE deleted_at IS NULL AND created_at >= NOW() - INTERVAL '7 days') AS new_operators_last_7_days,
			(SELECT COUNT(*) FROM field_operations WHERE deleted_at IS NULL AND status IN ('planned', 'in_progress') AND created_at >= NOW() - INTERVAL '7 days') AS new_active_assignments_last_7_days
	`

	stats := &domain.DashboardCardStats{}
	err := repo.db.QueryRow(query).Scan(
		&stats.TotalMachines,
		&stats.ActiveMachines,
		&stats.TotalOperators,
		&stats.ActiveOperators,
		&stats.TotalAssignments,
		&stats.ActiveAssignments,
		&stats.NewMachinesLast7Days,
		&stats.NewActiveMachinesLast7Days,
		&stats.NewOperatorsLast7Days,
		&stats.NewActiveAssignmentsLast7Days,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetQuickStats calculează indicatorii rapizi din dashboard direct din baza de date.
func (repo *DashboardRepo) GetQuickStats() (*domain.DashboardQuickStats, error) {
	query := `
		SELECT
			(SELECT COUNT(*) FROM fields WHERE deleted_at IS NULL) AS total_fields,
			(SELECT COUNT(DISTINCT field_id) FROM field_operations
				WHERE deleted_at IS NULL AND status IN ('planned', 'in_progress')) AS active_fields,
			(SELECT COUNT(*) FROM field_operations
				WHERE deleted_at IS NULL AND status = 'in_progress') AS in_progress_operations,
			(SELECT COUNT(*) FROM field_operations
				WHERE deleted_at IS NULL
				  AND status IN ('planned', 'in_progress')
				  AND planned_end_at IS NOT NULL
				  AND planned_end_at < NOW()) AS overdue_operations,
			(SELECT COUNT(*) FROM machines
				WHERE deleted_at IS NULL AND asset_status = 'maintenance') AS maintenance_machines,
			(SELECT COUNT(*) FROM implements
				WHERE deleted_at IS NULL AND status = 'maintenance') AS maintenance_implements,
			(SELECT COUNT(*) FROM stocks
				WHERE minimum_quantity > 0 AND quantity <= minimum_quantity) AS low_stocks,
			(SELECT COUNT(*) FROM stocks) AS total_stocks
	`

	stats := &domain.DashboardQuickStats{}
	err := repo.db.QueryRow(query).Scan(
		&stats.TotalFields,
		&stats.ActiveFields,
		&stats.InProgressOperations,
		&stats.OverdueOperations,
		&stats.MaintenanceMachines,
		&stats.MaintenanceImplements,
		&stats.LowStocks,
		&stats.TotalStocks,
	)
	if err != nil {
		return nil, err
	}

	return stats, nil
}
