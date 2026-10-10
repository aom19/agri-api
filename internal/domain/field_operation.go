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
	ID      int64  `json:"id"`
	FieldID string `json:"field_id"`
	// OperationType este tipul propriu, reținut doar fără template; cu template, tipul vine din el.
	OperationType       OperationType        `json:"operation_type,omitempty"`
	OperationTemplateID *int64               `json:"operation_template_id,omitempty"`
	MachineID           *int64               `json:"machine_id,omitempty"`
	ImplementID         *int64               `json:"implement_id,omitempty"`
	OperatorID          *int64               `json:"operator_id,omitempty"`
	PlannedStartAt      *time.Time           `json:"planned_start_at,omitempty"`
	PlannedEndAt        *time.Time           `json:"planned_end_at,omitempty"`
	AreaPlannedHa       *float64             `json:"area_planned_ha,omitempty"`
	Notes               string               `json:"notes"`
	Status              FieldOperationStatus `json:"status"`
	FieldCropID         *int64               `json:"field_crop_id,omitempty"`
	ActualStartAt       *time.Time           `json:"actual_start_at,omitempty"`
	ActualEndAt         *time.Time           `json:"actual_end_at,omitempty"`
	AreaCompletedHa     *float64             `json:"area_completed_ha,omitempty"`
	FuelUsedL           *float64             `json:"fuel_used_l,omitempty"`
	MachineHours        *float64             `json:"machine_hours,omitempty"`
	CompletionNotes     string               `json:"completion_notes"`

	Auditfields
}

// AssetCompatibility descrie ce tipuri de mașini și echipamente acceptă template-ul unei
// operațiuni (listă goală = orice tip) și ce tip au mașina și echipamentul alese (gol = neales).
type AssetCompatibility struct {
	TemplateMachineTypes   []string
	TemplateImplementTypes []string
	MachineType            string
	ImplementType          string
}
