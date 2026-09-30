package usecase

import (
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
)

func readyOperation() *dto.FieldOperationResponse {
	return &dto.FieldOperationResponse{
		ID: 1, Status: string(domain.FieldOperationStatusPlanned),
		MachineID: ptr(int64(1)), MachineStatus: ptr("active"),
		ImplementID: ptr(int64(1)), ImplementStatus: ptr("active"),
		Checklist: dto.FieldOperationChecklistResponse{MachineStatus: true, ImplementStatus: true, FieldArea: true, NotesConfirmed: true},
	}
}

func TestFieldOperationService_Validate(t *testing.T) {
	svc := NewFieldOperationService(&fieldOperationRepoMock{})
	start := time.Now()
	before := start.Add(-time.Hour)

	cases := map[string]struct {
		in  *domain.FieldOperation
		err error
	}{
		"teren lipsă":    {&domain.FieldOperation{OperationTypeID: 1}, ErrFieldOperationFieldRequired},
		"tip lipsă":      {&domain.FieldOperation{FieldID: "f"}, ErrFieldOperationTypeRequired},
		"status invalid": {&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, Status: "x"}, ErrFieldOperationInvalidStatus},
		"date inversate": {&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, PlannedStartAt: &start, PlannedEndAt: &before}, ErrFieldOperationBadDates},
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
	svc := NewFieldOperationService(repo)

	created, err := svc.Create(&domain.FieldOperation{FieldID: "f", OperationTypeID: 1, Notes: "  note  "})
	if err != nil || created.ID != 5 {
		t.Fatalf("Create: %v, %+v", err, created)
	}
	if saved.Status != domain.FieldOperationStatusPlanned || saved.Notes != "note" {
		t.Errorf("statusul implicit / notele nu au fost normalizate: %+v", saved)
	}

	if _, err := svc.Update(9, &domain.FieldOperation{FieldID: "f", OperationTypeID: 1}); !errors.Is(err, ErrFieldOperationNotFound) {
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

	if _, err := svc.GetByID(9); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("GetByID inexistent: %v", err)
	}
	if op, err := svc.GetByID(1); err != nil || op.ID != 1 {
		t.Errorf("GetByID: %v", err)
	}
	if _, err := svc.GetAll(repository.FieldOperationFilter{}); err != nil {
		t.Errorf("GetAll: %v", err)
	}
	if err := svc.Delete(9); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("Delete inexistent: %v", err)
	}
	if err := svc.Delete(1); err != nil {
		t.Errorf("Delete: %v", err)
	}

	if _, err := svc.UpdateChecklist(9, domain.FieldOperationChecklist{}); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("UpdateChecklist inexistent: %v", err)
	}
	if _, err := svc.UpdateChecklist(1, domain.FieldOperationChecklist{MachineStatus: true}); err != nil {
		t.Errorf("UpdateChecklist: %v", err)
	}

	boom := errors.New("db down")
	repo.updateChecklist = func(int64, domain.FieldOperationChecklist) error { return boom }
	if _, err := svc.UpdateChecklist(1, domain.FieldOperationChecklist{}); !errors.Is(err, boom) {
		t.Error("eroarea de checklist trebuie propagată")
	}
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
	svc := NewFieldOperationService(repo)

	if _, err := svc.Start(9); !errors.Is(err, ErrFieldOperationNotFound) {
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
	if _, err := svc.Start(1); !errors.Is(err, ErrFieldOperationCannotStart) {
		t.Errorf("start pe operațiune finalizată: %v", err)
	}

	current = readyOperation()
	current.Checklist.FieldArea = false
	if _, err := svc.Start(1); !errors.Is(err, ErrFieldOperationChecklistIncomplete) {
		t.Errorf("checklist incomplet: %v", err)
	}

	current = readyOperation()
	current.MachineStatus = ptr("maintenance")
	if _, err := svc.Start(1); !errors.Is(err, ErrFieldOperationResourcesUnavailable) {
		t.Errorf("mașină indisponibilă: %v", err)
	}

	current = readyOperation()
	current.ImplementStatus = nil
	if _, err := svc.Start(1); !errors.Is(err, ErrFieldOperationResourcesUnavailable) {
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
	svc := NewFieldOperationService(repo)

	if _, err := svc.GetByIDForAssignedUser(1, 8); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("alt utilizator nu trebuie să vadă operațiunea: %v", err)
	}
	if op, err := svc.GetByIDForAssignedUser(1, 7); err != nil || op.ID != 1 {
		t.Errorf("GetByIDForAssignedUser: %v", err)
	}

	if _, err := svc.UpdateChecklistForAssignedUser(1, 8, domain.FieldOperationChecklist{}); !errors.Is(err, ErrFieldOperationNotFound) {
		t.Errorf("checklist pentru alt utilizator: %v", err)
	}
	if _, err := svc.UpdateChecklistForAssignedUser(1, 7, domain.FieldOperationChecklist{}); err != nil {
		t.Errorf("UpdateChecklistForAssignedUser: %v", err)
	}

	if _, err := svc.StartForAssignedUser(1, 8); !errors.Is(err, ErrFieldOperationNotFound) {
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
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, ErrFieldOperationCannotStart) {
		t.Errorf("start pe operațiune anulată: %v", err)
	}

	current = readyOperation()
	boom := errors.New("db down")
	repo.markStarted = func(int64, time.Time) error { return boom }
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de pornire trebuie propagată")
	}
	repo.updateChecklist = func(int64, domain.FieldOperationChecklist) error { return boom }
	if _, err := svc.UpdateChecklistForAssignedUser(1, 7, domain.FieldOperationChecklist{}); !errors.Is(err, boom) {
		t.Error("eroarea de checklist trebuie propagată")
	}
	repo.getByIDForAssignedUser = func(int64, int64) (*dto.FieldOperationResponse, error) { return nil, boom }
	if _, err := svc.GetByIDForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if _, err := svc.StartForAssignedUser(1, 7); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la start")
	}
	if _, err := svc.UpdateChecklistForAssignedUser(1, 7, domain.FieldOperationChecklist{}); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la checklist")
	}
}
