package dto

import (
	"encoding/json"
	"time"
)

type FieldOperationResponse struct {
	ID                  int64                           `json:"id"`
	FieldID             string                          `json:"field_id"`
	FieldName           string                          `json:"field_name"`
	FieldGeometry       json.RawMessage                 `json:"field_geometry,omitempty"`
	OperationTypeID     int64                           `json:"operation_type_id"`
	OperationTypeCode   string                          `json:"operation_type_code"`
	OperationTypeName   string                          `json:"operation_type_name"`
	OperationTemplateID *int64                          `json:"operation_template_id,omitempty"`
	OperationTemplate   *string                         `json:"operation_template_name,omitempty"`
	MachineID           *int64                          `json:"machine_id,omitempty"`
	MachineName         *string                         `json:"machine_name,omitempty"`
	MachineStatus       *string                         `json:"machine_status,omitempty"`
	ImplementID         *int64                          `json:"implement_id,omitempty"`
	ImplementName       *string                         `json:"implement_name,omitempty"`
	ImplementStatus     *string                         `json:"implement_status,omitempty"`
	OperatorID          *int64                          `json:"operator_id,omitempty"`
	OperatorName        *string                         `json:"operator_name,omitempty"`
	PlannedStartAt      *time.Time                      `json:"planned_start_at,omitempty"`
	PlannedEndAt        *time.Time                      `json:"planned_end_at,omitempty"`
	AreaPlannedHa       *float64                        `json:"area_planned_ha,omitempty"`
	Notes               string                          `json:"notes"`
	Status              string                          `json:"status"`
	Checklist           FieldOperationChecklistResponse `json:"checklist"`
	CreatedAt           time.Time                       `json:"created_at"`
	UpdatedAt           time.Time                       `json:"updated_at"`
}

type FieldOperationChecklistResponse struct {
	MachineStatus   bool       `json:"machine_status"`
	ImplementStatus bool       `json:"implement_status"`
	FieldArea       bool       `json:"field_area"`
	NotesConfirmed  bool       `json:"notes_confirmed"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}
