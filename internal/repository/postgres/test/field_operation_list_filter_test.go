package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
)

// Filtrul pe intervalul planificat și ordonarea „cele mai apropiate primele” (panoul operatorului).
func TestFieldOperations_PlannedRangeAndSort(t *testing.T) {
	db := requireDB(t)
	repo := postgres.NewFieldOperationRepo(db)
	field := insertField(t, db, "Lot interval")

	day := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	late := insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "planned", PlannedStart: day.Add(15 * time.Hour)})
	early := insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "planned", PlannedStart: day.Add(8 * time.Hour)})
	insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "planned", PlannedStart: day.Add(-time.Hour)})
	insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "soil_preparation", Status: "planned", PlannedStart: day.Add(24 * time.Hour)})

	from, to := day, day.Add(24*time.Hour)
	items, err := repo.GetAll(repository.FieldOperationFilter{PlannedFrom: &from, PlannedTo: &to, SortPlannedAsc: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != early || items[1].ID != late {
		t.Fatalf("lucrările zilei, cele mai apropiate primele: %+v", items)
	}

	all, err := repo.GetAll(repository.FieldOperationFilter{})
	if err != nil || len(all) != 4 || all[0].PlannedStartAt.Before(*all[3].PlannedStartAt) {
		t.Fatalf("implicit, cele mai recente primele: %v, %+v", err, all)
	}
}

// Căutarea text din lista lucrărilor: teren, note, cultură etc.
func TestFieldOperations_Search(t *testing.T) {
	db := requireDB(t)
	repo := postgres.NewFieldOperationRepo(db)
	north := insertField(t, db, "Lot Nord")
	south := insertField(t, db, "Lot Sud")
	insertFieldOperation(t, db, fieldOperation{FieldID: north, OperationType: "soil_preparation", Status: "planned", PlannedStart: time.Now()})
	insertFieldOperation(t, db, fieldOperation{FieldID: south, OperationType: "soil_preparation", Status: "planned", PlannedStart: time.Now()})

	items, err := repo.GetAll(repository.FieldOperationFilter{Search: "nord"})
	if err != nil || len(items) != 1 || items[0].FieldName != "Lot Nord" {
		t.Fatalf("căutare după teren: %v, %+v", err, items)
	}
	if items, err = repo.GetAll(repository.FieldOperationFilter{Search: "100%"}); err != nil || len(items) != 0 {
		t.Fatalf("caracterele speciale sunt căutate literal: %v, %+v", err, items)
	}
}
