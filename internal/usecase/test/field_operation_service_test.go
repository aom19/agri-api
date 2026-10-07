package usecase_test

import (
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
)

func readyOperation() *dto.FieldOperationResponse {
	return &dto.FieldOperationResponse{
		ID: 1, Status: string(domain.FieldOperationStatusPlanned),
		MachineID: ptr(int64(1)), MachineStatus: ptr("active"),
		ImplementID: ptr(int64(1)), ImplementStatus: ptr("active"),
	}
}

func TestFieldOperationService_Validate(t *testing.T) {
	svc := usecase.NewFieldOperationService(&fieldOperationRepoMock{})
	start := time.Now()
	before := start.Add(-time.Hour)

	cases := map[string]struct {
		in  *domain.FieldOperation
		err error
	}{
		"teren lipsă":    {&domain.FieldOperation{OperationTypeID: 1}, usecase.ErrFieldOperationFieldRequired},
		"tip lipsă":      {&domain.FieldOperation{FieldID: "f"}, usecase.ErrFieldOperationTypeRequired},
		"status invalid": {&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, Status: "x"}, usecase.ErrFieldOperationInvalidStatus},
		"date inversate": {&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, PlannedStartAt: &start, PlannedEndAt: &before}, usecase.ErrFieldOperationBadDates},
	}
	for name, c := range cases {
		if _, err := svc.Create(c.in); !errors.Is(err, c.err) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if _, err := svc.Create(nil); err == nil {
		t.Error("payload nil trebuie să dea eroare")
	}
	if _, err := svc.Create(&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, AreaPlannedHa: ptr(-1.0)}); err == nil {
		t.Error("suprafața negativă trebuie să dea eroare")
	}
}

func TestFieldOperationService_CreateUpdateDelete(t *testing.T) {
	var saved *domain.FieldOperation
	repo := &fieldOperationRepoMock{
		create: func(op *domain.FieldOperation) error { op.ID = 5; saved = op; return nil },
		getByID: func(id int64) (*dto.FieldOperationResponse, error) {
			if id == 5 || id == 1 {
				return &dto.FieldOperationResponse{ID: id, Status: "planned"}, nil
			}
			return nil, nil
		},
	}
	svc := usecase.NewFieldOperationService(repo)

	created, err := svc.Create(&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, Notes: "  note  "})
	if err != nil || created.ID != 5 {
		t.Fatalf("Create: %v, %+v", err, created)
	}
	if saved.Status != domain.FieldOperationStatusPlanned || saved.Notes != "note" {
		t.Errorf("statusul implicit / notele nu au fost normalizate: %+v", saved)
	}

	if _, err := svc.Update(9, &domain.FieldOperation{FieldID: "f", OperationTypeID: 1}); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("update pe operațiune inexistentă: %v", err)
	}
	if _, err := svc.Update(1, &domain.FieldOperation{}); err == nil {
		t.Error("update cu payload invalid trebuie să dea eroare")
	}
	var updated *domain.FieldOperation
	repo.update = func(_ int64, op *domain.FieldOperation) error { updated = op; return nil }
	if _, err := svc.Update(1, &domain.FieldOperation{FieldID: "f", OperationTypeID: 1}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Status != domain.FieldOperationStatusPlanned {
		t.Error("statusul existent trebuie păstrat când lipsește din payload")
	}

	if _, err := svc.GetByID(9); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("GetByID inexistent: %v", err)
	}
	if op, err := svc.GetByID(1); err != nil || op.ID != 1 {
		t.Errorf("GetByID: %v", err)
	}
	if _, err := svc.GetAll(repository.FieldOperationFilter{}); err != nil {
		t.Errorf("GetAll: %v", err)
	}
	if err := svc.Delete(9); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("Delete inexistent: %v", err)
	}
	if err := svc.Delete(1); err != nil {
		t.Errorf("Delete: %v", err)
	}

	boom := errors.New("db down")
	repo.create = func(*domain.FieldOperation) error { return boom }
	if _, err := svc.Create(&domain.FieldOperation{FieldID: "f", OperationTypeID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de creare trebuie propagată")
	}
	repo.getByID = func(int64) (*dto.FieldOperationResponse, error) { return nil, boom }
	if _, err := svc.GetByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if _, err := svc.Update(1, &domain.FieldOperation{FieldID: "f", OperationTypeID: 1}); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la update")
	}
	if err := svc.Delete(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la delete")
	}
}

func TestFieldOperationService_Start(t *testing.T) {
	current := readyOperation()
	started := false
	repo := &fieldOperationRepoMock{
		getByID: func(id int64) (*dto.FieldOperationResponse, error) {
			if id == 1 {
				return current, nil
			}
			return nil, nil
		},
		markStarted: func(int64, time.Time) error { started = true; return nil },
	}
	svc := usecase.NewFieldOperationService(repo)

	if _, err := svc.Start(9); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("start pe operațiune inexistentă: %v", err)
	}

	if _, err := svc.Start(1); err != nil || !started {
		t.Fatalf("Start: %v (started=%v)", err, started)
	}

	// deja în lucru → nu se mai marchează pornirea
	started = false
	current.Status = string(domain.FieldOperationStatusInProgress)
	if op, err := svc.Start(1); err != nil || started || op != current {
		t.Errorf("start pe operațiune în lucru trebuie să fie idempotent: %v", err)
	}

	current.Status = string(domain.FieldOperationStatusCompleted)
	if _, err := svc.Start(1); !errors.Is(err, usecase.ErrFieldOperationCannotStart) {
		t.Errorf("start pe operațiune finalizată: %v", err)
	}

	// fără mașină și echipament, pornirea nu cere nimic în plus
	current = &dto.FieldOperationResponse{ID: 1, Status: string(domain.FieldOperationStatusPlanned)}
	if _, err := svc.Start(1); err != nil {
		t.Errorf("start fără mașină și echipament: %v", err)
	}

	current = readyOperation()
	current.MachineStatus = ptr("maintenance")
	if _, err := svc.Start(1); !errors.Is(err, usecase.ErrFieldOperationResourcesUnavailable) {
		t.Errorf("mașină indisponibilă: %v", err)
	}

	current = readyOperation()
	current.ImplementStatus = nil
	if _, err := svc.Start(1); !errors.Is(err, usecase.ErrFieldOperationResourcesUnavailable) {
		t.Errorf("echipament fără status: %v", err)
	}

	current = readyOperation()
	boom := errors.New("db down")
	repo.markStarted = func(int64, time.Time) error { return boom }
	if _, err := svc.Start(1); !errors.Is(err, boom) {
		t.Error("eroarea de pornire trebuie propagată")
	}
}

func TestFieldOperationService_AssignedUser(t *testing.T) {
	current := readyOperation()
	repo := &fieldOperationRepoMock{
		getByIDForAssignedUser: func(id, userID int64) (*dto.FieldOperationResponse, error) {
			if id == 1 && userID == 7 {
				return current, nil
			}
			return nil, nil
		},
	}
	svc := usecase.NewFieldOperationService(repo)

	if _, err := svc.GetByIDForAssignedUser(1, 8); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("alt utilizator nu trebuie să vadă operațiunea: %v", err)
	}
	if op, err := svc.GetByIDForAssignedUser(1, 7); err != nil || op.ID != 1 {
		t.Errorf("GetByIDForAssignedUser: %v", err)
	}

	if _, err := svc.StartForAssignedUser(1, 8); !errors.Is(err, usecase.ErrFieldOperationNotFound) {
		t.Errorf("start pentru alt utilizator: %v", err)
	}
	if _, err := svc.StartForAssignedUser(1, 7); err != nil {
		t.Errorf("StartForAssignedUser: %v", err)
	}
	current.Status = string(domain.FieldOperationStatusInProgress)
	if _, err := svc.StartForAssignedUser(1, 7); err != nil {
		t.Errorf("StartForAssignedUser idempotent: %v", err)
	}
	current.Status = string(domain.FieldOperationStatusCanceled)
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, usecase.ErrFieldOperationCannotStart) {
		t.Errorf("start pe operațiune anulată: %v", err)
	}

	current = readyOperation()
	boom := errors.New("db down")
	repo.markStarted = func(int64, time.Time) error { return boom }
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de pornire trebuie propagată")
	}
	repo.getByIDForAssignedUser = func(int64, int64) (*dto.FieldOperationResponse, error) { return nil, boom }
	if _, err := svc.GetByIDForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la start")
	}
}

// T11: mașina și echipamentul trebuie să fie de un tip acceptat de template, și la creare, și la editare.
func TestFieldOperationService_AssetCompatibility(t *testing.T) {
	compat := &domain.AssetCompatibility{
		TemplateMachineTypes:   []string{"tractor"},
		TemplateImplementTypes: []string{"plow", "seeder"},
		MachineType:            "tractor",
		ImplementType:          "plow",
	}
	var asked []any
	repo := &fieldOperationRepoMock{
		getAssetCompatibility: func(templateID int64, machineID, implementID *int64) (*domain.AssetCompatibility, error) {
			asked = []any{templateID, machineID, implementID}
			return compat, nil
		},
		getByID: func(int64) (*dto.FieldOperationResponse, error) {
			return &dto.FieldOperationResponse{ID: 1, Status: "planned"}, nil
		},
	}
	svc := usecase.NewFieldOperationService(repo)
	input := func() *domain.FieldOperation {
		return &domain.FieldOperation{
			FieldID: "f", OperationTypeID: 1,
			OperationTemplateID: ptr(int64(3)), MachineID: ptr(int64(4)), ImplementID: ptr(int64(5)),
		}
	}

	if _, err := svc.Create(input()); err != nil {
		t.Fatalf("mașină și echipament potrivite: %v", err)
	}
	if asked[0] != int64(3) || *asked[1].(*int64) != 4 || *asked[2].(*int64) != 5 {
		t.Errorf("verificarea a primit: %v", asked)
	}

	compat.MachineType = "combine"
	if _, err := svc.Create(input()); !errors.Is(err, usecase.ErrFieldOperationMachineIncompatible) {
		t.Errorf("mașină nepotrivită la creare: %v", err)
	}
	if _, err := svc.Update(1, input()); !errors.Is(err, usecase.ErrFieldOperationMachineIncompatible) {
		t.Errorf("mașină nepotrivită la editare: %v", err)
	}

	compat.MachineType = "tractor"
	compat.ImplementType = "header"
	if _, err := svc.Create(input()); !errors.Is(err, usecase.ErrFieldOperationImplementIncompatible) {
		t.Errorf("echipament nepotrivit: %v", err)
	}

	// template fără tipuri: orice mașină și echipament
	compat.TemplateMachineTypes, compat.TemplateImplementTypes = nil, nil
	compat.MachineType = "combine"
	if _, err := svc.Create(input()); err != nil {
		t.Errorf("template fără tipuri: %v", err)
	}

	// fără template sau fără mașină și echipament, regula nu se aplică
	compat.TemplateMachineTypes = []string{"tractor"}
	asked = nil
	noTemplate := input()
	noTemplate.OperationTemplateID = nil
	noAssets := input()
	noAssets.MachineID, noAssets.ImplementID = nil, nil
	for name, in := range map[string]*domain.FieldOperation{"fără template": noTemplate, "fără mașină și echipament": noAssets} {
		if _, err := svc.Create(in); err != nil || asked != nil {
			t.Errorf("%s: %v (verificare apelată: %v)", name, err, asked != nil)
		}
	}

	boom := errors.New("db down")
	repo.getAssetCompatibility = func(int64, *int64, *int64) (*domain.AssetCompatibility, error) { return nil, boom }
	if _, err := svc.Create(input()); !errors.Is(err, boom) {
		t.Error("eroarea de citire a compatibilității trebuie propagată")
	}
}
