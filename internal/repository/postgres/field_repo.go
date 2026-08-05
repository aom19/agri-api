package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

// FieldRepo implementeaza repository.FieldRepository folosind PostgreSQL.
type FieldRepo struct {
	db *sql.DB
}

func NewFieldRepo(db *sql.DB) *FieldRepo {
	return &FieldRepo{db: db}
}

func (repo *FieldRepo) Create(field *domain.Field) error {
	query := `
		INSERT INTO fields (name, cadastral_number, area_ha, geometry)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return repo.db.QueryRow(query, field.Name, field.CadastralNumber, field.AreaHa, field.Geometry).Scan(
		&field.ID,
		&field.CreatedAt,
		&field.UpdatedAt,
	)
}

func (repo *FieldRepo) GetAll() ([]domain.Field, error) {
	rows, err := repo.db.Query(`
		SELECT id, name, cadastral_number, area_ha, geometry, created_at, updated_at
		FROM fields
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var fields []domain.Field
	for rows.Next() {
		var field domain.Field
		var geometry []byte
		if err := rows.Scan(
			&field.ID,
			&field.Name,
			&field.CadastralNumber,
			&field.AreaHa,
			&geometry,
			&field.CreatedAt,
			&field.UpdatedAt,
		); err != nil {
			return nil, err
		}
		field.Geometry = geometry
		fields = append(fields, field)
	}

	return fields, nil
}

func (repo *FieldRepo) GetByID(id string) (*domain.Field, error) {
	query := `
		SELECT id, name, cadastral_number, area_ha, geometry, created_at, updated_at
		FROM fields
		WHERE id = $1 AND deleted_at IS NULL`

	var field domain.Field
	var geometry []byte
	err := repo.db.QueryRow(query, id).Scan(
		&field.ID,
		&field.Name,
		&field.CadastralNumber,
		&field.AreaHa,
		&geometry,
		&field.CreatedAt,
		&field.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	field.Geometry = geometry
	return &field, nil
}

func (repo *FieldRepo) Update(id string, field *domain.Field) error {
	query := `
		UPDATE fields
		SET name = $1,
		    cadastral_number = $2,
		    area_ha = $3,
		    geometry = $4,
		    updated_at = NOW()
		WHERE id = $5 AND deleted_at IS NULL`

	_, err := repo.db.Exec(query, field.Name, field.CadastralNumber, field.AreaHa, field.Geometry, id)
	return err
}

func (repo *FieldRepo) Delete(id string) error {
	_, err := repo.db.Exec(`UPDATE fields SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id)
	return err
}
