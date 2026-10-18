package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
	"fmt"
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
	fields, _, _, err := repo.List(domain.FieldFilter{}, nil)
	return fields, err
}

func scanField(rows *sql.Rows) (domain.Field, error) {
	var field domain.Field
	var geometry []byte
	err := rows.Scan(&field.ID, &field.Name, &field.CadastralNumber, &field.AreaHa, &geometry, &field.CreatedAt, &field.UpdatedAt)
	field.Geometry = geometry
	return field, err
}

// List întoarce terenurile filtrate (nume sau număr cadastral, interval de suprafață), în ordinea
// cerută, și suprafața totală a tuturor celor filtrate; cu page, doar pagina cerută.
func (repo *FieldRepo) List(filter domain.FieldFilter, page *domain.PageRequest) ([]domain.Field, int, float64, error) {
	var args []interface{}
	query := `
		SELECT id, name, cadastral_number, area_ha, geometry, created_at, updated_at
		FROM fields
		WHERE deleted_at IS NULL` + textSearchClause(filter.Query, &args, "name", "cadastral_number")
	if filter.MinArea != nil {
		args = append(args, *filter.MinArea)
		query += fmt.Sprintf(" AND area_ha >= $%d", len(args))
	}
	if filter.MaxArea != nil {
		args = append(args, *filter.MaxArea)
		query += fmt.Sprintf(" AND area_ha <= $%d", len(args))
	}
	direction := "ASC"
	if filter.SortDesc {
		direction = "DESC"
	}
	orderBy := "name " + direction + ", id"
	if filter.SortBy == "area" {
		orderBy = "area_ha " + direction + " NULLS LAST, name, id"
	}
	fields, total, err := listPage(repo.db, query, args, orderBy, page, scanField)
	if err != nil {
		return nil, 0, 0, err
	}
	var totalArea float64
	if err := repo.db.QueryRow("SELECT COALESCE(SUM(area_ha), 0) FROM ("+query+") AS filtered", args...).Scan(&totalArea); err != nil {
		return nil, 0, 0, err
	}
	return fields, total, totalArea, nil
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
