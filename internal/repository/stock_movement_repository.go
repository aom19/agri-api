package repository

import (
	"database/sql"

	"agri-api/internal/domain"
)

// StockLock este stocul blocat pentru actualizare în cadrul unei tranzacții.
type StockLock struct {
	StockID    int64
	ResourceID int64
	Quantity   float64
	Minimum    float64
	PriceUnit  float64
}

type StockMovementRepository interface {
	// LockStockByID / LockStockByResource citesc stocul cu FOR UPDATE, în tranzacție.
	LockStockByID(tx *sql.Tx, stockID int64) (*StockLock, error)
	LockStockByResource(tx *sql.Tx, resourceID int64) (*StockLock, error)
	// ApplyMovement inserează mișcarea și actualizează cantitatea stocului, în tranzacție.
	ApplyMovement(tx *sql.Tx, movement *domain.StockMovement) error
	List(filter domain.StockMovementFilter) ([]domain.StockMovement, error)
}
