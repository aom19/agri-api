package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ResourceTypeRepo struct {
	db *sql.DB
}

func NewResourceTypeRepo(db *sql.DB) *ResourceTypeRepo {
	return &ResourceTypeRepo{db: db}
}

func (r *ResourceTypeRepo) GetAll() ([]domain.ResourceType, error) {
	rows, err := r.db.Query(
		`SELECT id, name, category, default_unit, created_at, updated_at
		 FROM resource_types
		 ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.ResourceType
	for rows.Next() {
		var rt domain.ResourceType
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.Category, &rt.DefaultUnit, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *ResourceTypeRepo) GetByID(id int64) (*domain.ResourceType, error) {
	var rt domain.ResourceType
	err := r.db.QueryRow(
		`SELECT id, name, category, default_unit, created_at, updated_at
		 FROM resource_types WHERE id = $1`,
		id,
	).Scan(&rt.ID, &rt.Name, &rt.Category, &rt.DefaultUnit, &rt.CreatedAt, &rt.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *ResourceTypeRepo) Create(rt *domain.ResourceType) error {
	return r.db.QueryRow(
		`INSERT INTO resource_types (name, category, default_unit)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		rt.Name, rt.Category, rt.DefaultUnit,
	).Scan(&rt.ID, &rt.CreatedAt, &rt.UpdatedAt)
}

func (r *ResourceTypeRepo) Update(id int64, rt *domain.ResourceType) error {
	_, err := r.db.Exec(
		`UPDATE resource_types
		 SET name=$1, category=$2, default_unit=$3, updated_at=NOW()
		 WHERE id=$4`,
		rt.Name, rt.Category, rt.DefaultUnit, id,
	)
	return err
}

func (r *ResourceTypeRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM resource_types WHERE id=$1`, id)
	return err
}
