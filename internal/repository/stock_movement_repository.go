package repository

import (
	"database/sql"

	"agri-api/internal/domain"
)

// StockLock este stocul unei resurse, blocat pentru actualizare în cadrul unei tranzacții.
type StockLock struct {
	ResourceID   int64
	ResourceName string
	Quantity     float64
	Minimum      float64
	PriceUnit    float64
	Category     string
}

type StockMovementRepository interface {
	// LockStock citește stocul resursei cu FOR UPDATE, în tranzacție; nil dacă resursa nu există.
	LockStock(tx *sql.Tx, resourceID int64) (*StockLock, error)
	// ApplyMovement inserează mișcarea și actualizează cantitatea resursei, în tranzacție.
	ApplyMovement(tx *sql.Tx, movement *domain.StockMovement) error
	List(filter domain.StockMovementFilter) ([]domain.StockMovement, error)
	// ListFuelStocks returnează resursele de combustibil, cu stocul lor.
	ListFuelStocks() ([]domain.FuelStock, error)
}
