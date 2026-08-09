package domain

import "time"

type OperationType struct {
	ID          int64      `json:"id"`
	Code        string     `json:"code"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"-"`
}

type OperationTemplate struct {
	ID              int64      `json:"id"`
	OperationTypeID int64      `json:"operation_type_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	Unit            string     `json:"unit"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"-"`

	OperationType  *OperationType     `json:"operation_type,omitempty"`
	Resources      []TemplateResource `json:"resources,omitempty"`
	MachineTypes   []string           `json:"machine_types,omitempty"`
	ImplementTypes []string           `json:"implement_types,omitempty"`
}

type TemplateResource struct {
	ID              int64     `json:"id"`
	TemplateID      int64     `json:"template_id"`
	ResourceID      int64     `json:"resource_id"`
	QuantityPerUnit float64   `json:"quantity_per_unit"`
	Notes           string    `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	Resource *Resource `json:"resource,omitempty"`
}
