package usecase

import (
	"context"
	"encoding/json"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

// WeatherSnapshotMonitor salvează periodic condițiile meteo curente pentru fiecare teren
// cu geometrie definită (în centrul poligonului), construind istoricul meteo al fermei.
type WeatherSnapshotMonitor struct {
	fields    repository.FieldRepository
	weather   *WeatherService
	snapshots repository.WeatherSnapshotRepository
	log       OverdueMonitorLogger
	interval  time.Duration
}

func NewWeatherSnapshotMonitor(
	fields repository.FieldRepository,
	weather *WeatherService,
	snapshots repository.WeatherSnapshotRepository,
	log OverdueMonitorLogger,
	interval time.Duration,
) *WeatherSnapshotMonitor {
	if interval <= 0 {
		interval = time.Hour
	}
	return &WeatherSnapshotMonitor{fields: fields, weather: weather, snapshots: snapshots, log: log, interval: interval}
}

func (m *WeatherSnapshotMonitor) Start(ctx context.Context) {
	m.log.Infof("Weather snapshot monitor started (interval %s)", m.interval)
	m.runSafely(ctx)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.log.Infof("Weather snapshot monitor stopped")
			return
		case <-ticker.C:
			m.runSafely(ctx)
		}
	}
}

func (m *WeatherSnapshotMonitor) runSafely(ctx context.Context) {
	saved, err := m.RunOnce(ctx)
	if err != nil {
		m.log.Errorf("Weather snapshot run failed: %v", err)
		return
	}
	if saved > 0 {
		m.log.Infof("Weather snapshot run: %d observation(s) saved", saved)
	}
}

// RunOnce colectează o observație pentru fiecare teren și returnează câte au fost salvate.
func (m *WeatherSnapshotMonitor) RunOnce(ctx context.Context) (int, error) {
	fields, err := m.fields.GetAll()
	if err != nil {
		return 0, err
	}

	saved := 0
	for _, field := range fields {
		lat, lon, ok := polygonCentroid(field.Geometry)
		if !ok {
			continue
		}
		requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		current, err := m.weather.GetCurrentFor(requestCtx, WeatherLocation{Latitude: lat, Longitude: lon, Name: field.Name})
		cancel()
		if err != nil {
			m.log.Errorf("Weather snapshot for field %s failed: %v", field.Name, err)
			continue
		}

		fieldID := field.ID
		code := current.WeatherCode
		snapshot := &domain.WeatherSnapshot{
			FieldID:           &fieldID,
			Latitude:          lat,
			Longitude:         lon,
			ObservedAt:        current.ObservedAt,
			TemperatureC:      float64(current.TemperatureC),
			Condition:         current.Condition,
			WeatherCode:       &code,
			HumidityPercent:   current.HumidityPercent,
			WindSpeedKmh:      current.WindSpeedKmh,
			WindDirectionDeg:  current.WindDirectionDeg,
			PrecipitationMM:   current.PrecipitationMM,
			CloudCoverPercent: current.CloudCoverPercent,
			Source:            current.Source,
		}
		if snapshot.ObservedAt.IsZero() {
			snapshot.ObservedAt = time.Now().UTC().Truncate(time.Minute)
		}
		inserted, err := m.snapshots.Insert(snapshot)
		if err != nil {
			m.log.Errorf("Weather snapshot save for field %s failed: %v", field.Name, err)
			continue
		}
		if inserted {
			saved++
		}
	}
	return saved, nil
}

// polygonCentroid întoarce media vârfurilor inelului exterior al unui poligon GeoJSON.
func polygonCentroid(geometry json.RawMessage) (float64, float64, bool) {
	if len(geometry) == 0 {
		return 0, 0, false
	}
	var polygon struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}
	if err := json.Unmarshal(geometry, &polygon); err != nil || len(polygon.Coordinates) == 0 {
		return 0, 0, false
	}
	ring := polygon.Coordinates[0]
	if len(ring) > 1 {
		first, last := ring[0], ring[len(ring)-1]
		if len(first) >= 2 && len(last) >= 2 && first[0] == last[0] && first[1] == last[1] {
			ring = ring[:len(ring)-1]
		}
	}
	if len(ring) == 0 {
		return 0, 0, false
	}
	var sumLon, sumLat float64
	count := 0
	for _, point := range ring {
		if len(point) < 2 {
			continue
		}
		sumLon += point[0]
		sumLat += point[1]
		count++
	}
	if count == 0 {
		return 0, 0, false
	}
	return sumLat / float64(count), sumLon / float64(count), true
}
