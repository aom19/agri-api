package domain

import "time"

// Season este un sezon agricol (interval de date) pentru care se urmăresc culturile și producția.
type Season struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	StartDate string    `json:"start_date"`
	EndDate   string    `json:"end_date"`
	IsActive  bool      `json:"is_active"`
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Crop este o cultură din catalog.
type Crop struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Code              *string   `json:"code,omitempty"`
	Category          string    `json:"category"`
	YieldUnit         string    `json:"yield_unit"`
	Notes             string    `json:"notes"`
	HarvestResourceID *int64    `json:"harvest_resource_id,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// FieldCrop leagă un teren de o cultură într-un sezon și reține producția obținută.
type FieldCrop struct {
	ID                 int64      `json:"id"`
	FieldID            string     `json:"field_id"`
	FieldName          string     `json:"field_name"`
	FieldAreaHa        *float64   `json:"field_area_ha,omitempty"`
	SeasonID           int64      `json:"season_id"`
	SeasonName         string     `json:"season_name"`
	SeasonStart        string     `json:"season_start"`
	SeasonEnd          string     `json:"season_end"`
	CropID             int64      `json:"crop_id"`
	CropName           string     `json:"crop_name"`
	YieldUnit          string     `json:"yield_unit"`
	PlantedAreaHa      *float64   `json:"planted_area_ha,omitempty"`
	PlantedAt          *string    `json:"planted_at,omitempty"`
	HarvestedAt        *string    `json:"harvested_at,omitempty"`
	ProductionTotal    *float64   `json:"production_total,omitempty"`
	ExpectedYieldPerHa *float64   `json:"expected_yield_per_ha,omitempty"`
	YieldPerHa         *float64   `json:"yield_per_ha,omitempty"`
	Notes              string     `json:"notes"`
	HarvestRecordedQty *float64   `json:"harvest_recorded_quantity,omitempty"`
	HarvestRecordedAt  *time.Time `json:"harvest_recorded_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type FieldCropFilter struct {
	SeasonID int64
	CropID   int64
	FieldID  string
}

// ComputeYield calculează randamentul la hectar când există producție și suprafață.
func (fc *FieldCrop) ComputeYield() {
	fc.YieldPerHa = nil
	if fc.ProductionTotal == nil || fc.PlantedAreaHa == nil || *fc.PlantedAreaHa <= 0 {
		return
	}
	value := *fc.ProductionTotal / *fc.PlantedAreaHa
	fc.YieldPerHa = &value
}
