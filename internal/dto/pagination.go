package dto

type PaginationQuery struct {
	Page       int
	Limit      int
	Status     string
	OperatorID string
	MachineID  string
	SortBy     string
	Order      string
}
