package postgres_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

// T11: o cerere directă cu o mașină sau un echipament de alt tip decât cele din template e
// respinsă, la creare și la editare; fără template, orice mașină e acceptată.
func TestFieldOperations_TemplateAssetCompatibility(t *testing.T) {
	db := requireDB(t)
	field := insertField(t, db, "Lot semănat")
	opType := insertID(t, db, `INSERT INTO operation_types (code, name) VALUES ('sowing', 'Semănat') RETURNING id`)
	templateID := insertID(t, db, `INSERT INTO operation_templates (name, operation_type_id) VALUES ('Semănat porumb', $1) RETURNING id`, opType)
	for _, query := range []string{
		`INSERT INTO template_machine_types (template_id, machine_type) VALUES ($1, 'tractor')`,
		`INSERT INTO template_implement_types (template_id, implement_type) VALUES ($1, 'seeder')`,
	} {
		if _, err := db.Exec(query, templateID); err != nil {
			t.Fatal(err)
		}
	}
	tractor := insertMachine(t, db, "Tractor", "TR-1")
	combine := insertID(t, db, `INSERT INTO machines (name, code, type) VALUES ('Combină', 'CB-1', 'combine') RETURNING id`)
	seeder := insertID(t, db, `INSERT INTO implements (name, code, type) VALUES ('Semănătoare', 'IMP-1', 'seeder') RETURNING id`)
	header := insertID(t, db, `INSERT INTO implements (name, code, type) VALUES ('Header', 'IMP-2', 'header') RETURNING id`)

	svc := usecase.NewFieldOperationService(postgres.NewFieldOperationRepo(db))
	operation := func(template, machine, implement *int64) *domain.FieldOperation {
		return &domain.FieldOperation{
			FieldID: field, OperationTypeID: opType,
			OperationTemplateID: template, MachineID: machine, ImplementID: implement,
		}
	}

	created, err := svc.Create(operation(&templateID, &tractor, &seeder))
	if err != nil {
		t.Fatalf("tractor + semănătoare: %v", err)
	}
	if _, err := svc.Create(operation(&templateID, &combine, &seeder)); !errors.Is(err, usecase.ErrFieldOperationMachineIncompatible) {
		t.Errorf("combină pe template de tractor: %v", err)
	}
	if _, err := svc.Create(operation(&templateID, &tractor, &header)); !errors.Is(err, usecase.ErrFieldOperationImplementIncompatible) {
		t.Errorf("header pe template de semănătoare: %v", err)
	}
	if _, err := svc.Update(created.ID, operation(&templateID, &combine, nil)); !errors.Is(err, usecase.ErrFieldOperationMachineIncompatible) {
		t.Errorf("editare cu mașină nepotrivită: %v", err)
	}
	if _, err := svc.Create(operation(nil, &combine, &header)); err != nil {
		t.Errorf("fără template, orice mașină și echipament: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM field_operations`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("operațiuni salvate: %d, aștept 2 (cele respinse nu se salvează)", count)
	}
	got, err := svc.GetByID(created.ID)
	if err != nil || got.MachineID == nil || *got.MachineID != tractor {
		t.Errorf("editarea respinsă nu trebuie să schimbe mașina: %v, %+v", err, got)
	}
}
