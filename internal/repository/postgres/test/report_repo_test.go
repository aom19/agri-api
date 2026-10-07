package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository/postgres"
)

// reportFixture: două mașini, trei operatori și operațiuni care acoperă fiecare caz numărat de rapoarte.
type reportFixture struct {
	filter                    domain.ReportFilter
	machine1, machine2        int64
	operator1, operator2, op3 int64
}

func seedReportData(t *testing.T, db *sql.DB) reportFixture {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Second)
	from, to := now.AddDate(0, 0, -30), now.AddDate(0, 0, 1)
	day := func(offset int) time.Time { return now.AddDate(0, 0, offset) }

	f := reportFixture{filter: domain.ReportFilter{From: from, To: to}}
	field := insertField(t, db, "Lot raport")
	opType := insertID(t, db, `INSERT INTO operation_types (code, name) VALUES ('plowing', 'Arat') RETURNING id`)
	f.machine1 = insertMachine(t, db, "Tractor A", "TR-A")
	f.machine2 = insertMachine(t, db, "Tractor B", "TR-B")
	f.operator1 = insertOperator(t, db, "Ana")
	f.operator2 = insertOperator(t, db, "Bogdan")
	f.op3 = insertOperator(t, db, "Cristi") // fără operațiuni: apare în raport cu zero

	base := fieldOperation{FieldID: field, OperationTypeID: opType}
	with := func(change func(*fieldOperation)) fieldOperation {
		op := base
		change(&op)
		return op
	}
	// mașina 1, operatorul 1: finalizată la timp
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine1, &f.operator1, "completed"
		op.PlannedStart, op.PlannedEnd, op.ActualEnd = day(-10), ptr(day(-8)), ptr(day(-9))
		op.AreaPlannedHa, op.FuelUsedL, op.MachineHours = 10, ptr(50.0), ptr(5.0)
	}))
	// mașina 1, operatorul 1: în lucru, cu sfârșitul planificat depășit (întârziată)
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine1, &f.operator1, "in_progress"
		op.PlannedStart, op.PlannedEnd, op.AreaPlannedHa = day(-5), ptr(day(-2)), 4
	}))
	// mașina 1, operatorul 2: planificată în viitor
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine1, &f.operator2, "planned"
		op.PlannedStart, op.PlannedEnd, op.AreaPlannedHa = day(-1), ptr(day(3)), 6
	}))
	// mașina 2, operatorul 2: finalizată cu întârziere
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine2, &f.operator2, "completed"
		op.PlannedStart, op.PlannedEnd, op.ActualEnd = day(-20), ptr(day(-19)), ptr(day(-15))
		op.AreaPlannedHa, op.FuelUsedL, op.MachineHours = 3, ptr(10.0), ptr(1.0)
	}))
	// excluse din raport: înainte de perioadă și ștearsă
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine1, &f.operator1, "completed"
		op.PlannedStart, op.AreaPlannedHa, op.FuelUsedL = day(-60), 100, ptr(999.0)
	}))
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine1, &f.operator1, "in_progress"
		op.PlannedStart, op.AreaPlannedHa, op.Deleted = day(-3), 100, true
	}))
	// planificată după perioadă: nu intră în cifrele perioadei, dar e o alocare activă
	insertFieldOperation(t, db, with(func(op *fieldOperation) {
		op.MachineID, op.OperatorID, op.Status = &f.machine2, &f.op3, "planned"
		op.PlannedStart, op.AreaPlannedHa = day(5), 7
	}))
	return f
}

func TestReportFleet_MachineRows(t *testing.T) {
	db := requireDB(t)
	f := seedReportData(t, db)
	repo := postgres.NewReportRepo(db)

	rows, err := repo.GetMachineRows(f.filter)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("mă așteptam la 2 mașini, am %d", len(rows))
	}
	// ordonate după numărul de operațiuni
	m1, m2 := rows[0], rows[1]
	if m1.ID != f.machine1 || m1.OperationsCount != 3 || m1.PlannedAreaHa != 20 || m1.FuelUsedL != 50 || m1.HoursInPeriod != 5 {
		t.Errorf("mașina 1: %+v", m1)
	}
	if m2.ID != f.machine2 || m2.OperationsCount != 1 || m2.PlannedAreaHa != 3 || m2.FuelUsedL != 10 || m2.HoursInPeriod != 1 {
		t.Errorf("mașina 2: %+v", m2)
	}
	// alocări active = operațiuni planned/in_progress neșterse, indiferent de perioadă (ca în dashboard)
	if m1.ActiveAssignments != 2 || m2.ActiveAssignments != 1 {
		t.Errorf("alocări active: mașina 1 = %d (aștept 2), mașina 2 = %d (aștept 1)", m1.ActiveAssignments, m2.ActiveAssignments)
	}

	filtered, err := repo.GetMachineRows(domain.ReportFilter{From: f.filter.From, To: f.filter.To, MachineID: f.machine2})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].ID != f.machine2 {
		t.Errorf("filtrul pe mașină: %+v", filtered)
	}
}

func TestReportOperators_OperatorRows(t *testing.T) {
	db := requireDB(t)
	f := seedReportData(t, db)

	rows, err := postgres.NewReportRepo(db).GetOperatorRows(f.filter)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[int64]domain.ReportOperatorRow{}
	for _, row := range rows {
		byID[row.ID] = row
	}
	if len(rows) != 3 {
		t.Fatalf("mă așteptam la 3 operatori (inclusiv cel fără operațiuni), am %d", len(rows))
	}

	o1 := byID[f.operator1]
	if o1.OperationsCount != 2 || o1.CompletedCount != 1 || o1.InProgressCount != 1 || o1.PlannedCount != 0 ||
		o1.OnTimeCount != 1 || o1.OverdueCount != 1 || o1.PlannedAreaHa != 14 {
		t.Errorf("operatorul 1: %+v", o1)
	}
	o2 := byID[f.operator2]
	if o2.OperationsCount != 2 || o2.CompletedCount != 1 || o2.PlannedCount != 1 ||
		o2.OnTimeCount != 0 || o2.OverdueCount != 0 || o2.PlannedAreaHa != 9 {
		t.Errorf("operatorul 2 (o finalizare întârziată, o planificare în termen): %+v", o2)
	}
	o3 := byID[f.op3]
	if o3.OperationsCount != 0 || o3.PlannedAreaHa != 0 {
		t.Errorf("operatorul fără operațiuni: %+v", o3)
	}
	// operatorul 3 nu are operațiuni în perioadă, dar are una planificată după ea
	if o1.ActiveAssignments != 1 || o2.ActiveAssignments != 1 || o3.ActiveAssignments != 1 {
		t.Errorf("alocări active: operator 1 = %d, operator 2 = %d, operator 3 = %d (aștept 1 la fiecare)",
			o1.ActiveAssignments, o2.ActiveAssignments, o3.ActiveAssignments)
	}
	// operatorii cu același număr de operațiuni sunt ordonați după nume
	if rows[0].ID != f.operator1 || rows[1].ID != f.operator2 || rows[2].ID != f.op3 {
		t.Errorf("ordinea operatorilor: %v, %v, %v", rows[0].Name, rows[1].Name, rows[2].Name)
	}
}
