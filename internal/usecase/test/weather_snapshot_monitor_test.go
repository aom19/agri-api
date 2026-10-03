package usecase_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

func TestPolygonCentroid(t *testing.T) {
	invalid := []string{"", "{", `{"type":"Polygon","coordinates":[]}`, `{"type":"Polygon","coordinates":[[[1]]]}`}
	for _, g := range invalid {
		if _, _, ok := usecase.PolygonCentroid(json.RawMessage(g)); ok {
			t.Errorf("%q: nu trebuie să aibă centroid", g)
		}
	}
	lat, lon, ok := usecase.PolygonCentroid(json.RawMessage(`{"type":"Polygon","coordinates":[[[28,46],[30,46],[30,48],[28,46]]]}`))
	if !ok || lat < 46.66 || lat > 46.67 || lon < 29.33 || lon > 29.34 {
		t.Errorf("inel închis: %v %v %v", lat, lon, ok)
	}
	lat, lon, ok = usecase.PolygonCentroid(json.RawMessage(`{"type":"Polygon","coordinates":[[[28,46],[30,46],[30,48]]]}`))
	if !ok || lat < 46.66 || lat > 46.67 {
		t.Errorf("inel deschis: %v %v %v", lat, lon, ok)
	}
	if lat, lon, ok = usecase.PolygonCentroid(json.RawMessage(`{"type":"Polygon","coordinates":[[[28,46]]]}`)); !ok || lat != 46 || lon != 28 {
		t.Errorf("un singur punct: %v %v %v", lat, lon, ok)
	}
}

func TestWeatherSnapshotMonitor_RunOnce(t *testing.T) {
	fields := &fieldRepoMock{getAll: func() ([]domain.Field, error) {
		return []domain.Field{
			{ID: "f1", Name: "Lot 1", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[28,46],[30,46],[30,48],[28,46]]]}`)},
			{ID: "f2", Name: "Fără geometrie"},
		}, nil
	}}
	calls := 0
	weather := usecase.NewWeatherService("")
	weather.SetHTTPClient(fakeWeatherClient(&calls, func() *http.Response { return jsonResponse(200, openMeteoBody) }, nil))
	var inserted []*domain.WeatherSnapshot
	snapshots := &weatherSnapshotRepoMock{insert: func(s *domain.WeatherSnapshot) (bool, error) { inserted = append(inserted, s); return true, nil }}
	log := &testLogger{}
	monitor := usecase.NewWeatherSnapshotMonitor(fields, weather, snapshots, log, 0)
	if monitor.Interval() != time.Hour {
		t.Errorf("intervalul implicit trebuie să fie 1h, am %s", monitor.Interval())
	}

	saved, err := monitor.RunOnce(context.Background())
	if err != nil || saved != 1 || len(inserted) != 1 {
		t.Fatalf("RunOnce: %v, salvate=%d", err, saved)
	}
	snap := inserted[0]
	if *snap.FieldID != "f1" || snap.TemperatureC != 22 || snap.Condition != "Ploaie slabă" || *snap.WeatherCode != 61 || snap.Source != "Open-Meteo" {
		t.Errorf("observație greșită: %+v", snap)
	}

	// deja există observația → nu se numără
	snapshots.insert = func(*domain.WeatherSnapshot) (bool, error) { return false, nil }
	if saved, err = monitor.RunOnce(context.Background()); err != nil || saved != 0 {
		t.Errorf("duplicat: %v, %d", err, saved)
	}
	snapshots.insert = func(*domain.WeatherSnapshot) (bool, error) { return false, errors.New("db down") }
	if saved, err = monitor.RunOnce(context.Background()); err != nil || saved != 0 || log.errorCount() != 1 {
		t.Errorf("eroare la salvare: %v, %d, erori=%d", err, saved, log.errorCount())
	}

	// furnizorul meteo nu răspunde
	weather.SetHTTPClient(fakeWeatherClient(&calls, func() *http.Response { return jsonResponse(500, "") }, nil))
	weather.ResetCache()
	if saved, err = monitor.RunOnce(context.Background()); err != nil || saved != 0 || log.errorCount() != 2 {
		t.Errorf("eroare meteo: %v, %d, erori=%d", err, saved, log.errorCount())
	}

	fields.getAll = func() ([]domain.Field, error) { return nil, errors.New("db down") }
	if _, err := monitor.RunOnce(context.Background()); err == nil {
		t.Error("eroarea de citire a terenurilor trebuie propagată")
	}
}

func TestWeatherSnapshotMonitor_Start(t *testing.T) {
	calls := 0
	fields := &fieldRepoMock{getAll: func() ([]domain.Field, error) {
		calls++
		if calls == 1 {
			return []domain.Field{{ID: "f1", Geometry: json.RawMessage(`{"type":"Polygon","coordinates":[[[28,46],[30,46],[30,48],[28,46]]]}`)}}, nil
		}
		return nil, errors.New("db down")
	}}
	httpCalls := 0
	weather := usecase.NewWeatherService("")
	weather.SetHTTPClient(fakeWeatherClient(&httpCalls, func() *http.Response { return jsonResponse(200, openMeteoBody) }, nil))
	log := &testLogger{}
	monitor := usecase.NewWeatherSnapshotMonitor(fields, weather, &weatherSnapshotRepoMock{}, log, 5*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	monitor.Start(ctx)

	joined := strings.Join(log.infos, "\n")
	if calls < 2 || log.errorCount() == 0 || !strings.Contains(joined, "1 observation(s) saved") || !strings.Contains(joined, "stopped") {
		t.Errorf("monitorul trebuie să ruleze imediat și periodic: apeluri=%d erori=%d log=%s", calls, log.errorCount(), joined)
	}
}
