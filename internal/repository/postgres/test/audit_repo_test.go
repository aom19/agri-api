package postgres_test

import (
	"strconv"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
)

// Jurnalul de audit leagă fiecare intrare de entitatea ei ca să afișeze numele; intrările
// rămase de la modulele șterse (ex. „assignment”) trebuie să apară în continuare, fără nume.
func TestAuditLog_ResolvesEntityNames(t *testing.T) {
	db := requireDB(t)
	if _, err := db.Exec(`TRUNCATE audit_log RESTART IDENTITY`); err != nil {
		t.Fatal(err)
	}
	repo := postgres.NewAuditRepo(db)
	machineID := insertMachine(t, db, "Tractor audit", "TR-AUD")

	for _, entry := range []domain.AuditEntry{
		{EntityType: "machine", EntityID: strconv.FormatInt(machineID, 10), Action: "create"},
		{EntityType: "assignment", EntityID: "7", Action: "delete"},
	} {
		entry := entry
		if err := repo.Log(&entry); err != nil {
			t.Fatal(err)
		}
	}

	entries, err := repo.GetAll(50, "", "")
	if err != nil {
		t.Fatal(err)
	}
	byType := map[string]domain.AuditEntry{}
	for _, entry := range entries {
		byType[entry.EntityType] = entry
	}
	if len(entries) != 2 {
		t.Fatalf("mă așteptam la 2 intrări, am %d", len(entries))
	}
	if name := byType["machine"].EntityName; name == nil || *name != "Tractor audit" {
		t.Errorf("numele mașinii nu e rezolvat: %v", name)
	}
	if legacy := byType["assignment"]; legacy.EntityID != "7" || legacy.EntityName != nil {
		t.Errorf("intrarea veche de tip assignment: %+v", legacy)
	}
}
