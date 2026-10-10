package domain

import "time"

// OperationType — tipul lucrării agricole. Lista e fixă: rapoartele grupează după ea.
type OperationType string

const (
	OperationTypeSoilPreparation OperationType = "soil_preparation"
	OperationTypeSeeding         OperationType = "seeding"
	OperationTypeFertilization   OperationType = "fertilization"
	OperationTypeSpraying        OperationType = "spraying"
	OperationTypeHarvesting      OperationType = "harvesting"
	OperationTypeIrrigation      OperationType = "irrigation"
)

// OperationTypes sunt toate tipurile, în ordinea de afișare.
var OperationTypes = []OperationType{
	OperationTypeSoilPreparation,
	OperationTypeSeeding,
	OperationTypeFertilization,
	OperationTypeSpraying,
	OperationTypeHarvesting,
	OperationTypeIrrigation,
}

var operationTypeLabels = map[OperationType]string{
	OperationTypeSoilPreparation: "Pregătire sol",
	OperationTypeSeeding:         "Semănat",
	OperationTypeFertilization:   "Fertilizare",
	OperationTypeSpraying:        "Stropire",
	OperationTypeHarvesting:      "Recoltare",
	OperationTypeIrrigation:      "Irigare",
}

func (t OperationType) IsValid() bool {
	_, ok := operationTypeLabels[t]
	return ok
}

// Label este numele afișat al tipului; un tip necunoscut rămâne codul lui.
func (t OperationType) Label() string {
	if label, ok := operationTypeLabels[t]; ok {
		return label
	}
	return string(t)
}

type OperationTemplate struct {
	ID            int64         `json:"id"`
	OperationType OperationType `json:"operation_type"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	Unit          string        `json:"unit"`
	CropID        *int64        `json:"crop_id,omitempty"`
	CropName      *string       `json:"crop_name,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	DeletedAt     *time.Time    `json:"-"`

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
