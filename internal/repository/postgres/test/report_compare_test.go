package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
)

// Comparația: lucrările unui sezon (prin cultura lor, fără filtru de dată) și numele entităților.
func TestReportRepo_SeasonFilterAndCompareNames(t *testing.T) {
	db := requireDB(t)
	repo := postgres.NewReportRepo(db)
	field := insertField(t, db, "Lot sezon")
	crop := insertCrop(t, db, "Porumb")
	old := insertSeason(t, db, "Sezon vechi", "2024-08-01", "2025-07-31")
	current := insertSeason(t, db, "Sezon nou", "2025-08-01", "2026-07-31")
	oldCrop := insertID(t, db, `INSERT INTO field_crops (field_id, crop_id, season_id, status) VALUES ($1, $2, $3, 'closed') RETURNING id`, field, crop, old)

	inOld := insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "completed", PlannedStart: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)})
	insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "planned", PlannedStart: time.Now()})
	if _, err := db.Exec(`UPDATE field_operations SET field_crop_id = $1 WHERE id = $2`, oldCrop, inOld); err != nil {
		t.Fatal(err)
	}

	metrics, err := repo.GetOperationsMetrics(domain.ReportFilter{SeasonID: old})
	if err != nil || metrics.OperationsTotal != 1 || metrics.OperationsCompleted != 1 {
		t.Fatalf("lucrările sezonului vechi: %v, %+v", err, metrics)
	}
	if metrics, err = repo.GetOperationsMetrics(domain.ReportFilter{SeasonID: current}); err != nil || metrics.OperationsTotal != 0 {
		t.Fatalf("sezonul nou nu are lucrări legate: %v, %+v", err, metrics)
	}

	operator := insertOperator(t, db, "Ion")
	machine := insertMachine(t, db, "John Deere", "JD1")
	if name, err := repo.GetCompareName(domain.ReportCompareOperators, operator); err != nil || name != "Ion" {
		t.Errorf("numele operatorului: %v, %q", err, name)
	}
	if name, err := repo.GetCompareName(domain.ReportCompareMachines, machine); err != nil || name != "John Deere" {
		t.Errorf("numele mașinii: %v, %q", err, name)
	}
	if name, err := repo.GetCompareName(domain.ReportCompareOperators, machine+999); err != nil || name != "" {
		t.Errorf("entitatea inexistentă: %v, %q", err, name)
	}
}
