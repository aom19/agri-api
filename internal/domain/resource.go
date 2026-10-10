package domain

import "time"

// ResourceCategory — categoria resursei agricole
type ResourceCategory string

const (
	ResourceCategoryFuel       ResourceCategory = "fuel"
	ResourceCategoryFertilizer ResourceCategory = "fertilizer"
	ResourceCategorySeed       ResourceCategory = "seed"
	ResourceCategoryPesticide  ResourceCategory = "pesticide"
	ResourceCategoryWater      ResourceCategory = "water"
	ResourceCategoryHarvest    ResourceCategory = "harvest"
	ResourceCategoryOther      ResourceCategory = "other"
)

var validResourceCategories = map[ResourceCategory]struct{}{
	ResourceCategoryFuel:       {},
	ResourceCategoryFertilizer: {},
	ResourceCategorySeed:       {},
	ResourceCategoryPesticide:  {},
	ResourceCategoryWater:      {},
	ResourceCategoryHarvest:    {},
	ResourceCategoryOther:      {},
}

func (c ResourceCategory) IsValid() bool {
	_, ok := validResourceCategories[c]
	return ok
}

// ResourceType — tipul de resursă agricolă
type ResourceType struct {
	ID          int64            `json:"id"`
	Name        string           `json:"name"`
	Category    ResourceCategory `json:"category"`
	DefaultUnit string           `json:"default_unit"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// Resource — resursa agricolă, cu stocul ei. Quantity se modifică doar prin mișcări de stoc,
// astfel încât suma mișcărilor unei resurse să fie egală cu stocul ei curent.
type Resource struct {
	ID              int64         `json:"id"`
	Name            string        `json:"name"`
	ResourceTypeID  int64         `json:"resource_type_id"`
	ResourceType    *ResourceType `json:"resource_type,omitempty"`
	PricePerUnit    float64       `json:"price_per_unit"`
	Quantity        float64       `json:"quantity"`
	MinimumQuantity float64       `json:"minimum_quantity"`
	Notes           string        `json:"notes,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
