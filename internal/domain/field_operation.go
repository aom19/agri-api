package domain

import "time"

type FieldOperationStatus string

const (
	FieldOperationStatusPlanned    FieldOperationStatus = "planned"
	FieldOperationStatusInProgress FieldOperationStatus = "in_progress"
	FieldOperationStatusCompleted  FieldOperationStatus = "completed"
	FieldOperationStatusCanceled   FieldOperationStatus = "canceled"
)

type FieldOperation struct {
	ID                  int64                `json:"id"`
	FieldID             string               `json:"field_id"`
	OperationTypeID     int64                `json:"operation_type_id"`
	OperationTemplateID *int64               `json:"operation_template_id,omitempty"`
	MachineID           *int64               `json:"machine_id,omitempty"`
	ImplementID         *int64               `json:"implement_id,omitempty"`
	OperatorID          *int64               `json:"operator_id,omitempty"`
	PlannedStartAt      *time.Time           `json:"planned_start_at,omitempty"`
	PlannedEndAt        *time.Time           `json:"planned_end_at,omitempty"`
	AreaPlannedHa       *float64             `json:"area_planned_ha,omitempty"`
	Notes               string               `json:"notes"`
	Status              FieldOperationStatus `json:"status"`
	ActualStartAt       *time.Time           `json:"actual_start_at,omitempty"`
	ActualEndAt         *time.Time           `json:"actual_end_at,omitempty"`
	AreaCompletedHa     *float64             `json:"area_completed_ha,omitempty"`
	FuelUsedL           *float64             `json:"fuel_used_l,omitempty"`
	MachineHours        *float64             `json:"machine_hours,omitempty"`
	CompletionNotes     string               `json:"completion_notes"`

	Auditfields
}

type FieldOperationChecklist struct {
	MachineStatus   bool `json:"machine_status"`
	ImplementStatus bool `json:"implement_status"`
	FieldArea       bool `json:"field_area"`
	NotesConfirmed  bool `json:"notes_confirmed"`
}
