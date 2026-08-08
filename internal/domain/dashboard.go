package domain

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
