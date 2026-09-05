package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"agri-api/internal/domain"
	"agri-api/internal/repository"

	"github.com/lib/pq"
)

func wrapUnique(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return repository.ErrDuplicateEntry
	}
	return err
}

type CropRepo struct {
	db *sql.DB
}

func NewCropRepo(db *sql.DB) *CropRepo {
	return &CropRepo{db: db}
}

// ─── Seasons ─────────────────────────────────────────────────────────────────

const seasonSelect = `
	SELECT id, name, to_char(start_date, 'YYYY-MM-DD'), to_char(end_date, 'YYYY-MM-DD'), is_active, notes, created_at, updated_at
	FROM seasons`

func scanSeason(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Season, error) {
	var season domain.Season
	if err := row.Scan(&season.ID, &season.Name, &season.StartDate, &season.EndDate, &season.IsActive, &season.Notes, &season.CreatedAt, &season.UpdatedAt); err != nil {
		return nil, err
	}
	return &season, nil
}

func (repo *CropRepo) GetSeasons() ([]domain.Season, error) {
	rows, err := repo.db.Query(seasonSelect + " ORDER BY start_date DESC, id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Season{}
	for rows.Next() {
		season, err := scanSeason(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *season)
	}
	return items, rows.Err()
}

func (repo *CropRepo) GetSeasonByID(id int64) (*domain.Season, error) {
	season, err := scanSeason(repo.db.QueryRow(seasonSelect+" WHERE id = $1", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return season, err
}

func (repo *CropRepo) GetActiveOrLatestSeason() (*domain.Season, error) {
	season, err := scanSeason(repo.db.QueryRow(seasonSelect + " ORDER BY is_active DESC, start_date DESC, id DESC LIMIT 1"))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return season, err
}

func (repo *CropRepo) CreateSeason(season *domain.Season) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if season.IsActive {
		if _, err := tx.Exec(`UPDATE seasons SET is_active = FALSE, updated_at = NOW() WHERE is_active`); err != nil {
			return err
		}
	}
	err = tx.QueryRow(`
		INSERT INTO seasons (name, start_date, end_date, is_active, notes)
		VALUES ($1, $2::date, $3::date, $4, $5)
		RETURNING id, created_at, updated_at`,
		season.Name, season.StartDate, season.EndDate, season.IsActive, season.Notes,
	).Scan(&season.ID, &season.CreatedAt, &season.UpdatedAt)
	if err != nil {
		return wrapUnique(err)
	}
	return tx.Commit()
}

func (repo *CropRepo) UpdateSeason(id int64, season *domain.Season) error {
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if season.IsActive {
		if _, err := tx.Exec(`UPDATE seasons SET is_active = FALSE, updated_at = NOW() WHERE is_active AND id <> $1`, id); err != nil {
			return err
		}
	}
	result, err := tx.Exec(`
		UPDATE seasons
		SET name = $1, start_date = $2::date, end_date = $3::date, is_active = $4, notes = $5, updated_at = NOW()
		WHERE id = $6`,
		season.Name, season.StartDate, season.EndDate, season.IsActive, season.Notes, id,
	)
	if err != nil {
		return wrapUnique(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (repo *CropRepo) DeleteSeason(id int64) error {
	result, err := repo.db.Exec(`DELETE FROM seasons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ─── Crops ───────────────────────────────────────────────────────────────────

const cropSelect = `SELECT id, name, code, category, yield_unit, notes, created_at, updated_at FROM crops`

func scanCrop(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Crop, error) {
	var (
		crop domain.Crop
		code sql.NullString
	)
	if err := row.Scan(&crop.ID, &crop.Name, &code, &crop.Category, &crop.YieldUnit, &crop.Notes, &crop.CreatedAt, &crop.UpdatedAt); err != nil {
		return nil, err
	}
	crop.Code = nullStringPtr(code)
	return &crop, nil
}

func (repo *CropRepo) GetCrops() ([]domain.Crop, error) {
	rows, err := repo.db.Query(cropSelect + " ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.Crop{}
	for rows.Next() {
		crop, err := scanCrop(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *crop)
	}
	return items, rows.Err()
}

func (repo *CropRepo) GetCropByID(id int64) (*domain.Crop, error) {
	crop, err := scanCrop(repo.db.QueryRow(cropSelect+" WHERE id = $1", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return crop, err
}

func (repo *CropRepo) CreateCrop(crop *domain.Crop) error {
	err := repo.db.QueryRow(`
		INSERT INTO crops (name, code, category, yield_unit, notes)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5)
		RETURNING id, created_at, updated_at`,
		crop.Name, derefString(crop.Code), crop.Category, crop.YieldUnit, crop.Notes,
	).Scan(&crop.ID, &crop.CreatedAt, &crop.UpdatedAt)
	return wrapUnique(err)
}

func (repo *CropRepo) UpdateCrop(id int64, crop *domain.Crop) error {
	result, err := repo.db.Exec(`
		UPDATE crops
		SET name = $1, code = NULLIF($2, ''), category = $3, yield_unit = $4, notes = $5, updated_at = NOW()
		WHERE id = $6`,
		crop.Name, derefString(crop.Code), crop.Category, crop.YieldUnit, crop.Notes, id,
	)
	if err != nil {
		return wrapUnique(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *CropRepo) DeleteCrop(id int64) error {
	result, err := repo.db.Exec(`DELETE FROM crops WHERE id = $1`, id)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return fmt.Errorf("cultura este folosită pe terenuri și nu poate fi ștearsă")
		}
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ─── Field crops ─────────────────────────────────────────────────────────────

const fieldCropSelect = `
	SELECT
		fc.id, fc.field_id, f.name, f.area_ha,
		fc.season_id, s.name,
		fc.crop_id, c.name, c.yield_unit,
		fc.planted_area_ha,
		to_char(fc.planted_at, 'YYYY-MM-DD'),
		to_char(fc.harvested_at, 'YYYY-MM-DD'),
		fc.production_total, fc.expected_yield_per_ha, fc.notes, fc.created_at, fc.updated_at
	FROM field_crops fc
	JOIN fields f ON f.id = fc.field_id
	JOIN seasons s ON s.id = fc.season_id
	JOIN crops c ON c.id = fc.crop_id`

func scanFieldCrop(row interface {
	Scan(dest ...interface{}) error
}) (*domain.FieldCrop, error) {
	var (
		item        domain.FieldCrop
		fieldArea   sql.NullFloat64
		plantedArea sql.NullFloat64
		plantedAt   sql.NullString
		harvestedAt sql.NullString
		production  sql.NullFloat64
		expected    sql.NullFloat64
	)
	if err := row.Scan(
		&item.ID, &item.FieldID, &item.FieldName, &fieldArea,
		&item.SeasonID, &item.SeasonName,
		&item.CropID, &item.CropName, &item.YieldUnit,
		&plantedArea, &plantedAt, &harvestedAt, &production, &expected, &item.Notes, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.FieldAreaHa = nullFloatPtr(fieldArea)
	item.PlantedAreaHa = nullFloatPtr(plantedArea)
	item.PlantedAt = nullStringPtr(plantedAt)
	item.HarvestedAt = nullStringPtr(harvestedAt)
	item.ProductionTotal = nullFloatPtr(production)
	item.ExpectedYieldPerHa = nullFloatPtr(expected)
	item.ComputeYield()
	return &item, nil
}

func (repo *CropRepo) ListFieldCrops(filter domain.FieldCropFilter) ([]domain.FieldCrop, error) {
	var where strings.Builder
	where.WriteString(" WHERE f.deleted_at IS NULL")
	args := []interface{}{}
	if filter.SeasonID > 0 {
		args = append(args, filter.SeasonID)
		fmt.Fprintf(&where, " AND fc.season_id = $%d", len(args))
	}
	if filter.CropID > 0 {
		args = append(args, filter.CropID)
		fmt.Fprintf(&where, " AND fc.crop_id = $%d", len(args))
	}
	if filter.FieldID != "" {
		args = append(args, filter.FieldID)
		fmt.Fprintf(&where, " AND fc.field_id = $%d", len(args))
	}

	rows, err := repo.db.Query(fieldCropSelect+where.String()+" ORDER BY s.start_date DESC, c.name, f.name", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.FieldCrop{}
	for rows.Next() {
		item, err := scanFieldCrop(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (repo *CropRepo) GetFieldCropByID(id int64) (*domain.FieldCrop, error) {
	item, err := scanFieldCrop(repo.db.QueryRow(fieldCropSelect+" WHERE fc.id = $1", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (repo *CropRepo) CreateFieldCrop(item *domain.FieldCrop) error {
	err := repo.db.QueryRow(`
		INSERT INTO field_crops (field_id, season_id, crop_id, planted_area_ha, planted_at, harvested_at, production_total, expected_yield_per_ha, notes)
		VALUES ($1, $2, $3, $4, NULLIF($5, '')::date, NULLIF($6, '')::date, $7, $8, $9)
		RETURNING id, created_at, updated_at`,
		item.FieldID, item.SeasonID, item.CropID, item.PlantedAreaHa,
		derefString(item.PlantedAt), derefString(item.HarvestedAt),
		item.ProductionTotal, item.ExpectedYieldPerHa, item.Notes,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return wrapUnique(err)
}

func (repo *CropRepo) UpdateFieldCrop(id int64, item *domain.FieldCrop) error {
	result, err := repo.db.Exec(`
		UPDATE field_crops
		SET field_id = $1, season_id = $2, crop_id = $3, planted_area_ha = $4,
		    planted_at = NULLIF($5, '')::date, harvested_at = NULLIF($6, '')::date,
		    production_total = $7, expected_yield_per_ha = $8, notes = $9, updated_at = NOW()
		WHERE id = $10`,
		item.FieldID, item.SeasonID, item.CropID, item.PlantedAreaHa,
		derefString(item.PlantedAt), derefString(item.HarvestedAt),
		item.ProductionTotal, item.ExpectedYieldPerHa, item.Notes, id,
	)
	if err != nil {
		return wrapUnique(err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (repo *CropRepo) DeleteFieldCrop(id int64) error {
	result, err := repo.db.Exec(`DELETE FROM field_crops WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nullFloatPtr(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}
	result := value.Float64
	return &result
}
