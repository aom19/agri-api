package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ImplementCompatibilityRepo struct {
	db *sql.DB
}

func NewImplementCompatibilityRepo(db *sql.DB) *ImplementCompatibilityRepo {
	return &ImplementCompatibilityRepo{db: db}
}

func (r *ImplementCompatibilityRepo) GetAll() ([]domain.ImplementCompatibility, error) {
	rows, err := r.db.Query(
		`SELECT id, machine_type, implement_type, created_at, updated_at
		 FROM implement_compatibilities
		 ORDER BY machine_type, implement_type`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.ImplementCompatibility
	for rows.Next() {
		var c domain.ImplementCompatibility
		if err := rows.Scan(&c.ID, &c.MachineType, &c.ImplementType, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
