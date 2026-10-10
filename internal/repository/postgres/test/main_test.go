package postgres_test

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Testele de integrare rulează pe un Postgres real, într-o bază separată, recreată la fiecare
// rulare din migrații: `make test-db`. Fără TEST_DATABASE_URL, testele sunt sărite, deci
// `make test` rămâne fără dependențe externe.

const migrationsDir = "../../../../migrations"

// tabelele golite înaintea fiecărui test; CASCADE golește și tabelele care le referă
var dataTables = []string{
	"stock_movements", "resources", "resource_types",
	"field_operations", "field_crops", "seasons", "crops",
	"machines", "implements", "users", "fields", "operation_templates",
}

var testDB *sql.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		os.Exit(m.Run())
	}

	db, err := setupDatabase(dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "baza de test nu a putut fi pregătită:", err)
		os.Exit(1)
	}
	testDB = db
	code := m.Run()
	_ = db.Close()
	os.Exit(code)
}

// setupDatabase recreează baza de test și aplică toate migrațiile .up.sql, în ordine.
func setupDatabase(dsn string) (*sql.DB, error) {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	name := strings.TrimPrefix(parsed.Path, "/")
	// protecție: baza e ștearsă la fiecare rulare, deci nu acceptăm baza de dezvoltare
	if !strings.HasSuffix(name, "_test") {
		return nil, fmt.Errorf("numele bazei trebuie să se termine în _test, am primit %q", name)
	}

	admin := *parsed
	admin.Path = "/postgres"
	adminDB, err := sql.Open("postgres", admin.String())
	if err != nil {
		return nil, err
	}
	defer adminDB.Close()
	ident := `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	if _, err := adminDB.Exec("DROP DATABASE IF EXISTS " + ident + " WITH (FORCE)"); err != nil {
		return nil, err
	}
	if _, err := adminDB.Exec("CREATE DATABASE " + ident); err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return nil, fmt.Errorf("%s: %w", filepath.Base(file), err)
		}
	}
	return db, nil
}

// requireDB întoarce baza de test golită de date, sau sare testul dacă nu există.
func requireDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDB == nil {
		t.Skip("TEST_DATABASE_URL nu e setat; rulează cu `make test-db`")
	}
	if _, err := testDB.Exec("TRUNCATE " + strings.Join(dataTables, ", ") + " RESTART IDENTITY CASCADE"); err != nil {
		t.Fatal(err)
	}
	return testDB
}

// ─── Date de test, inserate direct în SQL ca să nu depindă de codul testat ───

func insertID(t *testing.T, db *sql.DB, query string, args ...interface{}) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(query, args...).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

func insertField(t *testing.T, db *sql.DB, name string) string {
	t.Helper()
	var id string
	err := db.QueryRow(`
		INSERT INTO fields (name, area_ha, geometry)
		VALUES ($1, 10, '{"type":"Polygon","coordinates":[[[28.8,47.0],[28.9,47.0],[28.9,47.1],[28.8,47.0]]]}')
		RETURNING id`, name).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// insertStock creează tipul de resursă și resursa, cu stocul ei; întoarce id-ul resursei.
func insertStock(t *testing.T, db *sql.DB, name string, quantity, minimum, price float64) int64 {
	t.Helper()
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ($1, 'fuel', 'l') RETURNING id`, "Tip "+name)
	resourceID := insertID(t, db, `
		INSERT INTO resources (name, resource_type_id, price_per_unit, quantity, minimum_quantity)
		VALUES ($1, $2, $3, $4, $5) RETURNING id`, name, typeID, price, quantity, minimum)
	// cantitatea inițială intră ca mișcare, ca istoricul să explice stocul (vezi assertMovementsExplainStocks)
	if quantity != 0 {
		if _, err := db.Exec(`
			INSERT INTO stock_movements (resource_id, movement_type, quantity_delta, resulting_quantity, notes)
			VALUES ($1, 'adjustment', $2, $2, 'Stoc inițial')`, resourceID, quantity); err != nil {
			t.Fatal(err)
		}
	}
	return resourceID
}

// stockQuantity este stocul curent al resursei.
func stockQuantity(t *testing.T, db *sql.DB, resourceID int64) float64 {
	t.Helper()
	var quantity float64
	if err := db.QueryRow(`SELECT quantity FROM resources WHERE id = $1`, resourceID).Scan(&quantity); err != nil {
		t.Fatal(err)
	}
	return quantity
}

// assertMovementsExplainStocks verifică invariantul stocului: pentru fiecare resursă, suma
// variațiilor din mișcări este egală cu cantitatea curentă.
func assertMovementsExplainStocks(t *testing.T, db *sql.DB) {
	t.Helper()
	rows, err := db.Query(`
		SELECT r.id, r.quantity, COALESCE(SUM(sm.quantity_delta), 0)
		FROM resources r
		LEFT JOIN stock_movements sm ON sm.resource_id = r.id
		GROUP BY r.id, r.quantity
		HAVING r.quantity <> COALESCE(SUM(sm.quantity_delta), 0)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id            int64
			quantity, sum float64
		)
		if err := rows.Scan(&id, &quantity, &sum); err != nil {
			t.Fatal(err)
		}
		t.Errorf("resursa #%d are în stoc %.4f, dar mișcările însumează %.4f", id, quantity, sum)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

func insertMachine(t *testing.T, db *sql.DB, name, code string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO machines (name, code, type) VALUES ($1, $2, 'tractor') RETURNING id`, name, code)
}

// insertOperator creează un cont cu rolul operator și profilul lui; întoarce id-ul contului.
func insertOperator(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	id := insertID(t, db, `
		INSERT INTO users (email, password_hash, role_id)
		VALUES (LOWER($1) || '@test.ro', '!', (SELECT id FROM roles WHERE code = 'operator'))
		RETURNING id`, name)
	if _, err := db.Exec(`INSERT INTO user_profiles (user_id, first_name) VALUES ($1, $2)`, id, name); err != nil {
		t.Fatal(err)
	}
	return id
}

// fieldOperation descrie o operațiune pe teren; câmpurile nil rămân NULL în baza de date.
type fieldOperation struct {
	FieldID       string
	OperationType string // tipul propriu, doar fără template
	TemplateID    *int64
	MachineID     *int64
	OperatorID    *int64
	Status        string
	PlannedStart  time.Time
	PlannedEnd    *time.Time
	ActualEnd     *time.Time
	AreaPlannedHa float64
	FuelUsedL     *float64
	MachineHours  *float64
	Deleted       bool
}

// insertFieldOperation inserează operațiunea; FuelUsedL devine ieșire din stocul de motorină
// legată de ea, singura sursă din care se citește combustibilul.
func insertFieldOperation(t *testing.T, db *sql.DB, op fieldOperation) int64 {
	t.Helper()
	var deletedAt interface{}
	if op.Deleted {
		deletedAt = time.Now()
	}
	id := insertID(t, db, `
		INSERT INTO field_operations (
			field_id, operation_type, operation_template_id, machine_id, operator_id, status,
			planned_start_at, planned_end_at, actual_end_at, area_planned_ha,
			machine_hours, deleted_at
		) VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		op.FieldID, op.OperationType, op.TemplateID, op.MachineID, op.OperatorID, op.Status,
		op.PlannedStart, op.PlannedEnd, op.ActualEnd, op.AreaPlannedHa,
		op.MachineHours, deletedAt,
	)
	if op.FuelUsedL != nil {
		resourceID := fuelStock(t, db)
		if _, err := db.Exec(`
			WITH updated AS (
				UPDATE resources SET quantity = quantity - $2 WHERE id = $1 RETURNING quantity
			)
			INSERT INTO stock_movements (resource_id, field_operation_id, movement_type, quantity_delta, resulting_quantity)
			SELECT $1, $3, 'out', -$2::numeric, quantity FROM updated`,
			resourceID, *op.FuelUsedL, id); err != nil {
			t.Fatal(err)
		}
	}
	return id
}

// fuelStock întoarce resursa de motorină a testului, creând-o la prima folosire.
func fuelStock(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var resourceID int64
	err := db.QueryRow(`SELECT id FROM resources WHERE name = 'Motorină (test)'`).Scan(&resourceID)
	if err == sql.ErrNoRows {
		return insertStock(t, db, "Motorină (test)", 100000, 0, 7)
	}
	if err != nil {
		t.Fatal(err)
	}
	return resourceID
}

func ptr[T any](v T) *T { return &v }
