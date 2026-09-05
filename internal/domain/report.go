package domain

import (
	"encoding/json"
	"time"
)

// ReportFilter conține filtrele comune tuturor rapoartelor. Intervalul este
// [From, To) — To este exclusiv (ziua următoare ultimei zile selectate).
type ReportFilter struct {
	From            time.Time
	To              time.Time
	FieldID         string
	OperationTypeID int64
	MachineID       int64
	OperatorID      int64
}

// ReportPeriod descrie intervalul raportat și intervalul anterior de aceeași lungime,
// folosit pentru comparații.
type ReportPeriod struct {
	From         string `json:"from"`
	To           string `json:"to"`
	PreviousFrom string `json:"previous_from"`
	PreviousTo   string `json:"previous_to"`
	Granularity  string `json:"granularity"`
}

// ReportOperationsMetrics sunt indicatorii agregați ai operațiunilor pe teren dintr-un interval.
type ReportOperationsMetrics struct {
	OperationsTotal      int     `json:"operations_total"`
	OperationsPlanned    int     `json:"operations_planned"`
	OperationsInProgress int     `json:"operations_in_progress"`
	OperationsCompleted  int     `json:"operations_completed"`
	OperationsCanceled   int     `json:"operations_canceled"`
	OverdueOperations    int     `json:"overdue_operations"`
	OnTimeCompleted      int     `json:"on_time_completed"`
	PlannedAreaHa        float64 `json:"planned_area_ha"`
	CompletedAreaHa      float64 `json:"completed_area_ha"`
	EstimatedCost        float64 `json:"estimated_cost"`
	FieldsWorked         int     `json:"fields_worked"`
	MachinesUsed         int     `json:"machines_used"`
	OperatorsUsed        int     `json:"operators_used"`
	// Valori reale (faza 2): disponibile doar pentru operațiunile finalizate cu date reale.
	CompletedWithActuals  int     `json:"completed_with_actuals"`
	ActualDurationMinutes int64   `json:"actual_duration_minutes"`
	RealizedAreaHa        float64 `json:"realized_area_ha"`
	FuelUsedL             float64 `json:"fuel_used_l"`
	MachineHours          float64 `json:"machine_hours"`
	RealCost              float64 `json:"real_cost"`
}

// ReportInventorySnapshot este starea curentă a fermei, independentă de interval.
type ReportInventorySnapshot struct {
	TotalFields       int     `json:"total_fields"`
	TotalAreaHa       float64 `json:"total_area_ha"`
	TotalMachines     int     `json:"total_machines"`
	ActiveMachines    int     `json:"active_machines"`
	MaintenanceAssets int     `json:"maintenance_assets"`
	TotalOperators    int     `json:"total_operators"`
	ActiveOperators   int     `json:"active_operators"`
	TotalStocks       int     `json:"total_stocks"`
	LowStocks         int     `json:"low_stocks"`
	StockValue        float64 `json:"stock_value"`
}

type ReportSummary struct {
	Period    ReportPeriod            `json:"period"`
	Current   ReportOperationsMetrics `json:"current"`
	Previous  ReportOperationsMetrics `json:"previous"`
	Inventory ReportInventorySnapshot `json:"inventory"`
}

// ReportTimeBucket este un punct din seria temporală a operațiunilor.
type ReportTimeBucket struct {
	Bucket     string  `json:"bucket"`
	Planned    int     `json:"planned"`
	InProgress int     `json:"in_progress"`
	Completed  int     `json:"completed"`
	Canceled   int     `json:"canceled"`
	AreaHa     float64 `json:"area_ha"`
}

type ReportNamedCount struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type ReportNamedValue struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type ReportOperationTypeStat struct {
	OperationTypeID   int64   `json:"operation_type_id"`
	OperationTypeName string  `json:"operation_type_name"`
	Total             int     `json:"total"`
	Completed         int     `json:"completed"`
	AreaHa            float64 `json:"area_ha"`
	EstimatedCost     float64 `json:"estimated_cost"`
}

type ReportOperationRow struct {
	ID                int64      `json:"id"`
	FieldID           string     `json:"field_id"`
	FieldName         string     `json:"field_name"`
	OperationTypeName string     `json:"operation_type_name"`
	TemplateName      *string    `json:"template_name,omitempty"`
	MachineName       *string    `json:"machine_name,omitempty"`
	ImplementName     *string    `json:"implement_name,omitempty"`
	OperatorName      *string    `json:"operator_name,omitempty"`
	PlannedStartAt    *time.Time `json:"planned_start_at,omitempty"`
	PlannedEndAt      *time.Time `json:"planned_end_at,omitempty"`
	AreaPlannedHa     *float64   `json:"area_planned_ha,omitempty"`
	Status            string     `json:"status"`
	DelayMinutes      int64      `json:"delay_minutes"`
	EstimatedCost     float64    `json:"estimated_cost"`
	ActualStartAt     *time.Time `json:"actual_start_at,omitempty"`
	ActualEndAt       *time.Time `json:"actual_end_at,omitempty"`
	ActualDurationMin *int64     `json:"actual_duration_minutes,omitempty"`
	AreaCompletedHa   *float64   `json:"area_completed_ha,omitempty"`
	FuelUsedL         *float64   `json:"fuel_used_l,omitempty"`
	MachineHours      *float64   `json:"machine_hours,omitempty"`
	RealCost          float64    `json:"real_cost"`
}

type ReportOperations struct {
	Period   ReportPeriod              `json:"period"`
	Metrics  ReportOperationsMetrics   `json:"metrics"`
	Timeline []ReportTimeBucket        `json:"timeline"`
	ByType   []ReportOperationTypeStat `json:"by_type"`
	ByStatus []ReportNamedCount        `json:"by_status"`
	Items    []ReportOperationRow      `json:"items"`
}

type ReportFieldRow struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CadastralNumber *string         `json:"cadastral_number,omitempty"`
	AreaHa          *float64        `json:"area_ha,omitempty"`
	Geometry        json.RawMessage `json:"geometry,omitempty" swaggertype:"object"`
	OperationsCount int             `json:"operations_count"`
	CompletedCount  int             `json:"completed_count"`
	PlannedAreaHa   float64         `json:"planned_area_ha"`
	EstimatedCost   float64         `json:"estimated_cost"`
	RealCost        float64         `json:"real_cost"`
	RealizedAreaHa  float64         `json:"realized_area_ha"`
	FuelUsedL       float64         `json:"fuel_used_l"`
	LastOperationAt *time.Time      `json:"last_operation_at,omitempty"`
}

type ReportFields struct {
	Period               ReportPeriod     `json:"period"`
	TotalFields          int              `json:"total_fields"`
	FieldsWithOperations int              `json:"fields_with_operations"`
	TotalAreaHa          float64          `json:"total_area_ha"`
	WorkedAreaHa         float64          `json:"worked_area_ha"`
	Items                []ReportFieldRow `json:"items"`
}

type ReportMachineRow struct {
	ID                int64    `json:"id"`
	Name              string   `json:"name"`
	Code              string   `json:"code"`
	Type              string   `json:"type"`
	Status            string   `json:"status"`
	FuelType          *string  `json:"fuel_type,omitempty"`
	Year              *int     `json:"year,omitempty"`
	OperatingHours    *float64 `json:"operating_hours,omitempty"`
	OperationsCount   int      `json:"operations_count"`
	PlannedAreaHa     float64  `json:"planned_area_ha"`
	ActiveAssignments int      `json:"active_assignments"`
	FuelUsedL         float64  `json:"fuel_used_l"`
	HoursInPeriod     float64  `json:"hours_in_period"`
}

type ReportImplementRow struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Code            string   `json:"code"`
	Type            string   `json:"type"`
	Status          string   `json:"status"`
	WorkingWidth    *float64 `json:"working_width,omitempty"`
	OperationsCount int      `json:"operations_count"`
}

type ReportFleet struct {
	Period          ReportPeriod         `json:"period"`
	MachineStatus   []ReportNamedCount   `json:"machine_status"`
	ImplementStatus []ReportNamedCount   `json:"implement_status"`
	MachinesByType  []ReportNamedCount   `json:"machines_by_type"`
	MachinesByFuel  []ReportNamedCount   `json:"machines_by_fuel"`
	MachinesByYear  []ReportNamedCount   `json:"machines_by_year"`
	Machines        []ReportMachineRow   `json:"machines"`
	Implements      []ReportImplementRow `json:"implements"`
}

type ReportOperatorRow struct {
	ID                  int64    `json:"id"`
	Name                string   `json:"name"`
	Status              string   `json:"status"`
	AllowedMachineTypes []string `json:"allowed_machine_types"`
	OperationsCount     int      `json:"operations_count"`
	CompletedCount      int      `json:"completed_count"`
	InProgressCount     int      `json:"in_progress_count"`
	PlannedCount        int      `json:"planned_count"`
	OnTimeCount         int      `json:"on_time_count"`
	OverdueCount        int      `json:"overdue_count"`
	PlannedAreaHa       float64  `json:"planned_area_ha"`
	ActiveAssignments   int      `json:"active_assignments"`
}

type ReportOperators struct {
	Period          ReportPeriod        `json:"period"`
	TotalOperators  int                 `json:"total_operators"`
	ActiveOperators int                 `json:"active_operators"`
	Items           []ReportOperatorRow `json:"items"`
}

type ReportStockRow struct {
	ID              int64   `json:"id"`
	ResourceName    string  `json:"resource_name"`
	Category        string  `json:"category"`
	Unit            string  `json:"unit"`
	Quantity        float64 `json:"quantity"`
	MinimumQuantity float64 `json:"minimum_quantity"`
	PricePerUnit    float64 `json:"price_per_unit"`
	Value           float64 `json:"value"`
	BelowMinimum    bool    `json:"below_minimum"`
}

// ReportResourceConsumption este consumul estimat al unei resurse, calculat din
// șabloanele operațiunilor din interval (cantitate/unitate × suprafață planificată).
type ReportResourceConsumption struct {
	ResourceID    int64    `json:"resource_id"`
	ResourceName  string   `json:"resource_name"`
	Category      string   `json:"category"`
	Unit          string   `json:"unit"`
	Quantity      float64  `json:"quantity"`
	Cost          float64  `json:"cost"`
	StockQuantity *float64 `json:"stock_quantity,omitempty"`
}

// ReportMovementTotals sunt totalurile mișcărilor de stoc din interval.
type ReportMovementTotals struct {
	InCount  int     `json:"in_count"`
	InValue  float64 `json:"in_value"`
	OutCount int     `json:"out_count"`
	OutValue float64 `json:"out_value"`
}

type ReportStocks struct {
	Period               ReportPeriod                `json:"period"`
	RealConsumption      []ReportResourceConsumption `json:"real_consumption"`
	RealConsumptionCost  float64                     `json:"real_consumption_cost"`
	MovementTotals       ReportMovementTotals        `json:"movement_totals"`
	TotalStocks          int                         `json:"total_stocks"`
	LowStocks            int                         `json:"low_stocks"`
	TotalValue           float64                     `json:"total_value"`
	ValueByCategory      []ReportNamedValue          `json:"value_by_category"`
	Items                []ReportStockRow            `json:"items"`
	EstimatedConsumption []ReportResourceConsumption `json:"estimated_consumption"`
}
