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
	"stock_movements", "stocks", "resources", "resource_types",
	"field_operations", "field_crops", "seasons", "crops",
	"assignments", "machines", "operators", "fields", "operation_types",
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

// insertStock creează tipul de resursă, resursa și stocul ei; întoarce id-ul stocului și al resursei.
func insertStock(t *testing.T, db *sql.DB, name string, quantity, minimum, price float64) (stockID, resourceID int64) {
	t.Helper()
	typeID := insertID(t, db, `INSERT INTO resource_types (name, category, default_unit) VALUES ($1, 'fuel', 'l') RETURNING id`, "Tip "+name)
	resourceID = insertID(t, db, `INSERT INTO resources (name, resource_type_id, price_per_unit) VALUES ($1, $2, $3) RETURNING id`, name, typeID, price)
	stockID = insertID(t, db, `INSERT INTO stocks (resource_id, quantity, minimum_quantity) VALUES ($1, $2, $3) RETURNING id`, resourceID, quantity, minimum)
	return stockID, resourceID
}

func insertMachine(t *testing.T, db *sql.DB, name, code string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO machines (name, code, type) VALUES ($1, $2, 'tractor') RETURNING id`, name, code)
}

func insertOperator(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	return insertID(t, db, `INSERT INTO operators (name) VALUES ($1) RETURNING id`, name)
}

func insertAssignment(t *testing.T, db *sql.DB, machineID, operatorID int64, status string, deleted bool) {
	t.Helper()
	var deletedAt interface{}
	if deleted {
		deletedAt = time.Now()
	}
	if _, err := db.Exec(`
		INSERT INTO assignments (machine_id, operator_id, start_date, status, deleted_at)
		VALUES ($1, $2, NOW(), $3, $4)`, machineID, operatorID, status, deletedAt); err != nil {
		t.Fatal(err)
	}
}

// fieldOperation descrie o operațiune pe teren; câmpurile nil rămân NULL în baza de date.
type fieldOperation struct {
	FieldID         string
	OperationTypeID int64
	MachineID       *int64
	OperatorID      *int64
	Status          string
	PlannedStart    time.Time
	PlannedEnd      *time.Time
	ActualEnd       *time.Time
	AreaPlannedHa   float64
	FuelUsedL       *float64
	MachineHours    *float64
	Deleted         bool
}

func insertFieldOperation(t *testing.T, db *sql.DB, op fieldOperation) int64 {
	t.Helper()
	var deletedAt interface{}
	if op.Deleted {
		deletedAt = time.Now()
	}
	return insertID(t, db, `
		INSERT INTO field_operations (
			field_id, operation_type_id, machine_id, operator_id, status,
			planned_start_at, planned_end_at, actual_end_at, area_planned_ha,
			fuel_used_l, machine_hours, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id`,
		op.FieldID, op.OperationTypeID, op.MachineID, op.OperatorID, op.Status,
		op.PlannedStart, op.PlannedEnd, op.ActualEnd, op.AreaPlannedHa,
		op.FuelUsedL, op.MachineHours, deletedAt,
	)
}

func ptr[T any](v T) *T { return &v }
