package postgres_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"

	"github.com/lib/pq"
)

func newResourceService(db *sql.DB) *usecase.ResourceService {
	return usecase.NewResourceService(db, postgres.NewResourceTypeRepo(db), postgres.NewResourceRepo(db), postgres.NewStockMovementRepo(db))
}

// T12: resursa se creează cu stocul ei într-un singur pas. Cantitatea inițială intră în istoric,
// iar editarea schimbă doar minimul.
func TestStocks_QuantityChangesOnlyThroughMovements(t *testing.T) {
	db := requireDB(t)
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ('Fertilizant', 'fertilizer', 'kg') RETURNING id`)
	svc := newResourceService(db)

	resource, err := svc.CreateResource(&domain.Resource{Name: "NPK", ResourceTypeID: typeID, PricePerUnit: 2.5, Quantity: 40.1234, MinimumQuantity: 5}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if resource.Quantity != 40.1234 || resource.MinimumQuantity != 5 {
		t.Fatalf("resursa creată: %+v", resource)
	}
	empty, err := svc.CreateResource(&domain.Resource{Name: "Uree", ResourceTypeID: typeID, PricePerUnit: 2}, nil)
	if err != nil || empty.Quantity != 0 {
		t.Fatalf("resursă fără stoc inițial: %v, %+v", err, empty)
	}
	if _, err := newStockMovementService(db).Create(domain.StockMovementInput{ResourceID: resource.ID, MovementType: domain.StockMovementOut, Quantity: 0.33335}); err != nil {
		t.Fatal(err)
	}
	updated, err := svc.UpdateResource(resource.ID, &domain.Resource{Name: "NPK 16-16-16", ResourceTypeID: typeID, PricePerUnit: 3, Quantity: 999, MinimumQuantity: 8})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "NPK 16-16-16" || updated.MinimumQuantity != 8 || updated.Quantity != stockQuantity(t, db, resource.ID) || updated.Quantity == 999 {
		t.Errorf("UpdateResource trebuie să lase cantitatea neschimbată: %+v", updated)
	}

	var movements int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE resource_id = $1`, empty.ID).Scan(&movements); err != nil || movements != 0 {
		t.Errorf("resursa fără stoc inițial nu are mișcări: %d, %v", movements, err)
	}
	assertMovementsExplainStocks(t, db)
}

// O resursă cu istoric de stoc nu poate fi ștearsă: ștergerea ar lăsa mișcările fără resursă.
func TestStocks_ResourceWithHistoryIsNotDeleted(t *testing.T) {
	db := requireDB(t)
	svc := newResourceService(db)
	withHistory := insertStock(t, db, "Motorină", 100, 0, 7)
	unused := insertStock(t, db, "Benzină", 0, 0, 8)

	if err := svc.DeleteResource(withHistory); err == nil || !strings.Contains(err.Error(), "mișcări de stoc") {
		t.Errorf("ștergerea resursei cu mișcări trebuie refuzată, am %v", err)
	}
	if got := stockQuantity(t, db, withHistory); got != 100 {
		t.Errorf("resursa refuzată la ștergere și-a pierdut stocul: %.3f", got)
	}
	if err := svc.DeleteResource(unused); err != nil {
		t.Errorf("resursa fără mișcări se șterge: %v", err)
	}
}

// ─── Migrații pe schema de dinainte de T12 ───────────────────────────────────

const mergeStocksMigration = "000052_merge_stocks_into_resources"

func runMigrationFile(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(migrationsDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(content)); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

// schemaBeforeT12 dă înapoi migrația 052, deci stocul stă din nou în tabelul stocks. Întoarce
// funcția care o reaplică; dacă testul se oprește înainte, migrația se reaplică la final.
func schemaBeforeT12(t *testing.T, db *sql.DB) (migrateUp func()) {
	t.Helper()
	runMigrationFile(t, db, mergeStocksMigration+".down.sql")
	applied := false
	migrateUp = func() {
		t.Helper()
		applied = true
		runMigrationFile(t, db, mergeStocksMigration+".up.sql")
	}
	t.Cleanup(func() {
		if !applied {
			migrateUp()
		}
	})
	return migrateUp
}

// insertStockBeforeT12 creează, pe schema de dinainte de T12, resursa și stocul ei separat, cu
// cantitatea inițială ca mișcare; întoarce id-ul stocului și al resursei.
func insertStockBeforeT12(t *testing.T, db *sql.DB, name string, quantity, minimum, price float64) (stockID, resourceID int64) {
	t.Helper()
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ($1, 'fuel', 'l') RETURNING id`, "Tip "+name)
	resourceID = insertID(t, db, `INSERT INTO resources (name, resource_type_id, price_per_unit) VALUES ($1, $2, $3) RETURNING id`, name, typeID, price)
	stockID = insertID(t, db, `INSERT INTO stocks (resource_id, quantity, minimum_quantity) VALUES ($1, $2, $3) RETURNING id`, resourceID, quantity, minimum)
	if quantity != 0 {
		if _, err := db.Exec(`
			INSERT INTO stock_movements (stock_id, resource_id, movement_type, quantity_delta, resulting_quantity, notes)
			VALUES ($1, $2, 'adjustment', $3, $3, 'Stoc inițial')`, stockID, resourceID, quantity); err != nil {
			t.Fatal(err)
		}
	}
	return stockID, resourceID
}

// Migrația 052 mută cantitatea și minimul pe resursă, păstrează istoricul și trimite jurnalul de
// audit și permisiunile de stoc la resursă.
func TestStocks_MigrationMovesStockToResource(t *testing.T) {
	db := requireDB(t)
	migrateUp := schemaBeforeT12(t, db)

	// id-uri diferite pentru stoc și resursă, ca rescrierea jurnalului de audit să conteze
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ('Fără stoc', 'other', 'buc') RETURNING id`)
	withoutStock := insertID(t, db, `INSERT INTO resources (name, resource_type_id) VALUES ('Lavete', $1) RETURNING id`, typeID)
	stockID, resourceID := insertStockBeforeT12(t, db, "Motorină", 120, 30, 7)
	if stockID == resourceID {
		t.Fatal("stocul și resursa trebuie să aibă id-uri diferite")
	}

	itoa := func(id int64) string { return strconv.FormatInt(id, 10) }
	var auditIDs []int64
	for _, entry := range []struct{ entityID, action, changes string }{
		{itoa(stockID), "movement", `{}`},
		// stoc șters înainte de migrare: resursa vine din intrarea de creare
		{"999", "create", `{"resource_id": ` + itoa(withoutStock) + `}`},
		{"999", "delete", `{}`},
	} {
		auditIDs = append(auditIDs, insertID(t, db, `
			INSERT INTO audit_log (entity_type, entity_id, action, changes)
			VALUES ('stock', $1, $2, $3::jsonb) RETURNING id`, entry.entityID, entry.action, entry.changes))
	}
	roleID := insertID(t, db, `INSERT INTO roles (code, name) VALUES ('magazioner_test', 'Magazioner (test)') RETURNING id`)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM roles WHERE id = $1`, roleID)
		_, _ = db.Exec(`DELETE FROM audit_log WHERE id = ANY($1)`, pq.Array(auditIDs))
	})
	if _, err := db.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		SELECT $1, id FROM permissions WHERE name IN ('stock.view', 'stock.update')`, roleID); err != nil {
		t.Fatal(err)
	}

	migrateUp()

	var stocksTable sql.NullString
	if err := db.QueryRow(`SELECT to_regclass('public.stocks')::text`).Scan(&stocksTable); err != nil || stocksTable.Valid {
		t.Errorf("tabelul stocks trebuie șters: %v, %v", stocksTable, err)
	}
	resources := postgres.NewResourceRepo(db)
	if got, err := resources.GetByID(resourceID); err != nil || got.Quantity != 120 || got.MinimumQuantity != 30 {
		t.Errorf("stocul mutat pe resursă: %v, %+v", err, got)
	}
	if got, err := resources.GetByID(withoutStock); err != nil || got.Quantity != 0 || got.MinimumQuantity != 0 {
		t.Errorf("resursa fără stoc pornește de la 0: %v, %+v", err, got)
	}
	movements, err := postgres.NewStockMovementRepo(db).List(domain.StockMovementFilter{ResourceID: resourceID})
	if err != nil || len(movements) != 1 || movements[0].Notes != "Stoc inițial" {
		t.Errorf("istoricul resursei: %v, %+v", err, movements)
	}
	assertMovementsExplainStocks(t, db)

	wantEntity := []string{itoa(resourceID), itoa(withoutStock), itoa(withoutStock)}
	for i, id := range auditIDs {
		var entityID string
		if err := db.QueryRow(`SELECT entity_id FROM audit_log WHERE id = $1`, id).Scan(&entityID); err != nil || entityID != wantEntity[i] {
			t.Errorf("intrarea de audit #%d trimite la %q, mă așteptam la %q (%v)", i, entityID, wantEntity[i], err)
		}
	}

	var granted []string
	rows, err := db.Query(`
		SELECT p.name FROM role_permissions rp JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1 ORDER BY p.name`, roleID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		granted = append(granted, name)
	}
	if strings.Join(granted, ",") != "resources:read,resources:write" {
		t.Errorf("stock.view și stock.update devin resources:read și resources:write, am %v", granted)
	}
	var leftover int
	if err := db.QueryRow(`SELECT COUNT(*) FROM permissions WHERE name LIKE 'stock.%'`).Scan(&leftover); err != nil || leftover != 0 {
		t.Errorf("permisiunile stock.* trebuie șterse: %d, %v", leftover, err)
	}
}

// Migrația 000046 adaugă o ajustare pentru stocurile editate direct, pe care istoricul nu le explica.
// Rulează pe schema de atunci; după migrația 052, istoricul explică stocul fiecărei resurse.
func TestStocks_MigrationReconcilesHistory(t *testing.T) {
	db := requireDB(t)
	migrateUp := schemaBeforeT12(t, db)

	// editat în sus după o ieșire: soldul de deschidere e pozitiv
	raisedID, raisedResource := insertStockBeforeT12(t, db, "Motorină", 100, 0, 7)
	if _, err := db.Exec(`
		INSERT INTO stock_movements (stock_id, resource_id, movement_type, quantity_delta, resulting_quantity)
		VALUES ($1, $2, 'out', -30, 70)`, raisedID, raisedResource); err != nil {
		t.Fatal(err)
	}
	// editat în jos: diferența negativă intră ca ajustare la zi
	loweredID, _ := insertStockBeforeT12(t, db, "Uree", 50, 0, 2)
	_, untouchedResource := insertStockBeforeT12(t, db, "Semințe", 10, 0, 1)
	if _, err := db.Exec(`UPDATE stocks SET quantity = 500 WHERE id = $1`, raisedID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE stocks SET quantity = 20 WHERE id = $1`, loweredID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM stock_movements WHERE stock_id = $1`, loweredID); err != nil {
		t.Fatal(err)
	}

	runMigrationFile(t, db, "000046_stock_movements_explain_stock.up.sql")
	migrateUp()
	assertMovementsExplainStocks(t, db)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE resource_id = $1`, untouchedResource).Scan(&count); err != nil || count != 1 {
		t.Errorf("stocul deja explicat nu trebuie atins: %d mișcări, %v", count, err)
	}
	// ajustarea pentru stocul crescut e prima în istoric, înaintea ieșirii
	var notes string
	if err := db.QueryRow(`SELECT notes FROM stock_movements WHERE resource_id = $1 ORDER BY created_at, id LIMIT 1`, raisedResource).Scan(&notes); err != nil || notes != "Sold inițial (reconciliere istoric)" {
		t.Errorf("prima mișcare a stocului crescut: %q, %v", notes, err)
	}
}
