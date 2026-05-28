package domain

type OperatorStatus string

const (
	OperatorStatusActive   OperatorStatus = "active"
	OperatorStatusInactive OperatorStatus = "inactive"
)

type Operator struct {
	ID     int64          `json:"id"`
	Name   string         `json:"name"`
	Status OperatorStatus `json:"status"`

	Auditfields
}
