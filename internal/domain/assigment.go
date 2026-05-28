package domain

import "time"

type AssigmentStatus string

const (
	AssigmentStatusActive AssigmentStatus = "active"
	AssigmentStatusClosed AssigmentStatus = "closed"
)

type Assigment struct {
	ID         int64           `json:"id"`
	MachineID  int64           `json:"machine_id"`
	OperatorID int64           `json:"operator_id"`
	StartDate  time.Time       `json:"start_date"`
	EndDate    *time.Time      `json:"end_date,omitempty"`
	Status     AssigmentStatus `json:"status"`

	Auditfields
}
