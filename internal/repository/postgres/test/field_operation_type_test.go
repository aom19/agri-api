package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
)

// T14: tipul unei operațiuni cu template este tipul template-ului. Baza refuză un tip propriu
// lângă template, iar lista, filtrul și rapoartele urmează template-ul când acesta se schimbă.
func TestFieldOperations_TypeComesFromTemplate(t *testing.T) {
	db := requireDB(t)
	now := time.Now().UTC()
	field := insertField(t, db, "Lot tip")
	templateID := insertID(t, db, `INSERT INTO operation_templates (name, operation_type) VALUES ('Erbicidare', 'spraying') RETURNING id`)
	withTemplate := insertFieldOperation(t, db, fieldOperation{
		FieldID: field, TemplateID: &templateID, Status: "in_progress",
		PlannedStart: now.Add(-2 * time.Hour), PlannedEnd: ptr(now.Add(-time.Hour)),
	})
	insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationType: "seeding", Status: "planned", PlannedStart: now})

	if _, err := db.Exec(`UPDATE field_operations SET operation_type = 'seeding' WHERE id = $1`, withTemplate); err == nil {
		t.Error("o operațiune cu template nu poate avea tip propriu")
	}
	if _, err := db.Exec(`INSERT INTO field_operations (field_id, status) VALUES ($1, 'planned')`, field); err == nil {
		t.Error("o operațiune fără template are nevoie de tip")
	}

	ops := postgres.NewFieldOperationRepo(db)
	got, err := ops.GetByID(withTemplate)
	if err != nil || got.OperationType != "spraying" || got.OperationTypeName != "Stropire" {
		t.Fatalf("tipul din template: %v, %+v", err, got)
	}
	overdue, err := ops.GetOverdueInProgress(now)
	if err != nil || len(overdue) != 1 || overdue[0].OperationTypeName != "Stropire" {
		t.Errorf("notificarea de depășire folosește tipul din template: %v, %+v", err, overdue)
	}
	sprayed, err := ops.GetAll(repository.FieldOperationFilter{OperationType: "spraying"})
	if err != nil || len(sprayed) != 1 || sprayed[0].ID != withTemplate {
		t.Errorf("filtrul pe tip: %v, %+v", err, sprayed)
	}

	if _, err := db.Exec(`UPDATE operation_templates SET operation_type = 'seeding' WHERE id = $1`, templateID); err != nil {
		t.Fatal(err)
	}
	filter := domain.ReportFilter{From: now.AddDate(0, 0, -1), To: now.AddDate(0, 0, 1)}
	byType, err := postgres.NewReportRepo(db).GetOperationsByType(filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(byType) != 1 || byType[0].OperationType != domain.OperationTypeSeeding || byType[0].OperationTypeName != "Semănat" || byType[0].Total != 2 {
		t.Errorf("după schimbarea template-ului, ambele operațiuni sunt Semănat: %+v", byType)
	}
	filter.OperationType = domain.OperationTypeSpraying
	if rows, err := postgres.NewReportRepo(db).GetOperationRows(filter, 10); err != nil || len(rows) != 0 {
		t.Errorf("nu mai există operațiuni de stropire: %v, %+v", err, rows)
	}
}
