package domain

import "encoding/json"

type Permission struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`

	Auditfields
}
