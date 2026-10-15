package dto

import (
	"encoding/json"
	"time"
)

type FieldOperationResponse struct {
	ID            int64           `json:"id"`
	FieldID       string          `json:"field_id"`
	FieldName     string          `json:"field_name"`
	FieldGeometry json.RawMessage `json:"field_geometry,omitempty"`
	// OperationType este tipul efectiv: al template-ului, sau tipul propriu fără template.
	OperationType       string     `json:"operation_type"`
	OperationTypeName   string     `json:"operation_type_name"`
	OperationTemplateID *int64     `json:"operation_template_id,omitempty"`
	OperationTemplate   *string    `json:"operation_template_name,omitempty"`
	MachineID           *int64     `json:"machine_id,omitempty"`
	MachineName         *string    `json:"machine_name,omitempty"`
	MachineStatus       *string    `json:"machine_status,omitempty"`
	ImplementID         *int64     `json:"implement_id,omitempty"`
	ImplementName       *string    `json:"implement_name,omitempty"`
	ImplementStatus     *string    `json:"implement_status,omitempty"`
	OperatorID          *int64     `json:"operator_id,omitempty"`
	OperatorName        *string    `json:"operator_name,omitempty"`
	PlannedStartAt      *time.Time `json:"planned_start_at,omitempty"`
	PlannedEndAt        *time.Time `json:"planned_end_at,omitempty"`
	AreaPlannedHa       *float64   `json:"area_planned_ha,omitempty"`
	Notes               string     `json:"notes"`
	Status              string     `json:"status"`
	FieldCropID         *int64     `json:"field_crop_id,omitempty"`
	CropName            *string    `json:"crop_name,omitempty"`
	SeasonName          *string    `json:"season_name,omitempty"`
	ActualStartAt       *time.Time `json:"actual_start_at,omitempty"`
	ActualEndAt         *time.Time `json:"actual_end_at,omitempty"`
	ActualDurationMin   *int64     `json:"actual_duration_minutes,omitempty"`
	AreaCompletedHa     *float64   `json:"area_completed_ha,omitempty"`
	FuelUsedL           *float64   `json:"fuel_used_l,omitempty"`
	MachineHours        *float64   `json:"machine_hours,omitempty"`
	CompletionNotes     string     `json:"completion_notes"`
	// Conflicts sunt celelalte lucrări planificate sau în lucru care folosesc același operator,
	// aceeași mașină sau același echipament într-un interval suprapus. Doar avertisment: nu
	// blochează salvarea, ci doar pornirea cât timp o lucrare suprapusă e în lucru.
	Conflicts []FieldOperationConflict `json:"conflicts"`
	CreatedAt time.Time                `json:"created_at"`
	UpdatedAt time.Time                `json:"updated_at"`
}

// Resursele care pot fi ocupate de două lucrări în același timp.
const (
	ConflictResourceOperator  = "operator"
	ConflictResourceMachine   = "machine"
	ConflictResourceImplement = "implement"
)

// FieldOperationConflict este o altă lucrare care ține ocupată, în același interval, una din
// resursele lucrării: operatorul, mașina sau echipamentul (Resources spune care).
type FieldOperationConflict struct {
	OperationID       int64     `json:"operation_id"`
	FieldName         string    `json:"field_name"`
	OperationTypeName string    `json:"operation_type_name"`
	Status            string    `json:"status"`
	StartAt           time.Time `json:"start_at"`
	EndAt             time.Time `json:"end_at"`
	Resources         []string  `json:"resources"`
}

// OverdueFieldOperation descrie o operațiune pe teren aflată în lucru care a depășit
// sfârșitul planificat (timpul estimat de lucru); folosită de monitorul de notificări.
type OverdueFieldOperation struct {
	ID                int64
	FieldName         string
	OperationTypeName string
	OperatorID        *int64
	OperatorName      *string
	OperatorUserID    *int64
	PlannedStartAt    *time.Time
	PlannedEndAt      time.Time
}
