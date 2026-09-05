package postgres

import (
	"database/sql"
	"fmt"

	"agri-api/internal/domain"
)

type WeatherSnapshotRepo struct {
	db *sql.DB
}

func NewWeatherSnapshotRepo(db *sql.DB) *WeatherSnapshotRepo {
	return &WeatherSnapshotRepo{db: db}
}

// Insert salvează observația; întoarce false dacă exista deja una identică (același teren și moment).
func (repo *WeatherSnapshotRepo) Insert(snapshot *domain.WeatherSnapshot) (bool, error) {
	err := repo.db.QueryRow(`
		INSERT INTO weather_snapshots (
			field_id, latitude, longitude, observed_at, temperature_c, condition, weather_code,
			humidity_percent, wind_speed_kmh, wind_direction_deg, precipitation_mm, cloud_cover_percent, source
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (field_id, observed_at) DO NOTHING
		RETURNING id`,
		snapshot.FieldID, snapshot.Latitude, snapshot.Longitude, snapshot.ObservedAt, snapshot.TemperatureC,
		snapshot.Condition, snapshot.WeatherCode, snapshot.HumidityPercent, snapshot.WindSpeedKmh,
		snapshot.WindDirectionDeg, snapshot.PrecipitationMM, snapshot.CloudCoverPercent, snapshot.Source,
	).Scan(&snapshot.ID)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func weatherWhere(filter domain.ReportFilter) (string, []interface{}) {
	where := "ws.observed_at >= $1 AND ws.observed_at < $2"
	args := []interface{}{filter.From, filter.To}
	if filter.FieldID != "" {
		args = append(args, filter.FieldID)
		where += fmt.Sprintf(" AND ws.field_id = $%d", len(args))
	}
	return where, args
}

// GetDailyAggregates agregă observațiile pe zile. Precipitațiile sunt însumate pe zi
// și mediate între terenuri; temperaturile sunt medii/min/max ale tuturor observațiilor.
func (repo *WeatherSnapshotRepo) GetDailyAggregates(filter domain.ReportFilter) ([]domain.WeatherDailyAggregate, error) {
	where, args := weatherWhere(filter)
	query := fmt.Sprintf(`
		WITH per_field AS (
			SELECT
				ws.field_id,
				to_char(ws.observed_at, 'YYYY-MM-DD') AS day,
				AVG(ws.temperature_c) AS avg_t,
				MIN(ws.temperature_c) AS min_t,
				MAX(ws.temperature_c) AS max_t,
				COALESCE(SUM(ws.precipitation_mm), 0) AS precip,
				AVG(ws.humidity_percent) AS humidity,
				AVG(ws.wind_speed_kmh) AS wind,
				COUNT(*) AS samples
			FROM weather_snapshots ws
			WHERE %s
			GROUP BY ws.field_id, to_char(ws.observed_at, 'YYYY-MM-DD')
		)
		SELECT day, AVG(avg_t), MIN(min_t), MAX(max_t), AVG(precip), AVG(humidity), AVG(wind), SUM(samples)
		FROM per_field
		GROUP BY day
		ORDER BY day`, where)

	rows, err := repo.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.WeatherDailyAggregate{}
	for rows.Next() {
		var (
			item     domain.WeatherDailyAggregate
			humidity sql.NullFloat64
			wind     sql.NullFloat64
		)
		if err := rows.Scan(&item.Day, &item.AvgTemperatureC, &item.MinTemperatureC, &item.MaxTemperatureC, &item.PrecipitationMM, &humidity, &wind, &item.Samples); err != nil {
			return nil, err
		}
		item.AvgHumidity = nullFloatPtr(humidity)
		item.AvgWindKmh = nullFloatPtr(wind)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (repo *WeatherSnapshotRepo) GetLatestPerField() ([]domain.WeatherSnapshot, error) {
	rows, err := repo.db.Query(`
		SELECT DISTINCT ON (ws.field_id)
			ws.id, ws.field_id, f.name, ws.latitude, ws.longitude, ws.observed_at, ws.temperature_c, ws.condition,
			ws.weather_code, ws.humidity_percent, ws.wind_speed_kmh, ws.wind_direction_deg, ws.precipitation_mm,
			ws.cloud_cover_percent, ws.source
		FROM weather_snapshots ws
		JOIN fields f ON f.id = ws.field_id
		WHERE f.deleted_at IS NULL
		ORDER BY ws.field_id, ws.observed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []domain.WeatherSnapshot{}
	for rows.Next() {
		var (
			item      domain.WeatherSnapshot
			fieldName sql.NullString
			code      sql.NullInt64
			humidity  sql.NullInt64
			wind      sql.NullFloat64
			windDir   sql.NullInt64
			precip    sql.NullFloat64
			cloud     sql.NullInt64
		)
		if err := rows.Scan(&item.ID, &item.FieldID, &fieldName, &item.Latitude, &item.Longitude, &item.ObservedAt, &item.TemperatureC, &item.Condition,
			&code, &humidity, &wind, &windDir, &precip, &cloud, &item.Source); err != nil {
			return nil, err
		}
		item.FieldName = nullStringPtr(fieldName)
		item.WeatherCode = nullIntPtr(code)
		item.HumidityPercent = nullIntPtr(humidity)
		item.WindSpeedKmh = nullFloatPtr(wind)
		item.WindDirectionDeg = nullIntPtr(windDir)
		item.PrecipitationMM = nullFloatPtr(precip)
		item.CloudCoverPercent = nullIntPtr(cloud)
		items = append(items, item)
	}
	return items, rows.Err()
}

// CountInPeriod returnează numărul de observații și numărul de terenuri acoperite în interval.
func (repo *WeatherSnapshotRepo) CountInPeriod(filter domain.ReportFilter) (int, int, error) {
	where, args := weatherWhere(filter)
	var samples, fields int
	err := repo.db.QueryRow(fmt.Sprintf(`
		SELECT COUNT(*), COUNT(DISTINCT ws.field_id) FROM weather_snapshots ws WHERE %s`, where), args...,
	).Scan(&samples, &fields)
	return samples, fields, err
}

func nullIntPtr(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	result := int(value.Int64)
	return &result
}
