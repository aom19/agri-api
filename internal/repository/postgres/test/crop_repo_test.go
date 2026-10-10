package postgres_test

import (
	"database/sql"
	"errors"
	"strings"
	"sync"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

func newCropService(db *sql.DB) *usecase.CropService {
	return usecase.NewCropService(postgres.NewCropRepo(db), postgres.NewFieldRepo(db), db, postgres.NewStockMovementRepo(db))
}

// insertFieldCrop creează sezonul (o dată), cultura (o dată, după nume) și cultura pe teren cu producția dată.
func insertFieldCrop(t *testing.T, db *sql.DB, cropName, fieldName string, production float64) (fieldCropID, cropID int64) {
	t.Helper()
	seasonID := insertID(t, db, `
		INSERT INTO seasons (name, start_date, end_date) VALUES ('2026', '2026-01-01', '2026-12-31')
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`)
	cropID = insertID(t, db, `
		INSERT INTO crops (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, cropName)
	fieldCropID = insertID(t, db, `
		INSERT INTO field_crops (field_id, season_id, crop_id, production_total)
		VALUES ($1, $2, $3, $4) RETURNING id`,
		insertField(t, db, fieldName), seasonID, cropID, production)
	return fieldCropID, cropID
}

func setProduction(t *testing.T, db *sql.DB, fieldCropID int64, production float64) {
	t.Helper()
	if _, err := db.Exec(`UPDATE field_crops SET production_total = $1 WHERE id = $2`, production, fieldCropID); err != nil {
		t.Fatal(err)
	}
}

// harvestState întoarce stocul resursei de recoltă a culturii și câte mișcări are.
func harvestState(t *testing.T, db *sql.DB, cropID int64) (quantity float64, movements int) {
	t.Helper()
	err := db.QueryRow(`
		SELECT r.quantity, (SELECT COUNT(*) FROM stock_movements sm WHERE sm.resource_id = r.id)
		FROM crops c JOIN resources r ON r.id = c.harvest_resource_id
		WHERE c.id = $1`, cropID).Scan(&quantity, &movements)
	if err != nil {
		t.Fatal(err)
	}
	return quantity, movements
}

func recordedQuantity(t *testing.T, db *sql.DB, fieldCropID int64) float64 {
	t.Helper()
	var recorded sql.NullFloat64
	if err := db.QueryRow(`SELECT harvest_recorded_quantity FROM field_crops WHERE id = $1`, fieldCropID).Scan(&recorded); err != nil {
		t.Fatal(err)
	}
	return recorded.Float64
}

func TestRecordHarvest_MovesOnlyTheDifference(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	fieldCropID, cropID := insertFieldCrop(t, db, "Grâu test", "Lot 1", 40)

	result, err := svc.RecordHarvest(fieldCropID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Movement == nil || result.Movement.MovementType != domain.StockMovementIn || result.Movement.QuantityDelta != 40 {
		t.Fatalf("prima înregistrare trebuie să fie o intrare de 40: %+v", result.Movement)
	}
	// resursa de recoltă e creată automat, cu tipul „harvest” în unitatea culturii
	var resourceName, category, unit string
	err = db.QueryRow(`
		SELECT r.name, rt.category, rt.default_unit
		FROM crops c JOIN resources r ON r.id = c.harvest_resource_id JOIN resource_types rt ON rt.id = r.resource_type_id
		WHERE c.id = $1`, cropID).Scan(&resourceName, &category, &unit)
	if err != nil {
		t.Fatal(err)
	}
	if resourceName != "Recoltă Grâu test" || category != "harvest" || unit != "t" {
		t.Errorf("resursa de recoltă: %q, %q, %q", resourceName, category, unit)
	}
	if quantity, _ := harvestState(t, db, cropID); quantity != 40 {
		t.Errorf("stocul de recoltă trebuie să fie 40, am %.3f", quantity)
	}

	// corecția în jos mută doar diferența, ca ieșire, pe aceeași resursă
	setProduction(t, db, fieldCropID, 35)
	result, err = svc.RecordHarvest(fieldCropID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Movement.MovementType != domain.StockMovementOut || result.Movement.QuantityDelta != -5 ||
		!strings.HasPrefix(result.Movement.Notes, "Corecție recoltă") {
		t.Errorf("corecția trebuie să fie o ieșire de 5: %+v", result.Movement)
	}
	quantity, movements := harvestState(t, db, cropID)
	if quantity != 35 || movements != 2 || recordedQuantity(t, db, fieldCropID) != 35 {
		t.Errorf("după corecție: stoc %.3f, %d mișcări, înregistrat %.3f", quantity, movements, recordedQuantity(t, db, fieldCropID))
	}

	// fără schimbare de producție nu se mai mută nimic
	result, err = svc.RecordHarvest(fieldCropID, nil)
	if err != nil || result.Movement != nil {
		t.Fatalf("a doua înregistrare identică trebuie să nu facă nimic: %v, %+v", err, result)
	}
	if _, movements := harvestState(t, db, cropID); movements != 2 {
		t.Errorf("înregistrarea repetată a adăugat mișcări: %d", movements)
	}
	var resources int
	if err := db.QueryRow(`SELECT COUNT(*) FROM resources`).Scan(&resources); err != nil || resources != 1 {
		t.Errorf("trebuie să existe o singură resursă de recoltă, am %d (%v)", resources, err)
	}
	assertMovementsExplainStocks(t, db)
}

func TestRecordHarvest_SameCropOnTwoFieldsSharesTheStock(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	first, cropID := insertFieldCrop(t, db, "Porumb test", "Lot 1", 40)
	second, _ := insertFieldCrop(t, db, "Porumb test", "Lot 2", 25)

	for _, id := range []int64{first, second} {
		if _, err := svc.RecordHarvest(id, nil); err != nil {
			t.Fatal(err)
		}
	}
	if quantity, movements := harvestState(t, db, cropID); quantity != 65 || movements != 2 {
		t.Errorf("recolta de pe două loturi trebuie adunată în același stoc: %.3f în %d mișcări", quantity, movements)
	}
	assertMovementsExplainStocks(t, db)
}

func TestRecordHarvest_FailedCorrectionChangesNothing(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	fieldCropID, cropID := insertFieldCrop(t, db, "Orz test", "Lot 1", 40)
	if _, err := svc.RecordHarvest(fieldCropID, nil); err != nil {
		t.Fatal(err)
	}

	// se vând 30 t, apoi producția e corectată la 20: ar trebui scoase 20 t, dar în stoc mai sunt 10
	var resourceID int64
	if err := db.QueryRow(`SELECT harvest_resource_id FROM crops WHERE id = $1`, cropID).Scan(&resourceID); err != nil {
		t.Fatal(err)
	}
	if _, err := newStockMovementService(db).Create(domain.StockMovementInput{ResourceID: resourceID, MovementType: domain.StockMovementOut, Quantity: 30}); err != nil {
		t.Fatal(err)
	}
	setProduction(t, db, fieldCropID, 20)

	if _, err := svc.RecordHarvest(fieldCropID, nil); !errors.Is(err, usecase.ErrHarvestNotRecordable) {
		t.Fatalf("corecția peste stocul disponibil trebuie refuzată, am %v", err)
	}
	quantity, movements := harvestState(t, db, cropID)
	if quantity != 10 || movements != 2 || recordedQuantity(t, db, fieldCropID) != 40 {
		t.Errorf("corecția refuzată a lăsat urme: stoc %.3f, %d mișcări, înregistrat %.3f",
			quantity, movements, recordedQuantity(t, db, fieldCropID))
	}
	assertMovementsExplainStocks(t, db)
}

// Două cereri simultane (ex. dublu-click pe „Înregistrează recolta”) trebuie să mute recolta o singură dată.
func TestRecordHarvest_ConcurrentCallsRecordOnce(t *testing.T) {
	db := requireDB(t)
	svc := newCropService(db)
	fieldCropID, _ := insertFieldCrop(t, db, "Floarea-soarelui test", "Lot 1", 40)

	const workers = 5
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.RecordHarvest(fieldCropID, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	// tot ce a intrat în stoc ca recoltă, indiferent pe câte resurse s-a împărțit
	var resources, movements int
	var total float64
	err := db.QueryRow(`
		SELECT COUNT(DISTINCT r.id), COUNT(sm.id), COALESCE(SUM(sm.quantity_delta), 0)
		FROM resources r
		JOIN resource_types rt ON rt.id = r.resource_type_id AND rt.category = 'harvest'
		LEFT JOIN stock_movements sm ON sm.resource_id = r.id`).Scan(&resources, &movements, &total)
	if err != nil {
		t.Fatal(err)
	}
	if resources != 1 || movements != 1 || total != 40 {
		t.Errorf("o producție de 40 t a intrat în stoc ca %.0f t, în %d mișcări, pe %d resurse de recoltă", total, movements, resources)
	}
	if recordedQuantity(t, db, fieldCropID) != 40 {
		t.Errorf("cantitatea înregistrată: %.3f", recordedQuantity(t, db, fieldCropID))
	}
	assertMovementsExplainStocks(t, db)
}
