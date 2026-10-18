package domain

import "encoding/json"

type Field struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CadastralNumber *string         `json:"cadastral_number,omitempty"`
	AreaHa          *float64        `json:"area_ha,omitempty"`
	Geometry        json.RawMessage `json:"geometry" swaggertype:"object"`

	Auditfields
}

// FieldFilter: căutare după nume sau număr cadastral, interval de suprafață și ordinea listei.
type FieldFilter struct {
	Query    string
	MinArea  *float64
	MaxArea  *float64
	SortBy   string // "name" (implicit) sau "area"
	SortDesc bool
}

// FieldPage este o pagină de terenuri, cu suprafața totală a tuturor terenurilor filtrate.
type FieldPage struct {
	Page[Field]
	TotalAreaHa float64 `json:"total_area_ha"`
}
