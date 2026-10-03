package postgres_test

import (
	"os"
	"path/filepath"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

// Cantitatea inițială a unui stoc nou intră în istoric, iar editarea schimbă doar minimul.
func TestStocks_QuantityChangesOnlyThroughMovements(t *testing.T) {
	db := requireDB(t)
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ('Fertilizant', 'fertilizer', 'kg') RETURNING id`)
	resourceID := insertID(t, db, `INSERT INTO resources (name, resource_type_id, price_per_unit) VALUES ('NPK', $1, 2.5) RETURNING id`, typeID)
	svc := usecase.NewStockService(db, postgres.NewStockRepo(db), postgres.NewResourceRepo(db), postgres.NewStockMovementRepo(db))

	stock, err := svc.CreateStock(&domain.Stock{ResourceID: resourceID, Quantity: 40.1234, MinimumQuantity: 5}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if stock.Quantity != 40.1234 || stock.MinimumQuantity != 5 {
		t.Fatalf("stocul creat: %+v", stock)
	}
	if _, err := newStockMovementService(db).Create(domain.StockMovementInput{StockID: stock.ID, MovementType: domain.StockMovementOut, Quantity: 0.33335}); err != nil {
		t.Fatal(err)
	}
	updated, err := svc.UpdateMinimum(stock.ID, 8)
	if err != nil {
		t.Fatal(err)
	}
	if updated.MinimumQuantity != 8 || updated.Quantity != stockQuantity(t, db, stock.ID) {
		t.Errorf("UpdateMinimum: %+v", updated)
	}
	assertMovementsExplainStocks(t, db)
}

// Migrația 000046 adaugă o ajustare pentru stocurile editate direct, pe care istoricul nu le explica.
func TestStocks_MigrationReconcilesHistory(t *testing.T) {
	db := requireDB(t)
	// editat în sus după o ieșire: soldul de deschidere e pozitiv
	raisedID, _ := insertStock(t, db, "Motorină", 100, 0, 7)
	if _, err := newStockMovementService(db).Create(domain.StockMovementInput{StockID: raisedID, MovementType: domain.StockMovementOut, Quantity: 30}); err != nil {
		t.Fatal(err)
	}
	// editat în jos: diferența negativă intră ca ajustare la zi
	loweredID, _ := insertStock(t, db, "Uree", 50, 0, 2)
	untouchedID, _ := insertStock(t, db, "Semințe", 10, 0, 1)
	if _, err := db.Exec(`UPDATE stocks SET quantity = 500 WHERE id = $1`, raisedID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`UPDATE stocks SET quantity = 20 WHERE id = $1`, loweredID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`DELETE FROM stock_movements WHERE stock_id = $1`, loweredID); err != nil {
		t.Fatal(err)
	}

	migration, err := os.ReadFile(filepath.Join(migrationsDir, "000046_stock_movements_explain_stock.up.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(string(migration)); err != nil {
		t.Fatal(err)
	}
	assertMovementsExplainStocks(t, db)

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stock_movements WHERE stock_id = $1`, untouchedID).Scan(&count); err != nil || count != 1 {
		t.Errorf("stocul deja explicat nu trebuie atins: %d mișcări, %v", count, err)
	}
	// ajustarea pentru stocul crescut e prima în istoric, înaintea ieșirii
	var notes string
	if err := db.QueryRow(`SELECT notes FROM stock_movements WHERE stock_id = $1 ORDER BY created_at, id LIMIT 1`, raisedID).Scan(&notes); err != nil || notes != "Sold inițial (reconciliere istoric)" {
		t.Errorf("prima mișcare a stocului crescut: %q, %v", notes, err)
	}
}
