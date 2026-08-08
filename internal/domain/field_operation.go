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

	Auditfields
}
