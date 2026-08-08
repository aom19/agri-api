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
