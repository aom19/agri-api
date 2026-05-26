package domain

type MachineStatus string

const (
	MachineStatusActive    MachineStatus = "active"
	MachineStatusAvailable MachineStatus = "available"
	MachineStatusInactive  MachineStatus = "inactive"
	MachineStatusInService MachineStatus = "in_service"
	MachineStatusInUse     MachineStatus = "in_use"
)

type Machine struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Status      MachineStatus `json:"status"`
	Description string        `json:"description,omitempty"`
}
