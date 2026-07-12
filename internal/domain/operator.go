package domain

type OperatorStatus string

const (
	OperatorStatusActive   OperatorStatus = "active"
	OperatorStatusInactive OperatorStatus = "inactive"
)

type Operator struct {
	ID                  int64          `json:"id"`
	Name                string         `json:"name"`
	Phone               string         `json:"phone"`
	Email               string         `json:"email"`
	Notes               string         `json:"notes"`
	Status              OperatorStatus `json:"status"`
	AllowedMachineTypes []MachineType  `json:"allowed_machine_types"`

	Auditfields
}
