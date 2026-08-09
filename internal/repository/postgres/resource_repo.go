package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type ResourceRepo struct {
	db *sql.DB
}

func NewResourceRepo(db *sql.DB) *ResourceRepo {
	return &ResourceRepo{db: db}
}

const resourceSelectWithType = `
	SELECT r.id, r.name, r.resource_type_id, r.price_per_unit, COALESCE(r.notes, ''),
	       r.created_at, r.updated_at,
	       rt.id, rt.name, rt.category, rt.default_unit, rt.created_at, rt.updated_at
	FROM resources r
	JOIN resource_types rt ON rt.id = r.resource_type_id`

func scanResource(scanner interface {
	Scan(...any) error
}) (domain.Resource, error) {
	var res domain.Resource
	var rt domain.ResourceType
	err := scanner.Scan(
		&res.ID, &res.Name, &res.ResourceTypeID, &res.PricePerUnit, &res.Notes,
		&res.CreatedAt, &res.UpdatedAt,
		&rt.ID, &rt.Name, &rt.Category, &rt.DefaultUnit, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		return domain.Resource{}, err
	}
	res.ResourceType = &rt
	return res, nil
}

func (r *ResourceRepo) GetAll() ([]domain.Resource, error) {
	rows, err := r.db.Query(resourceSelectWithType + ` ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var result []domain.Resource
	for rows.Next() {
		res, err := scanResource(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	return result, nil
}

func (r *ResourceRepo) GetByID(id int64) (*domain.Resource, error) {
	res, err := scanResource(r.db.QueryRow(resourceSelectWithType+` WHERE r.id = $1`, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *ResourceRepo) Create(res *domain.Resource) error {
	return r.db.QueryRow(
		`INSERT INTO resources (name, resource_type_id, price_per_unit, notes)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, created_at, updated_at`,
		res.Name, res.ResourceTypeID, res.PricePerUnit, res.Notes,
	).Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
}

func (r *ResourceRepo) Update(id int64, res *domain.Resource) error {
	_, err := r.db.Exec(
		`UPDATE resources
		 SET name=$1, resource_type_id=$2, price_per_unit=$3, notes=$4, updated_at=NOW()
		 WHERE id=$5`,
		res.Name, res.ResourceTypeID, res.PricePerUnit, res.Notes, id,
	)
	return err
}

func (r *ResourceRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM resources WHERE id=$1`, id)
	return err
}
