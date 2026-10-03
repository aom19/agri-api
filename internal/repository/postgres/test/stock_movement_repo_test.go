package postgres_test

import (
	"database/sql"
	"errors"
	"sort"
	"sync"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

func newStockMovementService(db *sql.DB) *usecase.StockMovementService {
	return usecase.NewStockMovementService(db, postgres.NewStockMovementRepo(db))
}

func stockQuantity(t *testing.T, db *sql.DB, stockID int64) float64 {
	t.Helper()
	var quantity float64
	if err := db.QueryRow(`SELECT quantity FROM stocks WHERE id = $1`, stockID).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	return quantity
}

func TestStockMovements_UpdateStockAndHistory(t *testing.T) {
	db := requireDB(t)
	svc := newStockMovementService(db)
	stockID, resourceID := insertStock(t, db, "Motorină flotă", 100, 20, 7.5)

	steps := []struct {
		movementType domain.StockMovementType
		quantity     float64
		want         float64
	}{
		{domain.StockMovementIn, 50, 150},
		{domain.StockMovementOut, 30, 120},
		{domain.StockMovementAdjustment, 118, 118}, // inventar: nivelul real, nu variația
	}
	for _, step := range steps {
		if _, err := svc.Create(domain.StockMovementInput{StockID: stockID, MovementType: step.movementType, Quantity: step.quantity}); err != nil {
			t.Fatalf("%s %.0f: %v", step.movementType, step.quantity, err)
		}
		if got := stockQuantity(t, db, stockID); got != step.want {
			t.Fatalf("după %s %.0f stocul e %.3f, mă așteptam la %.0f", step.movementType, step.quantity, got, step.want)
		}
	}

	// ieșirea peste stoc e refuzată și nu lasă urme
	if _, err := svc.Create(domain.StockMovementInput{StockID: stockID, MovementType: domain.StockMovementOut, Quantity: 1000}); !errors.Is(err, usecase.ErrInvalidStockMovement) {
		t.Fatalf("ieșirea peste stoc trebuie refuzată, am %v", err)
	}
	if got := stockQuantity(t, db, stockID); got != 118 {
		t.Errorf("ieșirea refuzată a modificat stocul: %.3f", got)
	}

	movements, err := postgres.NewStockMovementRepo(db).List(domain.StockMovementFilter{StockID: stockID})
	if err != nil {
		t.Fatal(err)
	}
	if len(movements) != 4 {
		t.Fatalf("mă așteptam la 4 mișcări (stocul inițial + 3), am %d", len(movements))
	}
	// lista e de la cea mai nouă la cea mai veche și aduce datele resursei prin JOIN
	latest := movements[0]
	if latest.MovementType != domain.StockMovementAdjustment || latest.QuantityDelta != -2 || latest.ResultingQuantity != 118 {
		t.Errorf("ultima mișcare: %+v", latest)
	}
	if latest.ResourceID != resourceID || latest.ResourceName != "Motorină flotă" || latest.Category != "fuel" || latest.Unit != "l" {
		t.Errorf("datele resursei lipsesc din listă: %+v", latest)
	}
	if out := movements[1]; out.QuantityDelta != -30 || out.TotalCost == nil || *out.TotalCost != 225 {
		t.Errorf("ieșirea trebuie evaluată la prețul resursei (30 × 7,5): %+v", out)
	}

	// istoricul explică stocul curent
	assertMovementsExplainStocks(t, db)
}

// Fără FOR UPDATE, două ieșiri simultane citesc aceeași cantitate și una dintre ele se pierde.
func TestStockMovements_ConcurrentOutputsAreSerialized(t *testing.T) {
	db := requireDB(t)
	svc := newStockMovementService(db)
	stockID, _ := insertStock(t, db, "Azotat", 100, 0, 1)

	const workers = 20
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Create(domain.StockMovementInput{StockID: stockID, MovementType: domain.StockMovementOut, Quantity: 1})
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	if got := stockQuantity(t, db, stockID); got != 100-workers {
		t.Fatalf("după %d ieșiri simultane de 1, stocul e %.3f, mă așteptam la %d", workers, got, 100-workers)
	}
	assertMovementsExplainStocks(t, db)
	// fiecare mișcare a văzut stocul lăsat de cea dinainte: 99, 98, ..., 80
	rows, err := db.Query(`SELECT resulting_quantity FROM stock_movements WHERE stock_id = $1 AND movement_type = 'out'`, stockID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var resulting []float64
	for rows.Next() {
		var value float64
		if err := rows.Scan(&value); err != nil {
			t.Fatal(err)
		}
		resulting = append(resulting, value)
	}
	sort.Float64s(resulting)
	for i, value := range resulting {
		if value != float64(100-workers+i) {
			t.Fatalf("cantitățile rezultate nu sunt consecutive: %v", resulting)
		}
	}
}

func TestStockMovements_ConcurrentOutputsNeverGoNegative(t *testing.T) {
	db := requireDB(t)
	svc := newStockMovementService(db)
	stockID, _ := insertStock(t, db, "Erbicid", 5, 0, 1)

	const workers = 12
	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		succeeded int
	)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := svc.Create(domain.StockMovementInput{StockID: stockID, MovementType: domain.StockMovementOut, Quantity: 1})
			if err != nil && !errors.Is(err, usecase.ErrInvalidStockMovement) {
				t.Error(err)
				return
			}
			if err == nil {
				mu.Lock()
				succeeded++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	if succeeded != 5 {
		t.Errorf("din %d ieșiri de 1 pe un stoc de 5, trebuiau să reușească exact 5, au reușit %d", workers, succeeded)
	}
	if got := stockQuantity(t, db, stockID); got != 0 {
		t.Errorf("stocul final trebuie să fie 0, am %.3f", got)
	}
	assertMovementsExplainStocks(t, db)
}
