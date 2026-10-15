package usecase_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
)

func TestMapCropRepoError(t *testing.T) {
	if usecase.MapCropRepoError(nil) != nil {
		t.Error("nil trebuie să rămână nil")
	}
	if !errors.Is(usecase.MapCropRepoError(sql.ErrNoRows), usecase.ErrCropNotFound) {
		t.Error("ErrNoRows → ErrCropNotFound")
	}
	if !errors.Is(usecase.MapCropRepoError(repository.ErrDuplicateEntry), usecase.ErrCropDuplicate) {
		t.Error("ErrDuplicateEntry → ErrCropDuplicate")
	}
	other := errors.New("x")
	if !errors.Is(usecase.MapCropRepoError(other), other) {
		t.Error("alte erori trec neschimbate")
	}
}

func TestCropService_Seasons(t *testing.T) {
	repo := &cropRepoMock{
		createSeason:  func(s *domain.Season) error { s.ID = 1; return nil },
		getSeasonByID: func(id int64) (*domain.Season, error) { return &domain.Season{ID: id}, nil },
	}
	svc := usecase.NewCropService(repo, &fieldRepoMock{}, nil, nil)

	invalid := map[string]*domain.Season{
		"nume lipsă":        {StartDate: "2026-01-01", EndDate: ptr("2026-06-01")},
		"început invalid":   {Name: "S", StartDate: "azi", EndDate: ptr("2026-06-01")},
		"sfârșit invalid":   {Name: "S", StartDate: "2026-01-01", EndDate: ptr("mâine")},
		"interval inversat": {Name: "S", StartDate: "2026-06-01", EndDate: ptr("2026-01-01")},
	}
	for name, in := range invalid {
		if _, err := svc.CreateSeason(in); !errors.Is(err, usecase.ErrCropInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}

	season := &domain.Season{Name: "  2026  ", StartDate: " 2026-01-01", EndDate: ptr("2026-12-31 ")}
	created, err := svc.CreateSeason(season)
	if err != nil || created.ID != 1 || season.Name != "2026" || season.StartDate != "2026-01-01" {
		t.Fatalf("CreateSeason: %v, %+v", err, season)
	}
	if season.EndDate == nil || *season.EndDate != "2026-12-31" {
		t.Errorf("sfârșitul trebuie normalizat: %v", season.EndDate)
	}
	// sezon în curs: fără dată de sfârșit (lipsă sau goală)
	ongoing := &domain.Season{Name: "În curs", StartDate: "2026-08-01", EndDate: ptr("  ")}
	if _, err := svc.CreateSeason(ongoing); err != nil || ongoing.EndDate != nil {
		t.Errorf("sezonul fără sfârșit trebuie acceptat cu EndDate nil: %v, %v", err, ongoing.EndDate)
	}
	if _, err := svc.CreateSeason(&domain.Season{Name: "Fără", StartDate: "2026-08-01"}); err != nil {
		t.Errorf("sezonul fără câmpul end_date: %v", err)
	}
	if _, err := svc.UpdateSeason(1, season); err != nil {
		t.Errorf("UpdateSeason: %v", err)
	}
	if err := svc.DeleteSeason(1); err != nil {
		t.Errorf("DeleteSeason: %v", err)
	}
	if _, err := svc.GetSeasons(); err != nil {
		t.Errorf("GetSeasons: %v", err)
	}

	repo.createSeason = func(*domain.Season) error { return repository.ErrDuplicateEntry }
	if _, err := svc.CreateSeason(season); !errors.Is(err, usecase.ErrCropDuplicate) {
		t.Errorf("duplicat: %v", err)
	}
	repo.updateSeason = func(int64, *domain.Season) error { return sql.ErrNoRows }
	if _, err := svc.UpdateSeason(1, season); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("update inexistent: %v", err)
	}
	repo.updateSeason = nil
	repo.deleteSeason = func(int64) error { return sql.ErrNoRows }
	if err := svc.DeleteSeason(1); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("delete inexistent: %v", err)
	}
}

func TestCropService_Crops(t *testing.T) {
	repo := &cropRepoMock{
		createCrop:  func(c *domain.Crop) error { c.ID = 1; return nil },
		getCropByID: func(id int64) (*domain.Crop, error) { return &domain.Crop{ID: id}, nil },
	}
	svc := usecase.NewCropService(repo, &fieldRepoMock{}, nil, nil)

	if _, err := svc.CreateCrop(&domain.Crop{Name: "  "}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("nume lipsă: %v", err)
	}
	crop := &domain.Crop{Name: " Grâu ", Code: ptr("  "), YieldUnit: " "}
	if _, err := svc.CreateCrop(crop); err != nil {
		t.Fatalf("CreateCrop: %v", err)
	}
	if crop.Name != "Grâu" || crop.Code != nil || crop.YieldUnit != "t" {
		t.Errorf("normalizarea culturii: %+v", crop)
	}
	crop.Code = ptr(" GR ")
	if _, err := svc.UpdateCrop(1, crop); err != nil || *crop.Code != "GR" {
		t.Errorf("UpdateCrop: %v, cod=%v", err, crop.Code)
	}
	if err := svc.DeleteCrop(1); err != nil {
		t.Errorf("DeleteCrop: %v", err)
	}
	if _, err := svc.GetCrops(); err != nil {
		t.Errorf("GetCrops: %v", err)
	}

	repo.createCrop = func(*domain.Crop) error { return repository.ErrDuplicateEntry }
	if _, err := svc.CreateCrop(crop); !errors.Is(err, usecase.ErrCropDuplicate) {
		t.Errorf("duplicat: %v", err)
	}
	repo.updateCrop = func(int64, *domain.Crop) error { return sql.ErrNoRows }
	if _, err := svc.UpdateCrop(1, crop); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("update inexistent: %v", err)
	}
	if _, err := svc.UpdateCrop(1, &domain.Crop{}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("update invalid: %v", err)
	}
}

func TestCropService_FieldCrops(t *testing.T) {
	fields := &fieldRepoMock{getByID: func(id string) (*domain.Field, error) {
		if id == "f1" {
			return &domain.Field{ID: "f1", AreaHa: ptr(10.0)}, nil
		}
		if id == "f2" {
			return &domain.Field{ID: "f2"}, nil
		}
		return nil, nil
	}}
	repo := &cropRepoMock{
		getSeasonByID: func(id int64) (*domain.Season, error) {
			if id == 1 {
				return &domain.Season{ID: 1}, nil
			}
			return nil, nil
		},
		getCropByID: func(id int64) (*domain.Crop, error) {
			if id == 1 {
				return &domain.Crop{ID: 1}, nil
			}
			return nil, nil
		},
		createFieldCrop:  func(fc *domain.FieldCrop) error { fc.ID = 5; return nil },
		getFieldCropByID: func(id int64) (*domain.FieldCrop, error) { return &domain.FieldCrop{ID: id}, nil },
	}
	svc := usecase.NewCropService(repo, fields, nil, nil)

	invalid := map[string]*domain.FieldCrop{
		"câmpuri lipsă":       {FieldID: "f1"},
		"teren inexistent":    {FieldID: "nope", SeasonID: 1, CropID: 1},
		"sezon inexistent":    {FieldID: "f1", SeasonID: 9, CropID: 1},
		"cultură inexistentă": {FieldID: "f1", SeasonID: 1, CropID: 9},
		"suprafață zero":      {FieldID: "f1", SeasonID: 1, CropID: 1, PlantedAreaHa: ptr(0.0)},
		"suprafață prea mare": {FieldID: "f1", SeasonID: 1, CropID: 1, PlantedAreaHa: ptr(11.0)},
		"producție negativă":  {FieldID: "f1", SeasonID: 1, CropID: 1, ProductionTotal: ptr(-1.0)},
		"randament negativ":   {FieldID: "f1", SeasonID: 1, CropID: 1, ExpectedYieldPerHa: ptr(-1.0)},
		"dată invalidă":       {FieldID: "f1", SeasonID: 1, CropID: 1, PlantedAt: ptr("ieri")},
	}
	for name, in := range invalid {
		if _, err := svc.CreateFieldCrop(in); !errors.Is(err, usecase.ErrCropInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}

	item := &domain.FieldCrop{FieldID: "f1", SeasonID: 1, CropID: 1, PlantedAt: ptr("2026-03-01"), HarvestedAt: ptr(" ")}
	created, err := svc.CreateFieldCrop(item)
	if err != nil || created.ID != 5 {
		t.Fatalf("CreateFieldCrop: %v", err)
	}
	if item.PlantedAreaHa == nil || *item.PlantedAreaHa != 10 {
		t.Error("suprafața cultivată trebuie preluată implicit din teren")
	}
	// teren fără suprafață → rămâne nil
	noArea := &domain.FieldCrop{FieldID: "f2", SeasonID: 1, CropID: 1, PlantedAreaHa: ptr(3.0)}
	if _, err := svc.CreateFieldCrop(noArea); err != nil {
		t.Errorf("teren fără suprafață: %v", err)
	}
	if _, err := svc.UpdateFieldCrop(5, item); err != nil {
		t.Errorf("UpdateFieldCrop: %v", err)
	}
	if err := svc.DeleteFieldCrop(5); err != nil {
		t.Errorf("DeleteFieldCrop: %v", err)
	}
	if _, err := svc.ListFieldCrops(domain.FieldCropFilter{}); err != nil {
		t.Errorf("ListFieldCrops: %v", err)
	}

	boom := errors.New("db down")
	repo.createFieldCrop = func(*domain.FieldCrop) error { return repository.ErrDuplicateEntry }
	if _, err := svc.CreateFieldCrop(item); !errors.Is(err, usecase.ErrCropDuplicate) {
		t.Errorf("duplicat: %v", err)
	}
	repo.updateFieldCrop = func(int64, *domain.FieldCrop) error { return sql.ErrNoRows }
	if _, err := svc.UpdateFieldCrop(5, item); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("update inexistent: %v", err)
	}
	fields.getByID = func(string) (*domain.Field, error) { return nil, boom }
	if _, err := svc.CreateFieldCrop(item); !errors.Is(err, boom) {
		t.Errorf("eroarea de citire a terenului: %v", err)
	}
}

func TestCropService_RecordHarvest(t *testing.T) {
	db, mock := newSQLMock(t)
	item := &domain.FieldCrop{ID: 5, CropID: 1, FieldName: "Lot 1", SeasonName: "2026", ProductionTotal: ptr(40.0)}
	repo := &cropRepoMock{
		getFieldCropByID: func(id int64) (*domain.FieldCrop, error) {
			if id == 5 {
				return item, nil
			}
			return nil, nil
		},
		getCropByID: func(id int64) (*domain.Crop, error) {
			if id == 1 {
				return &domain.Crop{ID: 1, Name: "Grâu"}, nil
			}
			return nil, nil
		},
		ensureHarvestResource: func(*sql.Tx, *domain.Crop) (int64, error) { return 77, nil },
	}
	// în tranzacție, producția și cantitatea înregistrată sunt recitite sub blocare
	repo.lockFieldCropHarvest = func(*sql.Tx, int64) (*float64, float64, error) {
		recorded := 0.0
		if item.HarvestRecordedQty != nil {
			recorded = *item.HarvestRecordedQty
		}
		return item.ProductionTotal, recorded, nil
	}
	lock := &repository.StockLock{ResourceID: 77, Quantity: 10, PriceUnit: 1}
	movements := &stockMovementRepoMock{
		lockStock: func(_ *sql.Tx, id int64) (*repository.StockLock, error) {
			if id == 77 {
				return lock, nil
			}
			return nil, nil
		},
	}
	svc := usecase.NewCropService(repo, &fieldRepoMock{}, db, movements)

	if _, err := svc.RecordHarvest(9, nil); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("cultură pe teren inexistentă: %v", err)
	}
	item.ProductionTotal = nil
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, usecase.ErrHarvestNotRecordable) {
		t.Errorf("fără producție: %v", err)
	}
	item.ProductionTotal = ptr(40.0)
	item.CropID = 9
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("cultură inexistentă: %v", err)
	}
	item.CropID = 1

	// nimic de înregistrat: producția e deja în stoc
	item.HarvestRecordedQty = ptr(40.0)
	if res, err := svc.RecordHarvest(5, nil); err != nil || res.Movement != nil {
		t.Errorf("fără diferență: %v, %+v", err, res)
	}

	// prima înregistrare: intrare de 40
	item.HarvestRecordedQty = nil
	var recordedQty float64
	repo.markHarvestRecorded = func(_ *sql.Tx, _ int64, qty float64) error { recordedQty = qty; return nil }
	mock.ExpectBegin()
	mock.ExpectCommit()
	res, err := svc.RecordHarvest(5, ptr(int64(1)))
	if err != nil {
		t.Fatalf("RecordHarvest: %v", err)
	}
	if res.Movement.MovementType != domain.StockMovementIn || res.Movement.QuantityDelta != 40 || *res.Movement.FieldCropID != 5 || recordedQty != 40 {
		t.Errorf("mișcare greșită: %+v", res.Movement)
	}

	// corecție în minus: producția scade de la 40 la 35 → ieșire de 5
	item.HarvestRecordedQty = ptr(40.0)
	item.ProductionTotal = ptr(35.0)
	mock.ExpectBegin()
	mock.ExpectCommit()
	res, err = svc.RecordHarvest(5, nil)
	if err != nil || res.Movement.MovementType != domain.StockMovementOut || res.Movement.QuantityDelta != -5 {
		t.Fatalf("corecție: %v, %+v", err, res.Movement)
	}

	// stocul de recoltă lipsește
	repo.ensureHarvestResource = func(*sql.Tx, *domain.Crop) (int64, error) { return 1, nil }
	mock.ExpectBegin()
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, usecase.ErrHarvestNotRecordable) {
		t.Errorf("stoc lipsă: %v", err)
	}

	boom := errors.New("db down")
	repo.ensureHarvestResource = func(*sql.Tx, *domain.Crop) (int64, error) { return 0, boom }
	mock.ExpectBegin()
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la resursă trebuie propagată: %v", err)
	}
	mock.ExpectBegin().WillReturnError(boom)
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la begin trebuie propagată: %v", err)
	}

	// ce se vede sub blocare are prioritate față de citirea de dinaintea tranzacției
	underLock := func(production *float64, recorded float64, err error) {
		repo.lockFieldCropHarvest = func(*sql.Tx, int64) (*float64, float64, error) { return production, recorded, err }
		mock.ExpectBegin()
	}
	underLock(nil, 0, boom)
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, boom) {
		t.Errorf("eroarea la blocare trebuie propagată: %v", err)
	}
	underLock(nil, 0, sql.ErrNoRows)
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("cultura pe teren ștearsă între timp: %v", err)
	}
	underLock(nil, 0, nil)
	if _, err := svc.RecordHarvest(5, nil); !errors.Is(err, usecase.ErrHarvestNotRecordable) {
		t.Errorf("producția ștearsă între timp: %v", err)
	}
	underLock(ptr(35.0), 35, nil)
	if res, err := svc.RecordHarvest(5, nil); err != nil || res.Movement != nil || res.FieldCrop == nil {
		t.Errorf("o cerere simultană a înregistrat deja producția: %v, %+v", err, res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

// Regula „pe un teren e cel mult o cultură deschisă”: crearea și redeschiderea sunt refuzate cu
// numele culturii care trebuie închisă; închiderea nu cere nimic și nu atinge lucrările.
func TestCropService_OpenCropPerField(t *testing.T) {
	fields := &fieldRepoMock{getByID: func(id string) (*domain.Field, error) {
		return &domain.Field{ID: id, AreaHa: ptr(10.0)}, nil
	}}
	open := &domain.FieldCrop{ID: 7, FieldID: "f1", CropName: "Grâu", SeasonName: "Ciclu toamnă 2025/2026", Status: domain.FieldCropStatusOpen}
	var closed []domain.FieldCropClose
	var reopened []int64
	repo := &cropRepoMock{
		getSeasonByID: func(int64) (*domain.Season, error) { return &domain.Season{ID: 1}, nil },
		getCropByID:   func(int64) (*domain.Crop, error) { return &domain.Crop{ID: 1}, nil },
		getOpenFieldCrop: func(fieldID string) (*domain.FieldCrop, error) {
			if fieldID == "f1" {
				return open, nil
			}
			return nil, nil
		},
		createFieldCrop: func(fc *domain.FieldCrop) error { fc.ID = 9; return nil },
		getFieldCropByID: func(id int64) (*domain.FieldCrop, error) {
			if id == 7 {
				return open, nil
			}
			if id == 8 {
				return &domain.FieldCrop{ID: 8, FieldID: "f1", Status: domain.FieldCropStatusClosed}, nil
			}
			if id == 9 {
				return &domain.FieldCrop{ID: 9, FieldID: "f2", Status: domain.FieldCropStatusOpen}, nil
			}
			return nil, nil
		},
		closeFieldCrop:  func(_ int64, in domain.FieldCropClose) error { closed = append(closed, in); return nil },
		reopenFieldCrop: func(id int64) error { reopened = append(reopened, id); return nil },
	}
	svc := usecase.NewCropService(repo, fields, nil, nil)

	// a doua cultură deschisă pe f1 → 409, cu numele celei deschise
	_, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: "f1", SeasonID: 1, CropID: 1})
	if !errors.Is(err, usecase.ErrFieldHasOpenCrop) || !errors.Is(err, usecase.ErrCropDuplicate) {
		t.Fatalf("a doua cultură deschisă trebuie refuzată ca duplicat: %v", err)
	}
	if !strings.Contains(err.Error(), "Grâu") || !strings.Contains(err.Error(), "Ciclu toamnă 2025/2026") {
		t.Errorf("mesajul trebuie să spună ce cultură e deschisă: %v", err)
	}
	// pe alt teren merge, și starea implicită e open
	in := &domain.FieldCrop{FieldID: "f2", SeasonID: 1, CropID: 1}
	if _, err := svc.CreateFieldCrop(in); err != nil {
		t.Fatalf("cultură pe teren liber: %v", err)
	}
	if in.Status != domain.FieldCropStatusOpen {
		t.Errorf("starea implicită trebuie să fie open, am %q", in.Status)
	}
	// o cultură istorică se poate crea direct închisă, dar cere data închiderii
	hist := &domain.FieldCrop{FieldID: "f1", SeasonID: 1, CropID: 1, Status: domain.FieldCropStatusClosed}
	if _, err := svc.CreateFieldCrop(hist); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("închisă fără dată: %v", err)
	}
	hist.ClosedAt = ptr("2025-07-10")
	if _, err := svc.CreateFieldCrop(hist); err != nil {
		t.Errorf("cultură istorică închisă pe un teren cu alta deschisă trebuie acceptată: %v", err)
	}
	if _, err := svc.CreateFieldCrop(&domain.FieldCrop{FieldID: "f2", SeasonID: 1, CropID: 1, Status: "paused"}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Error("o stare necunoscută trebuie refuzată")
	}

	// închiderea: compromisă ⇒ producție 0 și observații obligatorii
	if _, err := svc.CloseFieldCrop(7, domain.FieldCropClose{Status: "compromised", ClosedAt: "2026-02-28"}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("compromisă fără observații: %v", err)
	}
	if _, err := svc.CloseFieldCrop(7, domain.FieldCropClose{Status: "closed", ClosedAt: "azi"}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("dată invalidă: %v", err)
	}
	if _, err := svc.CloseFieldCrop(7, domain.FieldCropClose{Status: "open", ClosedAt: "2026-02-28"}); !errors.Is(err, usecase.ErrCropInvalid) {
		t.Errorf("închiderea nu poate duce la open: %v", err)
	}
	if _, err := svc.CloseFieldCrop(7, domain.FieldCropClose{Status: "compromised", ClosedAt: "2026-02-28", ProductionTotal: ptr(5.0), Notes: ptr("îngheț")}); err != nil {
		t.Fatalf("CloseFieldCrop: %v", err)
	}
	if len(closed) != 1 || closed[0].ProductionTotal == nil || *closed[0].ProductionTotal != 0 {
		t.Errorf("la compromisă producția devine 0: %+v", closed)
	}
	if _, err := svc.CloseFieldCrop(404, domain.FieldCropClose{Status: "closed", ClosedAt: "2026-07-10"}); !errors.Is(err, usecase.ErrCropNotFound) {
		t.Errorf("inexistentă: %v", err)
	}

	// redeschiderea: refuzată cât timp terenul are alta deschisă; altfel merge; deschisă deja = no-op
	if _, err := svc.ReopenFieldCrop(8); !errors.Is(err, usecase.ErrFieldHasOpenCrop) {
		t.Errorf("redeschidere cu alta deschisă pe teren: %v", err)
	}
	repo.getOpenFieldCrop = func(string) (*domain.FieldCrop, error) { return nil, nil }
	if _, err := svc.ReopenFieldCrop(8); err != nil || len(reopened) != 1 || reopened[0] != 8 {
		t.Errorf("ReopenFieldCrop: %v, %v", err, reopened)
	}
	if _, err := svc.ReopenFieldCrop(9); err != nil || len(reopened) != 1 {
		t.Errorf("redeschiderea unei culturi deschise nu face nimic: %v, %v", err, reopened)
	}
	repo.reopenFieldCrop = func(int64) error { return repository.ErrDuplicateEntry }
	if _, err := svc.ReopenFieldCrop(8); !errors.Is(err, usecase.ErrFieldHasOpenCrop) {
		t.Errorf("cursa pe indexul unic trebuie raportată ca 409: %v", err)
	}

	// modificarea sezonului nu mai are nimic de relegat: nu apelează niciun repo în afară de update
	repo.updateSeason = func(int64, *domain.Season) error { return nil }
	if _, err := svc.UpdateSeason(1, &domain.Season{Name: "Ciclu", StartDate: "2025-08-01", EndDate: ptr("2026-07-15")}); err != nil {
		t.Errorf("UpdateSeason: %v", err)
	}
}
