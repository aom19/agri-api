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

const cropSelect = `SELECT id, name, code, category, yield_unit, notes, harvest_resource_id, created_at, updated_at FROM crops`

func scanCrop(row interface {
	Scan(dest ...interface{}) error
}) (*domain.Crop, error) {
	var (
		crop    domain.Crop
		code    sql.NullString
		harvest sql.NullInt64
	)
	if err := row.Scan(&crop.ID, &crop.Name, &code, &crop.Category, &crop.YieldUnit, &crop.Notes, &harvest, &crop.CreatedAt, &crop.UpdatedAt); err != nil {
		return nil, err
	}
	crop.Code = nullStringPtr(code)
	if harvest.Valid {
		value := harvest.Int64
		crop.HarvestResourceID = &value
	}
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
		fc.season_id, s.name, to_char(s.start_date, 'YYYY-MM-DD'), to_char(s.end_date, 'YYYY-MM-DD'),
		fc.crop_id, c.name, c.yield_unit,
		fc.planted_area_ha,
		to_char(fc.planted_at, 'YYYY-MM-DD'),
		to_char(fc.harvested_at, 'YYYY-MM-DD'),
		fc.production_total, fc.expected_yield_per_ha, fc.notes,
		fc.harvest_recorded_quantity, fc.harvest_recorded_at,
		fc.created_at, fc.updated_at
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
		recordedQty sql.NullFloat64
		recordedAt  sql.NullTime
	)
	if err := row.Scan(
		&item.ID, &item.FieldID, &item.FieldName, &fieldArea,
		&item.SeasonID, &item.SeasonName, &item.SeasonStart, &item.SeasonEnd,
		&item.CropID, &item.CropName, &item.YieldUnit,
		&plantedArea, &plantedAt, &harvestedAt, &production, &expected, &item.Notes,
		&recordedQty, &recordedAt, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	item.HarvestRecordedQty = nullFloatPtr(recordedQty)
	if recordedAt.Valid {
		value := recordedAt.Time
		item.HarvestRecordedAt = &value
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

// RelinkOperations leagă operațiunile terenurilor din sezon de cultura corespunzătoare
// (după data planificată) și desface legăturile către culturi care nu mai corespund.
func (repo *CropRepo) RelinkOperations(seasonID int64) error {
	if _, err := repo.db.Exec(`
		UPDATE field_operations fo
		SET field_crop_id = NULL
		FROM field_crops fc
		JOIN seasons s ON s.id = fc.season_id
		WHERE fo.field_crop_id = fc.id AND s.id = $1
		  AND (fo.field_id <> fc.field_id
		       OR COALESCE(fo.planned_start_at, fo.created_at) < s.start_date
		       OR COALESCE(fo.planned_start_at, fo.created_at) >= s.end_date + INTERVAL '1 day')`, seasonID); err != nil {
		return err
	}
	_, err := repo.db.Exec(`
		UPDATE field_operations fo
		SET field_crop_id = fc.id
		FROM field_crops fc
		JOIN seasons s ON s.id = fc.season_id
		WHERE s.id = $1
		  AND fo.deleted_at IS NULL
		  AND fo.field_id = fc.field_id
		  AND fo.field_crop_id IS NULL
		  AND COALESCE(fo.planned_start_at, fo.created_at) >= s.start_date
		  AND COALESCE(fo.planned_start_at, fo.created_at) < s.end_date + INTERVAL '1 day'`, seasonID)
	return err
}

// EnsureHarvestResource garantează că cultura are o resursă de stoc pentru recoltă
// (tip de resursă cu categoria "harvest" și unitatea culturii) și întoarce id-ul resursei.
func (repo *CropRepo) EnsureHarvestResource(tx *sql.Tx, crop *domain.Crop) (int64, error) {
	if crop.HarvestResourceID != nil {
		var exists int64
		err := tx.QueryRow(`SELECT id FROM resources WHERE id = $1`, *crop.HarvestResourceID).Scan(&exists)
		if err == nil {
			return exists, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}

	unit := strings.TrimSpace(crop.YieldUnit)
	if unit == "" {
		unit = "t"
	}
	var typeID int64
	err := tx.QueryRow(`SELECT id FROM resource_types WHERE category = 'harvest' AND default_unit = $1 ORDER BY id LIMIT 1`, unit).Scan(&typeID)
	if err == sql.ErrNoRows {
		typeName := "Recoltă"
		if unit != "t" {
			typeName = fmt.Sprintf("Recoltă (%s)", unit)
		}
		err = tx.QueryRow(`INSERT INTO resource_types (name, category, default_unit) VALUES ($1, 'harvest', $2) RETURNING id`, typeName, unit).Scan(&typeID)
	}
	if err != nil {
		return 0, err
	}

	var resourceID int64
	if err := tx.QueryRow(`
		INSERT INTO resources (name, resource_type_id, price_per_unit, notes)
		VALUES ($1, $2, 0, $3)
		RETURNING id`,
		"Recoltă "+crop.Name, typeID, "Creată automat la înregistrarea recoltei. Setează prețul unitar pentru valoarea stocului.",
	).Scan(&resourceID); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`UPDATE crops SET harvest_resource_id = $1, updated_at = NOW() WHERE id = $2`, resourceID, crop.ID); err != nil {
		return 0, err
	}
	return resourceID, nil
}

// EnsureStock creează rândul de stoc al resursei dacă nu există.
func (repo *CropRepo) EnsureStock(tx *sql.Tx, resourceID int64) error {
	_, err := tx.Exec(`
		INSERT INTO stocks (resource_id, quantity, minimum_quantity)
		SELECT $1, 0, 0
		WHERE NOT EXISTS (SELECT 1 FROM stocks WHERE resource_id = $1)`, resourceID)
	return err
}

func (repo *CropRepo) MarkHarvestRecorded(tx *sql.Tx, fieldCropID int64, quantity float64) error {
	_, err := tx.Exec(`
		UPDATE field_crops
		SET harvest_recorded_quantity = $1, harvest_recorded_at = NOW(), updated_at = NOW()
		WHERE id = $2`, quantity, fieldCropID)
	return err
}
