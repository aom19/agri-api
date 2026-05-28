package dto

import "time"

type AssigmentResponse struct {
	ID           int64      `json:"id"`
	MachineID    int64      `json:"machine_id"`
	MachineName  string     `json:"machine_name"`
	OperatorID   int64      `json:"operator_id"`
	OperatorName string     `json:"operator_name"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Status       string     `json:"status"`
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type PaginatedAssignmentsResponse struct {
	Data []AssigmentResponse `json:"data"`
	Meta PaginationMeta      `json:"meta"`
}
