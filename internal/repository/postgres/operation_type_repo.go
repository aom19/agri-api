package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type OperationTypeRepo struct {
	db *sql.DB
}

func NewOperationTypeRepo(db *sql.DB) *OperationTypeRepo {
	return &OperationTypeRepo{db: db}
}

func (r *OperationTypeRepo) GetAll() ([]domain.OperationType, error) {
	rows, err := r.db.Query(
		`SELECT id, code, name, COALESCE(description, ''), created_at, updated_at
		 FROM operation_types WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.OperationType
	for rows.Next() {
		var ot domain.OperationType
		if err := rows.Scan(&ot.ID, &ot.Code, &ot.Name, &ot.Description, &ot.CreatedAt, &ot.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, ot)
	}
	return result, rows.Err()
}

func (r *OperationTypeRepo) GetByID(id int64) (*domain.OperationType, error) {
	var ot domain.OperationType
	err := r.db.QueryRow(
		`SELECT id, code, name, COALESCE(description, ''), created_at, updated_at
		 FROM operation_types WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&ot.ID, &ot.Code, &ot.Name, &ot.Description, &ot.CreatedAt, &ot.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ot, nil
}

func (r *OperationTypeRepo) Create(ot *domain.OperationType) error {
	return r.db.QueryRow(
		`INSERT INTO operation_types (code, name, description)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		ot.Code, ot.Name, ot.Description,
	).Scan(&ot.ID, &ot.CreatedAt, &ot.UpdatedAt)
}

func (r *OperationTypeRepo) Update(id int64, ot *domain.OperationType) error {
	return r.db.QueryRow(
		`UPDATE operation_types
		 SET code=$1, name=$2, description=$3, updated_at=NOW()
		 WHERE id=$4 AND deleted_at IS NULL
		 RETURNING updated_at`,
		ot.Code, ot.Name, ot.Description, id,
	).Scan(&ot.UpdatedAt)
}

func (r *OperationTypeRepo) Delete(id int64) error {
	_, err := r.db.Exec(
		`UPDATE operation_types SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`, id)
	return err
}
