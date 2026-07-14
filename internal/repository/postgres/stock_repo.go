package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type StockRepo struct {
	db *sql.DB
}

func NewStockRepo(db *sql.DB) *StockRepo {
	return &StockRepo{db: db}
}

func (r *StockRepo) GetAll() ([]domain.Stock, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.resource_id, s.quantity, s.minimum_quantity, s.created_at, s.updated_at,
		       res.id, res.name, res.resource_type_id, res.unit, res.price_per_unit,
		       COALESCE(res.notes, ''), res.created_at, res.updated_at
		FROM stocks s
		JOIN resources res ON res.id = s.resource_id
		ORDER BY res.name`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.Stock
	for rows.Next() {
		var s domain.Stock
		var res domain.Resource
		if err := rows.Scan(
			&s.ID, &s.ResourceID, &s.Quantity, &s.MinimumQuantity, &s.CreatedAt, &s.UpdatedAt,
			&res.ID, &res.Name, &res.ResourceTypeID, &res.Unit, &res.PricePerUnit,
			&res.Notes, &res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, err
		}
		s.Resource = &res
		result = append(result, s)
	}
	return result, nil
}

func (r *StockRepo) GetByID(id int64) (*domain.Stock, error) {
	var s domain.Stock
	err := r.db.QueryRow(
		`SELECT id, resource_id, quantity, minimum_quantity, created_at, updated_at
		 FROM stocks WHERE id=$1`,
		id,
	).Scan(&s.ID, &s.ResourceID, &s.Quantity, &s.MinimumQuantity, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StockRepo) GetByResourceID(resourceID int64) (*domain.Stock, error) {
	var s domain.Stock
	err := r.db.QueryRow(
		`SELECT id, resource_id, quantity, minimum_quantity, created_at, updated_at
		 FROM stocks WHERE resource_id=$1`,
		resourceID,
	).Scan(&s.ID, &s.ResourceID, &s.Quantity, &s.MinimumQuantity, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *StockRepo) Create(s *domain.Stock) error {
	return r.db.QueryRow(
		`INSERT INTO stocks (resource_id, quantity, minimum_quantity)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		s.ResourceID, s.Quantity, s.MinimumQuantity,
	).Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

func (r *StockRepo) Update(id int64, s *domain.Stock) error {
	_, err := r.db.Exec(
		`UPDATE stocks
		 SET quantity=$1, minimum_quantity=$2, updated_at=NOW()
		 WHERE id=$3`,
		s.Quantity, s.MinimumQuantity, id,
	)
	return err
}

func (r *StockRepo) DecrementQuantity(tx *sql.Tx, resourceID int64, qty float64) error {
	_, err := tx.Exec(
		`UPDATE stocks
		 SET quantity = quantity - $1, updated_at = NOW()
		 WHERE resource_id = $2`,
		qty, resourceID,
	)
	return err
}
