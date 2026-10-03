package postgres_test

import (
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

// T6: o lucrare cu 100 l raportați scade exact 100 l din stoc, iar rapoartele arată 100 l.
func TestCompletion_FuelHasASingleSource(t *testing.T) {
	db := requireDB(t)
	now := time.Now().UTC()
	field := insertField(t, db, "Lot motorină")
	opType := insertID(t, db, `INSERT INTO operation_types (code, name) VALUES ('plowing', 'Arat') RETURNING id`)
	machine := insertMachine(t, db, "Tractor", "TR-1")
	fuelStockID, fuelResourceID := insertStock(t, db, "Motorină", 1000, 0, 7)

	// norma de motorină din șablon (8 l/ha × 10 ha = 80 l) e înlocuită de valoarea raportată
	templateID := insertID(t, db, `INSERT INTO operation_templates (name, operation_type_id) VALUES ('Arat standard', $1) RETURNING id`, opType)
	if _, err := db.Exec(`INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) VALUES ($1, $2, 8)`, templateID, fuelResourceID); err != nil {
		t.Fatal(err)
	}
	opID := insertFieldOperation(t, db, fieldOperation{
		FieldID: field, OperationTypeID: opType, MachineID: &machine, Status: "in_progress",
		PlannedStart: now.Add(-3 * time.Hour), AreaPlannedHa: 10,
	})
	if _, err := db.Exec(`UPDATE field_operations SET operation_template_id = $1 WHERE id = $2`, templateID, opID); err != nil {
		t.Fatal(err)
	}

	ops := postgres.NewFieldOperationRepo(db)
	svc := usecase.NewFieldOperationCompletionService(db, ops, postgres.NewStockMovementRepo(db))
	result, err := svc.Complete(opID, domain.FieldOperationCompletion{
		ConsumeFromTemplate: true,
		FuelUsedL:           ptr(100.0),
		FuelResourceID:      &fuelResourceID,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := stockQuantity(t, db, fuelStockID); got != 900 {
		t.Errorf("stocul de motorină trebuia să scadă cu 100 l, a rămas %.3f", got)
	}
	if result.Operation.FuelUsedL == nil || *result.Operation.FuelUsedL != 100 {
		t.Errorf("operațiunea trebuie să arate 100 l, am %v", result.Operation.FuelUsedL)
	}

	reports := postgres.NewReportRepo(db)
	filter := domain.ReportFilter{From: now.AddDate(0, 0, -1), To: now.AddDate(0, 0, 1)}
	metrics, err := reports.GetOperationsMetrics(filter)
	if err != nil {
		t.Fatal(err)
	}
	if metrics.FuelUsedL != 100 {
		t.Errorf("raportul Sumar arată %.3f l, mă așteptam la 100", metrics.FuelUsedL)
	}
	machines, err := reports.GetMachineRows(filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(machines) != 1 || machines[0].FuelUsedL != 100 {
		t.Errorf("raportul Flotă: %+v", machines)
	}
	assertMovementsExplainStocks(t, db)
}

// Estimarea folosește aceeași normă ca finalizarea cu consume_from_template.
func TestCompletion_EstimateMatchesTemplateConsumption(t *testing.T) {
	db := requireDB(t)
	field := insertField(t, db, "Lot estimare")
	opType := insertID(t, db, `INSERT INTO operation_types (code, name) VALUES ('fertilizing', 'Fertilizare') RETURNING id`)
	stockID, resourceID := insertStock(t, db, "Uree", 500, 0, 2)
	if _, err := db.Exec(`UPDATE resource_types SET category = 'fertilizer', default_unit = 'kg'`); err != nil {
		t.Fatal(err)
	}
	templateID := insertID(t, db, `INSERT INTO operation_templates (name, operation_type_id) VALUES ('Fertilizare uree', $1) RETURNING id`, opType)
	if _, err := db.Exec(`INSERT INTO template_resources (template_id, resource_id, quantity_per_unit) VALUES ($1, $2, 1.5)`, templateID, resourceID); err != nil {
		t.Fatal(err)
	}
	opID := insertFieldOperation(t, db, fieldOperation{FieldID: field, OperationTypeID: opType, Status: "in_progress", PlannedStart: time.Now(), AreaPlannedHa: 12})
	if _, err := db.Exec(`UPDATE field_operations SET operation_template_id = $1 WHERE id = $2`, templateID, opID); err != nil {
		t.Fatal(err)
	}

	svc := usecase.NewFieldOperationCompletionService(db, postgres.NewFieldOperationRepo(db), postgres.NewStockMovementRepo(db))
	estimate, err := svc.Estimate(opID, ptr(9.5))
	if err != nil {
		t.Fatal(err)
	}
	if len(estimate.Items) != 1 || estimate.Items[0].Quantity != 14.25 || estimate.Items[0].Unit != "kg" || estimate.Items[0].Category != "fertilizer" {
		t.Fatalf("estimare: %+v", estimate)
	}
	if len(estimate.FuelResources) != 0 {
		t.Errorf("ureea nu e combustibil: %+v", estimate.FuelResources)
	}

	if _, err := svc.Complete(opID, domain.FieldOperationCompletion{ConsumeFromTemplate: true, AreaCompletedHa: ptr(9.5)}, nil); err != nil {
		t.Fatal(err)
	}
	if got := stockQuantity(t, db, stockID); got != 500-estimate.Items[0].Quantity {
		t.Errorf("finalizarea a scăzut altă cantitate decât estimarea: stoc %.4f", got)
	}
}
