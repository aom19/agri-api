package postgres_test

import (
	"database/sql"
	"os"
	"testing"

	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

const seedFile = "../../../../seeds/seed.sql"

// runSeed rulează seeds/seed.sql ca `make seed`: tot fișierul, într-o singură tranzacție.
func runSeed(t *testing.T, db *sql.DB) {
	t.Helper()
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	content, err := os.ReadFile(seedFile)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(string(content)); err != nil {
		t.Fatalf("seed.sql: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

// T21: după `make seed`, paginile Culturi, Rapoarte și Alocări au date. Seed-ul rulează de două
// ori, ca după o a doua rulare pe aceeași bază.
func TestSeed_FillsCropsReportsAndAssignments(t *testing.T) {
	db := requireDB(t)
	runSeed(t, db)
	runSeed(t, db)

	reports := usecase.NewReportService(postgres.NewReportRepo(db)).WithCrops(postgres.NewCropRepo(db))

	// Culturi și raportul pe culturi: sezonul activ are culturi pe terenuri și producție
	crops, err := reports.GetCrops(0, "")
	if err != nil {
		t.Fatal(err)
	}
	if crops.Season == nil || len(crops.Items) == 0 || crops.Totals.ProductionTotal <= 0 || len(crops.Seasons) < 2 {
		t.Errorf("culturi: sezon %+v, %d culturi pe terenuri, producție %.1f, %d sezoane",
			crops.Season, len(crops.Items), crops.Totals.ProductionTotal, len(crops.Seasons))
	}

	// Rapoarte, pe intervalul implicit (ultimele 30 de zile, deci fără lucrările planificate de acum încolo)
	summary, err := reports.GetSummary(usecase.ReportQuery{})
	if err != nil {
		t.Fatal(err)
	}
	current := summary.Current
	if current.OperationsCompleted == 0 || current.OperationsInProgress == 0 ||
		current.OverdueOperations == 0 || current.FuelUsedL <= 0 || current.RealCost <= 0 || current.EstimatedCost <= 0 {
		t.Errorf("sumarul pe ultimele 30 de zile e incomplet: %+v", current)
	}

	// Alocări: operațiunile planificate sau în lucru, fiecare cu mașină și operator
	operations := postgres.NewFieldOperationRepo(db)
	for _, status := range []string{"planned", "in_progress"} {
		active, err := operations.GetAll(repository.FieldOperationFilter{Status: status})
		if err != nil {
			t.Fatal(err)
		}
		if len(active) == 0 {
			t.Errorf("nu există operațiuni %s", status)
		}
		for _, op := range active {
			if op.MachineID == nil || op.OperatorID == nil {
				t.Errorf("operațiunea #%d nu are mașină sau operator", op.ID)
			}
		}
	}

	// stocul rămâne explicat de mișcări, iar recolta înregistrată a intrat în stoc
	assertMovementsExplainStocks(t, db)
	var harvestIn int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE field_crop_id IS NOT NULL AND movement_type = 'in'`).Scan(&harvestIn); err != nil || harvestIn == 0 {
		t.Errorf("recolta înregistrată trebuie să intre în stoc: %d, %v", harvestIn, err)
	}
}
