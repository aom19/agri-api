package dto

import "time"

type FieldOperationResponse struct {
	ID                  int64      `json:"id"`
	FieldID             string     `json:"field_id"`
	FieldName           string     `json:"field_name"`
	OperationTypeID     int64      `json:"operation_type_id"`
	OperationTypeCode   string     `json:"operation_type_code"`
	OperationTypeName   string     `json:"operation_type_name"`
	OperationTemplateID *int64     `json:"operation_template_id,omitempty"`
	OperationTemplate   *string    `json:"operation_template_name,omitempty"`
	MachineID           *int64     `json:"machine_id,omitempty"`
	MachineName         *string    `json:"machine_name,omitempty"`
	ImplementID         *int64     `json:"implement_id,omitempty"`
	ImplementName       *string    `json:"implement_name,omitempty"`
	OperatorID          *int64     `json:"operator_id,omitempty"`
	OperatorName        *string    `json:"operator_name,omitempty"`
	PlannedStartAt      *time.Time `json:"planned_start_at,omitempty"`
	PlannedEndAt        *time.Time `json:"planned_end_at,omitempty"`
	AreaPlannedHa       *float64   `json:"area_planned_ha,omitempty"`
	Notes               string     `json:"notes"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
