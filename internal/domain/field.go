package domain

import "encoding/json"

type Field struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	CadastralNumber *string         `json:"cadastral_number,omitempty"`
	AreaHa          *float64        `json:"area_ha,omitempty"`
	Geometry        json.RawMessage `json:"geometry"`

	Auditfields
}
