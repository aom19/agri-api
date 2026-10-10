package domain

import "time"

type StockMovementType string

const (
	StockMovementIn         StockMovementType = "in"
	StockMovementOut        StockMovementType = "out"
	StockMovementAdjustment StockMovementType = "adjustment"
)

// StockMovement este o intrare din istoricul stocului: recepție, consum (eventual legat
// de o operațiune pe teren) sau ajustare de inventar.
type StockMovement struct {
	ID                  int64             `json:"id"`
	ResourceID          int64             `json:"resource_id"`
	ResourceName        string            `json:"resource_name"`
	Category            string            `json:"category"`
	Unit                string            `json:"unit"`
	FieldOperationID    *int64            `json:"field_operation_id,omitempty"`
	FieldOperationLabel *string           `json:"field_operation_label,omitempty"`
	FieldCropID         *int64            `json:"field_crop_id,omitempty"`
	MovementType        StockMovementType `json:"movement_type"`
	QuantityDelta       float64           `json:"quantity_delta"`
	ResultingQuantity   float64           `json:"resulting_quantity"`
	UnitCost            *float64          `json:"unit_cost,omitempty"`
	TotalCost           *float64          `json:"total_cost,omitempty"`
	Notes               string            `json:"notes"`
	ActorID             *int64            `json:"actor_id,omitempty"`
	ActorName           *string           `json:"actor_name,omitempty"`
	CreatedAt           time.Time         `json:"created_at"`
}

// StockMovementInput este cererea de înregistrare a unei mișcări.
// Pentru `in`/`out`, Quantity este cantitatea mutată (pozitivă); pentru `adjustment`,
// Quantity este noul nivel al stocului.
type StockMovementInput struct {
	ResourceID       int64
	FieldOperationID *int64
	MovementType     StockMovementType
	Quantity         float64
	UnitCost         *float64
	Notes            string
	ActorID          *int64
}

type StockMovementFilter struct {
	ResourceID       int64
	FieldOperationID int64
	From             *time.Time
	To               *time.Time
	Limit            int
}

// FieldOperationResourceUsage este consumul efectiv al unei resurse la finalizarea unei operațiuni.
type FieldOperationResourceUsage struct {
	ResourceID int64   `json:"resource_id"`
	Quantity   float64 `json:"quantity"`
}

// FieldOperationCompletion sunt datele reale înregistrate la finalizarea unei operațiuni.
type FieldOperationCompletion struct {
	ActualEndAt     *time.Time
	AreaCompletedHa *float64
	// FuelUsedL este combustibilul raportat; devine ieșire din stocul resursei FuelResourceID.
	FuelUsedL      *float64
	FuelResourceID *int64
	MachineHours   *float64
	Notes          string
	// Resources sunt consumurile declarate explicit. Cu ConsumeFromTemplate, ele corectează
	// (sau completează) consumul calculat din normele șablonului.
	Resources           []FieldOperationResourceUsage
	ConsumeFromTemplate bool
}

// TemplateResourceUsage este norma de consum a unei resurse dintr-un șablon de operațiune.
type TemplateResourceUsage struct {
	ResourceID      int64
	ResourceName    string
	Category        string
	Unit            string
	QuantityPerUnit float64
	PricePerUnit    float64
}

// ConsumptionEstimateItem este consumul estimat al unei resurse: normă × suprafață.
type ConsumptionEstimateItem struct {
	ResourceID      int64   `json:"resource_id"`
	ResourceName    string  `json:"resource_name"`
	Category        string  `json:"category"`
	Unit            string  `json:"unit"`
	QuantityPerUnit float64 `json:"quantity_per_unit"`
	Quantity        float64 `json:"quantity"`
}

// FuelStock este o resursă de combustibil, din al cărei stoc se poate scădea combustibilul raportat.
type FuelStock struct {
	ResourceID   int64   `json:"resource_id"`
	ResourceName string  `json:"resource_name"`
	Unit         string  `json:"unit"`
	Quantity     float64 `json:"quantity"`
}

// ConsumptionEstimate este previzualizarea consumului la finalizare. Calculul e același
// cu cel aplicat la finalizarea cu consume_from_template.
type ConsumptionEstimate struct {
	AreaHa        float64                   `json:"area_ha"`
	Items         []ConsumptionEstimateItem `json:"items"`
	FuelResources []FuelStock               `json:"fuel_resources"`
}
