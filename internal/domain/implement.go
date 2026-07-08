package domain

type ImplementType string

const (
	ImplementTypePlow               ImplementType = "plow"
	ImplementTypeDiscHarrow         ImplementType = "disc_harrow"
	ImplementTypeCultivator         ImplementType = "cultivator"
	ImplementTypeSeeder             ImplementType = "seeder"
	ImplementTypeFertilizerSpreader ImplementType = "fertilizer_spreader"
	ImplementTypeSprayer            ImplementType = "sprayer"
	ImplementTypeTrailer            ImplementType = "trailer"
	ImplementTypeHeader             ImplementType = "header"
	ImplementTypeOther              ImplementType = "other"
)

type Implement struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Code         string        `json:"code"`
	Type         ImplementType `json:"type"`
	Brand        string        `json:"brand,omitempty"`
	Model        string        `json:"model,omitempty"`
	Year         *int          `json:"year,omitempty"`
	WorkingWidth *float64      `json:"working_width,omitempty"`
	Capacity     *float64      `json:"capacity,omitempty"`
	Status       AssetStatus   `json:"status"`
	Notes        string        `json:"notes,omitempty"`

	Auditfields
}

type ImplementCompatibility struct {
	ID            int64         `json:"id"`
	MachineType   MachineType   `json:"machine_type"`
	ImplementType ImplementType `json:"implement_type"`

	Auditfields
}

var validImplementTypes = map[ImplementType]struct{}{
	ImplementTypePlow:               {},
	ImplementTypeDiscHarrow:         {},
	ImplementTypeCultivator:         {},
	ImplementTypeSeeder:             {},
	ImplementTypeFertilizerSpreader: {},
	ImplementTypeSprayer:            {},
	ImplementTypeTrailer:            {},
	ImplementTypeHeader:             {},
	ImplementTypeOther:              {},
}

func (implementType ImplementType) IsValid() bool {
	_, ok := validImplementTypes[implementType]
	return ok
}
