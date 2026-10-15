package postgres_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

// T24 „cultura deschisă pe teren”: lucrarea se leagă explicit (sau de cultura deschisă a
// terenului), pe un teren e cel mult o cultură deschisă, iar închiderea culturii sau modificarea
// sezonului nu mută nicio lucrare deja salvată.

func insertSeason(t *testing.T, db *sql.DB, name, start, end string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO seasons (name, start_date, end_date) VALUES ($1, $2, $3) RETURNING id`, name, start, end)
}

func insertCrop(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO crops (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id`, name)
}

func operationCrop(t *testing.T, db *sql.DB, opID int64) *int64 {
	t.Helper()
	var crop sql.NullInt64
	if err := db.QueryRow(`SELECT field_crop_id FROM field_operations WHERE id = $1`, opID).Scan(&crop); err != nil {
		t.Fatal(err)
	}
	if !crop.Valid {
		return nil
	}
	return &crop.Int64
}

func plannedOp(field string, opType domain.OperationType, day string) *domain.FieldOperation {
	at, _ := time.Parse("2006-01-02", day)
	return &domain.FieldOperation{FieldID: field, OperationType: opType, Status: domain.FieldOperationStatusPlanned, PlannedStartAt: &at, AreaPlannedHa: ptr(10.0)}
}

func TestFieldCrop_OnlyOneOpenPerField(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	field := insertField(t, db, "Tarla Mica")
	season := insertSeason(t, db, "Ciclu toamnă 2025/2026", "2025-08-01", "2026-07-15")
	wheat, corn := insertCrop(t, db, "Grâu"), insertCrop(t, db, "Porumb")

	first, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: season, CropID: wheat})
	if err != nil || first.Status != domain.FieldCropStatusOpen {
		t.Fatalf("prima cultură: %v, %+v", err, first)
	}
	// serviciul refuză a doua cultură deschisă, cu numele celei deschise
	_, err = svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: season, CropID: corn})
	if !errors.Is(err, usecase.ErrFieldHasOpenCrop) || !strings.Contains(err.Error(), "Grâu") {
		t.Fatalf("a doua cultură deschisă pe același teren: %v", err)
	}
	// baza o refuză și ea, dacă serviciul ar fi ocolit
	if _, err := db.Exec(`INSERT INTO field_crops (field_id, season_id, crop_id) VALUES ($1, $2, $3)`, field, season, corn); err == nil {
		t.Fatal("indexul unic parțial trebuie să refuze a doua cultură deschisă")
	}
	// două culturi în același sezon pe același teren sunt permise, dacă una e închisă
	if _, err := svc.CloseFieldCrop(first.ID, domain.FieldCropClose{Status: domain.FieldCropStatusCompromised, ClosedAt: "2026-02-28", Notes: ptr("îngheț")}); err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: season, CropID: corn})
	if err != nil {
		t.Fatalf("după închidere, cultura nouă în același sezon: %v", err)
	}
	// redeschiderea primei e refuzată cât timp a doua e deschisă
	if _, err := svc.ReopenFieldCrop(first.ID); !errors.Is(err, usecase.ErrFieldHasOpenCrop) {
		t.Errorf("redeschidere cu alta deschisă: %v", err)
	}
	if _, err := svc.CloseFieldCrop(second.ID, domain.FieldCropClose{Status: domain.FieldCropStatusClosed, ClosedAt: "2026-10-10", ProductionTotal: ptr(440.0)}); err != nil {
		t.Fatal(err)
	}
	reopened, err := svc.ReopenFieldCrop(first.ID)
	if err != nil || reopened.Status != domain.FieldCropStatusOpen || reopened.ClosedAt != nil {
		t.Errorf("ReopenFieldCrop: %v, %+v", err, reopened)
	}
	// compromisă: producția e 0, data e salvată; închisă: harvested_at ia data închiderii
	closed, _ := svc.ListFieldCrops(domain.FieldCropFilter{FieldID: field, Status: domain.FieldCropStatusClosed})
	if len(closed) != 1 || closed[0].HarvestedAt == nil || *closed[0].HarvestedAt != "2026-10-10" || *closed[0].ProductionTotal != 440 {
		t.Errorf("cultura închisă: %+v", closed)
	}
}

func TestFieldCrop_NewOperationGoesToTheOpenCrop(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	ops := postgres.NewFieldOperationRepo(db)
	field := insertField(t, db, "Tarla Mica")
	other := insertField(t, db, "Tarla Sud")
	opType := domain.OperationTypeSoilPreparation
	oldSeason := insertSeason(t, db, "Ciclu toamnă 2024/2025", "2024-08-01", "2025-07-15")
	season := insertSeason(t, db, "Ciclu toamnă 2025/2026", "2025-08-01", "2026-07-15")
	rapeseed, wheat := insertCrop(t, db, "Rapiță"), insertCrop(t, db, "Grâu")

	closedCrop, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: oldSeason, CropID: rapeseed, Status: domain.FieldCropStatusClosed, ClosedAt: ptr("2025-07-02")})
	if err != nil {
		t.Fatal(err)
	}
	openCrop, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: season, CropID: wheat})
	if err != nil {
		t.Fatal(err)
	}

	// data planificată nu contează: o lucrare din 2024 merge tot la cultura deschisă
	auto := plannedOp(field, opType, "2024-10-10")
	if err := ops.Create(auto); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, auto.ID); got == nil || *got != openCrop.ID {
		t.Errorf("lucrarea fără cultură explicită trebuie să meargă la cea deschisă: %v", got)
	}
	// cultura explicită, chiar închisă, e respectată (lucrări introduse târziu)
	explicit := plannedOp(field, opType, "2025-06-01")
	explicit.FieldCropID = &closedCrop.ID
	if err := ops.Create(explicit); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, explicit.ID); got == nil || *got != closedCrop.ID {
		t.Errorf("cultura explicită: %v", got)
	}
	// explicit fără cultură
	none := plannedOp(field, opType, "2025-06-01")
	none.FieldCropNone = true
	if err := ops.Create(none); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, none.ID); got != nil {
		t.Errorf("fără cultură explicit: %v", got)
	}
	// teren fără cultură deschisă → null, fără eroare
	orphan := plannedOp(other, opType, "2026-03-01")
	if err := ops.Create(orphan); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, orphan.ID); got != nil {
		t.Errorf("teren fără cultură deschisă: %v", got)
	}

	// Update fără cultură explicită păstrează legătura cât timp terenul nu se schimbă
	explicit.FieldCropID = nil
	explicit.Notes = "modificată"
	if err := ops.Update(explicit.ID, explicit); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, explicit.ID); got == nil || *got != closedCrop.ID {
		t.Errorf("update fără cultură explicită trebuie să păstreze legătura: %v", got)
	}
	// ...și o mută pe cultura deschisă a noului teren când terenul se schimbă
	moved := plannedOp(other, opType, "2026-03-01")
	moved.FieldCropID = nil
	if err := ops.Update(auto.ID, moved); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, auto.ID); got != nil {
		t.Errorf("pe terenul fără cultură deschisă legătura dispare: %v", got)
	}

	// închiderea culturii și modificarea sezonului nu ating lucrările deja salvate
	if _, err := svc.CloseFieldCrop(openCrop.ID, domain.FieldCropClose{Status: domain.FieldCropStatusClosed, ClosedAt: "2026-07-10", ProductionTotal: ptr(56.5)}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.UpdateSeason(season, &domain.Season{Name: "Ciclu toamnă 2025/2026", StartDate: "2026-06-01", EndDate: ptr("2026-06-02")}); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, explicit.ID); got == nil || *got != closedCrop.ID {
		t.Errorf("închiderea sau sezonul au mutat o lucrare: %v", got)
	}
	// după închidere, o lucrare nouă pe teren nu mai are unde să meargă
	after := plannedOp(field, opType, "2026-07-20")
	if err := ops.Create(after); err != nil {
		t.Fatal(err)
	}
	if got := operationCrop(t, db, after.ID); got != nil {
		t.Errorf("după închidere nu există cultură deschisă: %v", got)
	}
}

// Scenariul 8.2: paiele intră ca mișcare manuală legată de cultură, pe altă resursă decât cea de
// recoltă; raportul le arată ca produs secundar, iar istoricul le etichetează cu cultura.
func TestFieldCrop_SecondaryProductLinkedToCrop(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	fieldCropID, cropID := insertFieldCrop(t, db, "Grâu", "Tarla Mica", 56.5)
	if _, err := svc.RecordHarvest(fieldCropID, nil); err != nil {
		t.Fatal(err)
	}
	strawResource := insertStock(t, db, "Paie grâu baloți", 0, 0, 180)
	movements := usecase.NewStockMovementService(db, postgres.NewStockMovementRepo(db))
	if _, err := movements.Create(domain.StockMovementInput{ResourceID: strawResource, MovementType: domain.StockMovementIn, Quantity: 18, FieldCropID: &fieldCropID, Notes: "balotat 18.07"}); err != nil {
		t.Fatal(err)
	}
	missing := int64(999999)
	if _, err := movements.Create(domain.StockMovementInput{ResourceID: strawResource, MovementType: domain.StockMovementIn, Quantity: 1, FieldCropID: &missing}); !errors.Is(err, usecase.ErrInvalidStockMovement) {
		t.Errorf("cultura inexistentă trebuie să dea 400, nu %v", err)
	}

	var seasonID int64
	if err := db.QueryRow(`SELECT season_id FROM field_crops WHERE id = $1`, fieldCropID).Scan(&seasonID); err != nil {
		t.Fatal(err)
	}
	rows, err := postgres.NewReportRepo(db).GetFieldCropRows(seasonID, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || len(rows[0].SecondaryProducts) != 1 {
		t.Fatalf("raportul trebuie să arate exact un produs secundar: %+v", rows)
	}
	straw := rows[0].SecondaryProducts[0]
	if straw.ResourceID != strawResource || straw.Quantity != 18 || straw.TotalValue != 18*180 {
		t.Errorf("paiele în raport: %+v", straw)
	}
	if rows[0].HarvestRecordedQty == nil || *rows[0].HarvestRecordedQty != 56.5 || rows[0].RealCost != 0 {
		t.Errorf("recolta rămâne pe cultură și intrările nu umflă costul: %+v", rows[0])
	}
	// stocul a pornit de la 0, deci singura mișcare e balotatul (cea cu cultura inexistentă s-a anulat)
	list, err := movements.List(domain.StockMovementFilter{ResourceID: strawResource})
	if err != nil || len(list) != 1 {
		t.Fatalf("istoricul paielor: %v, %d", err, len(list))
	}
	if list[0].FieldOperationLabel == nil || *list[0].FieldOperationLabel != "Grâu - Tarla Mica" {
		t.Errorf("eticheta produsului secundar: %v", list[0].FieldOperationLabel)
	}
	_ = cropID
	assertMovementsExplainStocks(t, db)
}

// Scenariul 8.3: lucerna, patru coase cu producție cumulată; fiecare e „Recoltă”, nu „Corecție”.
func TestFieldCrop_RepeatedHarvestsAreLabeledAsHarvest(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	fieldCropID, cropID := insertFieldCrop(t, db, "Lucernă", "Lot lucernă", 12)
	for i, cumulated := range []float64{12, 26, 38, 47} {
		if i > 0 {
			setProduction(t, db, fieldCropID, cumulated)
		}
		result, err := svc.RecordHarvest(fieldCropID, nil)
		if err != nil || result.Movement == nil || result.Movement.MovementType != domain.StockMovementIn {
			t.Fatalf("coasa %d: %v, %+v", i+1, err, result)
		}
		if !strings.HasPrefix(result.Movement.Notes, "Recoltă ") {
			t.Errorf("coasa %d trebuie etichetată ca recoltă: %q", i+1, result.Movement.Notes)
		}
	}
	if quantity, movements := harvestState(t, db, cropID); quantity != 47 || movements != 4 {
		t.Errorf("după patru coase: stoc %.0f, %d mișcări", quantity, movements)
	}
	setProduction(t, db, fieldCropID, 45)
	result, err := svc.RecordHarvest(fieldCropID, nil)
	if err != nil || !strings.HasPrefix(result.Movement.Notes, "Corecție recoltă") {
		t.Errorf("scăderea rămâne corecție: %v, %+v", err, result.Movement)
	}
	assertMovementsExplainStocks(t, db)
}

// Sezonul în curs: fără dată de sfârșit, se salvează și se citește cu EndDate nil; o cultură se
// poate atribui pe el, iar raportul îl afișează.
func TestSeason_WithoutEndDate(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	season, err := svc.CreateSeason(&domain.Season{Name: "Ciclu toamnă 2026/2027", StartDate: "2026-08-01"})
	if err != nil || season.EndDate != nil {
		t.Fatalf("sezon în curs: %v, %+v", err, season)
	}
	field := insertField(t, db, "Tarla Nouă")
	if _, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: field, SeasonID: season.ID, CropID: insertCrop(t, db, "Rapiță")}); err != nil {
		t.Fatalf("cultură pe sezon în curs: %v", err)
	}
	rows, err := postgres.NewReportRepo(db).GetFieldCropRows(season.ID, "")
	if err != nil || len(rows) != 1 || rows[0].SeasonEnd != "" {
		t.Fatalf("raportul pe sezon în curs: %v, %+v", err, rows)
	}
	end := "2027-07-20"
	updated, err := svc.UpdateSeason(season.ID, &domain.Season{Name: season.Name, StartDate: season.StartDate, EndDate: &end})
	if err != nil || updated.EndDate == nil || *updated.EndDate != end {
		t.Errorf("încheierea sezonului: %v, %+v", err, updated)
	}
}
