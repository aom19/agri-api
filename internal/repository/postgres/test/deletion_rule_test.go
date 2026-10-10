package postgres_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/repository/postgres"
	"agri-api/internal/usecase"
)

// T22: regula de ștergere din README („Ștergere”).

func archive(t *testing.T, table string, id interface{}) {
	t.Helper()
	if _, err := testDB.Exec(`UPDATE `+table+` SET deleted_at = NOW() WHERE id = $1`, id); err != nil {
		t.Fatal(err)
	}
}

// Un element arhivat rămâne cu numele în lucrările vechi, dar nu mai poate fi ales într-una nouă
// și nu mai poate porni o lucrare.
func TestDeletion_ArchivedStaysInHistoryButCannotBeChosen(t *testing.T) {
	db := requireDB(t)
	service := usecase.NewFieldOperationService(postgres.NewFieldOperationRepo(db))

	fieldID := insertField(t, db, "Tarla Arhivă")
	machineID := insertMachine(t, db, "Tractor vechi", "TR-OLD")
	implementID := insertID(t, db, `INSERT INTO implements (name, code, type) VALUES ('Plug vechi', 'PL-OLD', 'plow') RETURNING id`)
	operatorID := insertOperator(t, db, "Ion")
	start := time.Now().Add(24 * time.Hour)
	input := func() *domain.FieldOperation {
		return &domain.FieldOperation{
			FieldID: fieldID, OperationType: domain.OperationTypeSoilPreparation,
			MachineID: ptr(machineID), ImplementID: ptr(implementID), OperatorID: ptr(operatorID),
			PlannedStartAt: &start,
		}
	}
	planned, err := service.Create(input())
	if err != nil {
		t.Fatal(err)
	}

	archive(t, "machines", machineID)
	archive(t, "implements", implementID)

	// istoricul păstrează numele; mașina arhivată apare inactivă, deci lucrarea nu pornește
	got, err := service.GetByID(planned.ID)
	if err != nil || got.MachineName == nil || *got.MachineName != "Tractor vechi" || got.MachineStatus == nil || *got.MachineStatus != "inactive" {
		t.Fatalf("lucrarea trebuie să păstreze mașina arhivată, ca inactivă: %v, %+v", err, got)
	}
	if _, err := service.Start(planned.ID); !errors.Is(err, usecase.ErrFieldOperationResourcesUnavailable) {
		t.Errorf("lucrarea nu poate porni cu o mașină arhivată: %v", err)
	}

	// editarea fără schimbarea mașinii merge; alegerea unei mașini arhivate, nu
	edited := input()
	edited.Notes = "mutată pe altă zi"
	if _, err := service.Update(planned.ID, edited); err != nil {
		t.Errorf("o lucrare veche rămâne editabilă: %v", err)
	}
	if _, err := service.Create(input()); !errors.Is(err, usecase.ErrFieldOperationArchivedReference) ||
		!strings.Contains(err.Error(), "mașina") || !strings.Contains(err.Error(), "echipamentul") {
		t.Errorf("o lucrare nouă nu poate folosi mașina și echipamentul arhivate: %v", err)
	}

	otherMachine := insertMachine(t, db, "Tractor nou", "TR-NEW")
	archive(t, "fields", fieldID)
	archive(t, "users", operatorID)
	fresh := input()
	fresh.MachineID, fresh.ImplementID = ptr(otherMachine), nil
	if _, err := service.Create(fresh); !errors.Is(err, usecase.ErrFieldOperationArchivedReference) ||
		!strings.Contains(err.Error(), "terenul") || !strings.Contains(err.Error(), "operatorul") {
		t.Errorf("terenul și operatorul arhivate nu pot fi alese: %v", err)
	}
	moved := input()
	moved.MachineID, moved.ImplementID = ptr(otherMachine), nil
	if _, err := service.Update(planned.ID, moved); err != nil {
		t.Errorf("schimbarea mașinii pe una activă merge, chiar dacă terenul s-a arhivat între timp: %v", err)
	}
}

// Codul echipamentului și numărul cadastral se pot refolosi după arhivare, ca la mașini.
func TestDeletion_CodesReusableAfterArchive(t *testing.T) {
	db := requireDB(t)

	implementID := insertID(t, db, `INSERT INTO implements (name, code, type) VALUES ('Plug', 'PL-1', 'plow') RETURNING id`)
	if _, err := db.Exec(`INSERT INTO implements (name, code, type) VALUES ('Plug 2', 'PL-1', 'plow')`); err == nil {
		t.Error("două echipamente active nu pot avea același cod")
	}
	archive(t, "implements", implementID)
	if _, err := db.Exec(`INSERT INTO implements (name, code, type) VALUES ('Plug 2', 'PL-1', 'plow')`); err != nil {
		t.Errorf("codul unui echipament arhivat se poate refolosi: %v", err)
	}

	fieldID := insertField(t, db, "Tarla cadastru")
	if _, err := db.Exec(`UPDATE fields SET cadastral_number = '2925.001' WHERE id = $1`, fieldID); err != nil {
		t.Fatal(err)
	}
	archive(t, "fields", fieldID)
	other := insertField(t, db, "Tarla nouă")
	if _, err := db.Exec(`UPDATE fields SET cadastral_number = '2925.001' WHERE id = $1`, other); err != nil {
		t.Errorf("numărul cadastral al unui teren arhivat se poate refolosi: %v", err)
	}
}

// Sezoanele și culturile pe terenuri se șterg definitiv, dar nu cu istoric: sezonul cu culturi și
// cultura cu recoltă în stoc sunt refuzate, în loc să fie șterse în cascadă.
func TestDeletion_HistoryIsNotDeletedInCascade(t *testing.T) {
	db := requireDB(t)
	crops := usecase.NewCropService(postgres.NewCropRepo(db), postgres.NewFieldRepo(db), db, postgres.NewStockMovementRepo(db))

	seasonID := insertID(t, db, `INSERT INTO seasons (name, start_date, end_date) VALUES ('Sezon test', '2026-01-01', '2026-12-31') RETURNING id`)
	emptySeason := insertID(t, db, `INSERT INTO seasons (name, start_date, end_date) VALUES ('Sezon gol', '2027-01-01', '2027-12-31') RETURNING id`)
	cropID := insertID(t, db, `INSERT INTO crops (name, yield_unit) VALUES ('Grâu test', 't') RETURNING id`)
	harvested := insertID(t, db, `
		INSERT INTO field_crops (field_id, season_id, crop_id, production_total)
		VALUES ($1, $2, $3, 40) RETURNING id`, insertField(t, db, "Tarla recoltată"), seasonID, cropID)
	plainField := insertField(t, db, "Tarla fără recoltă")
	plain := insertID(t, db, `INSERT INTO field_crops (field_id, season_id, crop_id) VALUES ($1, $2, $3) RETURNING id`, plainField, seasonID, cropID)
	operation := insertFieldOperation(t, db, fieldOperation{
		FieldID: plainField, OperationType: "seeding", Status: "planned", PlannedStart: time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC),
	})
	if _, err := db.Exec(`UPDATE field_operations SET field_crop_id = $1 WHERE id = $2`, plain, operation); err != nil {
		t.Fatal(err)
	}
	if _, err := crops.RecordHarvest(harvested, nil); err != nil {
		t.Fatal(err)
	}

	var inUse *repository.InUseError
	if err := crops.DeleteSeason(seasonID); !errors.As(err, &inUse) {
		t.Errorf("sezonul cu culturi pe terenuri nu se șterge: %v", err)
	}
	if err := crops.DeleteFieldCrop(harvested); !errors.As(err, &inUse) {
		t.Errorf("cultura cu recolta în stoc nu se șterge: %v", err)
	}
	if err := crops.DeleteCrop(cropID); !errors.As(err, &inUse) {
		t.Errorf("cultura folosită pe terenuri nu se șterge: %v", err)
	}

	// fără istoric se șterg; operațiunea rămâne, doar fără legătura cu cultura
	if err := crops.DeleteFieldCrop(plain); err != nil {
		t.Errorf("cultura fără recoltă se șterge: %v", err)
	}
	var linked *int64
	if err := db.QueryRow(`SELECT field_crop_id FROM field_operations WHERE id = $1`, operation).Scan(&linked); err != nil || linked != nil {
		t.Errorf("operațiunea rămâne, fără cultură: %v, %v", linked, err)
	}
	if err := crops.DeleteSeason(emptySeason); err != nil {
		t.Errorf("sezonul gol se șterge: %v", err)
	}
	assertMovementsExplainStocks(t, db)
}
