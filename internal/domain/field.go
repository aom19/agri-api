package domain

import "encoding/json"

type Field struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	AreaHa   *float64        `json:"area_ha,omitempty"`
	Geometry json.RawMessage `json:"geometry"`

	Auditfields
}
