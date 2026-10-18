package domain

import "time"

type AuditEntry struct {
	ID         int64                  `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	EntityName *string                `json:"entity_name,omitempty"`
	Action     string                 `json:"action"`
	ActorID    *int64                 `json:"actor_id,omitempty"`
	ActorName  *string                `json:"actor_name,omitempty"`
	Changes    map[string]interface{} `json:"changes,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditFilter restrânge jurnalul la un tip de entitate și, opțional, la o entitate.
type AuditFilter struct {
	EntityType string
	EntityID   string
}
