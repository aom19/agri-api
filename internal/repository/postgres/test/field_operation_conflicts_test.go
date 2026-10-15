package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
)

// Suprapunerile: două lucrări planificate sau în lucru care împart operatorul, mașina sau
// echipamentul pe intervale suprapuse sunt în conflict. Intervalele lipite, lucrările anulate și
// resursele diferite nu sunt; o lucrare în lucru ține resursele până acum, chiar depășită.
func TestFieldOperationConflicts(t *testing.T) {
	db := requireDB(t)
	repo := postgres.NewFieldOperationRepo(db)
	field := insertField(t, db, "Tarla Mare")
	machine := insertMachine(t, db, "JD 6R", "TR-01")
	ion, ana := insertOperator(t, db, "Ion"), insertOperator(t, db, "Ana")

	day := time.Now().UTC().Truncate(time.Hour).Add(48 * time.Hour)
	at := func(h int) time.Time { return day.Add(time.Duration(h) * time.Hour) }
	op := func(status string, operator, machineID *int64, from, to time.Time) int64 {
		return insertFieldOperation(t, db, fieldOperation{
			FieldID: field, OperationType: "soil_preparation", Status: status,
			OperatorID: operator, MachineID: machineID, PlannedStart: from, PlannedEnd: ptr(to), AreaPlannedHa: 10,
		})
	}
	a := op("planned", &ion, &machine, at(8), at(12))
	b := op("planned", &ana, &machine, at(11), at(15)) // aceeași mașină, suprapus cu A
	c := op("planned", &ion, nil, at(12), at(14))      // același operator, lipit de A: fără conflict
	op("canceled", &ion, &machine, at(9), at(10))      // anulată: nu ocupă nimic
	now := time.Now().UTC()
	late := op("in_progress", &ana, nil, now.Add(-26*time.Hour), now.Add(-24*time.Hour)) // depășită, încă în lucru

	list, err := repo.GetAll(repository.FieldOperationFilter{})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[int64]dto.FieldOperationResponse{}
	for _, item := range list {
		byID[item.ID] = item
	}
	expectConflict := func(id, other int64, resource string) {
		t.Helper()
		got := byID[id].Conflicts
		if len(got) != 1 || got[0].OperationID != other || len(got[0].Resources) != 1 || got[0].Resources[0] != resource {
			t.Errorf("lucrarea %d: aștept conflict cu %d pe %s, am %+v", id, other, resource, got)
		}
	}
	expectConflict(a, b, dto.ConflictResourceMachine)
	expectConflict(b, a, dto.ConflictResourceMachine)
	for _, id := range []int64{c, late} {
		if got := byID[id].Conflicts; len(got) != 0 {
			t.Errorf("lucrarea %d nu are conflicte: %+v", id, got)
		}
	}
	if got := byID[a].Conflicts[0]; got.FieldName != "Tarla Mare" || got.OperationTypeName == "" || !got.StartAt.Equal(at(11)) || !got.EndAt.Equal(at(15)) {
		t.Errorf("detaliile conflictului: %+v", got)
	}

	// formularul: Ion la 11:30 se suprapune cu A, dar nu și la editarea lui A
	query := repository.FieldOperationConflictQuery{OperatorID: &ion, StartAt: at(11).Add(30 * time.Minute), EndAt: ptr(at(11).Add(45 * time.Minute))}
	if got, err := repo.FindConflicts(query); err != nil || len(got) != 1 || got[0].OperationID != a {
		t.Errorf("Ion la 11:30: %v, %+v", err, got)
	}
	query.ExcludeID = a
	if got, err := repo.FindConflicts(query); err != nil || len(got) != 0 {
		t.Errorf("editarea lui A nu e în conflict cu ea însăși: %v, %+v", err, got)
	}
	// fără sfârșit, intervalul e de o oră: 12:30 nu atinge A (8–12), dar atinge C (12–14)
	if got, err := repo.FindConflicts(repository.FieldOperationConflictQuery{OperatorID: &ion, StartAt: at(12).Add(30 * time.Minute)}); err != nil || len(got) != 1 || got[0].OperationID != c {
		t.Errorf("Ion la 12:30 fără sfârșit: %v, %+v", err, got)
	}
	// Ana e ocupată acum de lucrarea depășită
	if got, err := repo.FindConflicts(repository.FieldOperationConflictQuery{OperatorID: &ana, StartAt: now.Add(-time.Minute), EndAt: ptr(now.Add(time.Hour))}); err != nil || len(got) != 1 || got[0].OperationID != late {
		t.Errorf("Ana acum: %v, %+v", err, got)
	}

	// pornirea: contează doar lucrările în lucru, indiferent de interval
	if got, err := repo.InProgressConflicts(repository.FieldOperationConflictQuery{ExcludeID: b, OperatorID: &ana, MachineID: &machine}); err != nil || len(got) != 1 || got[0].OperationID != late || got[0].Resources[0] != dto.ConflictResourceOperator {
		t.Errorf("B nu poate porni cât Ana e în lucru: %v, %+v", err, got)
	}
	if got, err := repo.InProgressConflicts(repository.FieldOperationConflictQuery{ExcludeID: a, OperatorID: &ion, MachineID: &machine}); err != nil || len(got) != 0 {
		t.Errorf("A poate porni: %v, %+v", err, got)
	}
}
