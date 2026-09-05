package domain

import "time"

type DashboardCardKey string

const (
	DashboardCardTotalMachines     DashboardCardKey = "total_machines"
	DashboardCardActiveMachines    DashboardCardKey = "active_machines"
	DashboardCardTotalOperators    DashboardCardKey = "total_operators"
	DashboardCardActiveAssignments DashboardCardKey = "active_assignments"
)

type DashboardCard struct {
	Key      DashboardCardKey `json:"key"`
	Label    string           `json:"label"`
	Value    int              `json:"value"`
	Trend    int              `json:"trend"`
	Progress int              `json:"progress"`
}

type DashboardCardStats struct {
	TotalMachines                 int
	ActiveMachines                int
	TotalOperators                int
	ActiveOperators               int
	TotalAssignments              int
	ActiveAssignments             int
	NewMachinesLast7Days          int
	NewActiveMachinesLast7Days    int
	NewOperatorsLast7Days         int
	NewActiveAssignmentsLast7Days int
}

// DashboardQuickStats conține indicatorii rapizi afișați în dashboard, calculați din datele reale.
type DashboardQuickStats struct {
	TotalFields           int `json:"total_fields"`
	ActiveFields          int `json:"active_fields"`
	InProgressOperations  int `json:"in_progress_operations"`
	OverdueOperations     int `json:"overdue_operations"`
	MaintenanceMachines   int `json:"maintenance_machines"`
	MaintenanceImplements int `json:"maintenance_implements"`
	LowStocks             int `json:"low_stocks"`
	TotalStocks           int `json:"total_stocks"`
}

// DashboardActivityItem este o intrare din jurnalul de audit, simplificată pentru fluxul
// „Activitate recentă” din dashboard (fără detaliile brute ale modificărilor).
type DashboardActivityItem struct {
	ID         int64     `json:"id"`
	EntityType string    `json:"entity_type"`
	EntityID   string    `json:"entity_id"`
	EntityName *string   `json:"entity_name,omitempty"`
	Action     string    `json:"action"`
	ActorName  *string   `json:"actor_name,omitempty"`
	Status     *string   `json:"status,omitempty"`
	OldStatus  *string   `json:"old_status,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
