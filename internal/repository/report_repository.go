package repository

import "agri-api/internal/domain"

type ReportRepository interface {
	GetOperationsMetrics(filter domain.ReportFilter) (*domain.ReportOperationsMetrics, error)
	GetInventorySnapshot() (*domain.ReportInventorySnapshot, error)
	GetOperationsTimeline(filter domain.ReportFilter, granularity string) ([]domain.ReportTimeBucket, error)
	GetOperationsByType(filter domain.ReportFilter) ([]domain.ReportOperationTypeStat, error)
	GetOperationRows(filter domain.ReportFilter, limit int) ([]domain.ReportOperationRow, error)
	GetFieldRows(filter domain.ReportFilter) ([]domain.ReportFieldRow, error)
	GetMachineStatusCounts() ([]domain.ReportNamedCount, error)
	GetImplementStatusCounts() ([]domain.ReportNamedCount, error)
	GetMachinesByType() ([]domain.ReportNamedCount, error)
	GetMachinesByFuel() ([]domain.ReportNamedCount, error)
	GetMachinesByYear() ([]domain.ReportNamedCount, error)
	GetMachineRows(filter domain.ReportFilter) ([]domain.ReportMachineRow, error)
	GetImplementRows(filter domain.ReportFilter) ([]domain.ReportImplementRow, error)
	GetOperatorRows(filter domain.ReportFilter) ([]domain.ReportOperatorRow, error)
	GetStockRows() ([]domain.ReportStockRow, error)
	GetEstimatedConsumption(filter domain.ReportFilter) ([]domain.ReportResourceConsumption, error)
	GetRealConsumption(filter domain.ReportFilter) ([]domain.ReportResourceConsumption, error)
	GetMovementTotals(filter domain.ReportFilter) (*domain.ReportMovementTotals, error)
	GetFieldCropRows(seasonID int64, fieldID string) ([]domain.ReportFieldCropRow, error)
}
