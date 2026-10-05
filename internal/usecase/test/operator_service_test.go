package usecase_test

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
)

// operatorStore simulează conturile operatorilor: repository-ul întoarce ce a salvat.
func operatorStore() (*operatorRepoMock, map[int64]*domain.Operator) {
	stored := map[int64]*domain.Operator{}
	repo := &operatorRepoMock{
		getByID: func(id int64) (*domain.Operator, error) {
			if o, ok := stored[id]; ok {
				copy := *o
				return &copy, nil
			}
			return nil, nil
		},
		create: func(o *domain.Operator) error {
			o.ID = int64(len(stored) + 1)
			o.Status = domain.OperatorStatusActive
			copy := *o
			stored[o.ID] = &copy
			return nil
		},
		update: func(id int64, o *domain.Operator) error {
			copy := *o
			copy.ID, copy.Status = id, stored[id].Status
			stored[id] = &copy
			return nil
		},
		setActive: func(id int64, active bool) error {
			stored[id].Status = domain.OperatorStatusInactive
			if active {
				stored[id].Status = domain.OperatorStatusActive
			}
			return nil
		},
		getAll: func() ([]domain.Operator, error) {
			items := []domain.Operator{}
			for _, o := range stored {
				items = append(items, *o)
			}
			return items, nil
		},
	}
	return repo, stored
}

func TestOperatorService_Create(t *testing.T) {
	repo, _ := operatorStore()
	svc := usecase.NewOperatorService(repo)

	for name, in := range map[string]*domain.Operator{
		"nil":            nil,
		"fără prenume":   {LastName: "Popescu"},
		"adresă tehnică": {FirstName: "Ion", Email: "operator-1@fara-email.local"},
	} {
		if _, err := svc.CreateOperator(in); !errors.Is(err, usecase.ErrOperatorInvalid) {
			t.Errorf("%s: %v", name, err)
		}
	}

	created, err := svc.CreateOperator(&domain.Operator{FirstName: "Ion", Phone: "0700", AllowedMachineTypes: []domain.MachineType{domain.MachineTypeTractor}})
	if err != nil {
		t.Fatalf("eroare neașteptată: %v", err)
	}
	if created.ID != 1 || created.Status != domain.OperatorStatusActive || len(created.AllowedMachineTypes) != 1 {
		t.Errorf("operator creat greșit: %+v", created)
	}

	repo.create = func(*domain.Operator) error { return repository.ErrEmailTaken }
	if _, err := svc.CreateOperator(&domain.Operator{FirstName: "Ion", Email: "ion@x.ro"}); !errors.Is(err, usecase.ErrOperatorInvalid) {
		t.Errorf("e-mailul folosit trebuie să fie eroare de validare: %v", err)
	}
	repo.create = func(*domain.Operator) error { return errors.New("db down") }
	if _, err := svc.CreateOperator(&domain.Operator{FirstName: "Ion"}); err == nil || errors.Is(err, usecase.ErrOperatorInvalid) {
		t.Errorf("eroarea din repo trebuie propagată: %v", err)
	}
}

func TestOperatorService_UpdateDeleteAndStatus(t *testing.T) {
	repo, stored := operatorStore()
	stored[1] = &domain.Operator{ID: 1, FirstName: "Vechi", Status: domain.OperatorStatusActive}
	svc := usecase.NewOperatorService(repo)

	if _, err := svc.UpdateOperator(2, &domain.Operator{FirstName: "X"}); !errors.Is(err, usecase.ErrOperatorNotFound) {
		t.Errorf("operatorul inexistent: %v", err)
	}
	if _, err := svc.UpdateOperator(1, &domain.Operator{}); !errors.Is(err, usecase.ErrOperatorInvalid) {
		t.Errorf("prenumele lipsă: %v", err)
	}
	updated, err := svc.UpdateOperator(1, &domain.Operator{FirstName: "Nou", Email: "nou@x.ro"})
	if err != nil || updated.FirstName != "Nou" || updated.Email != "nou@x.ro" {
		t.Fatalf("UpdateOperator: %v, %+v", err, updated)
	}

	// ștergerea dezactivează contul; lucrările îl păstrează ca operator
	if err := svc.DeleteOperator(2); !errors.Is(err, usecase.ErrOperatorNotFound) {
		t.Errorf("ștergerea unui operator inexistent: %v", err)
	}
	if err := svc.DeleteOperator(1); err != nil || stored[1].Status != domain.OperatorStatusInactive {
		t.Errorf("DeleteOperator: %v, %+v", err, stored[1])
	}
	enabled, err := svc.EnableOperator(1)
	if err != nil || enabled.Status != domain.OperatorStatusActive {
		t.Fatalf("EnableOperator: %v, %+v", err, enabled)
	}
	disabled, err := svc.DisableOperator(1)
	if err != nil || disabled.Status != domain.OperatorStatusInactive {
		t.Fatalf("DisableOperator: %v, %+v", err, disabled)
	}
	if _, err := svc.EnableOperator(2); !errors.Is(err, usecase.ErrOperatorNotFound) {
		t.Errorf("activarea unui operator inexistent: %v", err)
	}

	if all, err := svc.GetOperators(); err != nil || len(all) != 1 {
		t.Errorf("GetOperators: %v, %d", err, len(all))
	}

	boom := errors.New("db down")
	repo.setActive = func(int64, bool) error { return boom }
	if _, err := svc.DisableOperator(1); !errors.Is(err, boom) {
		t.Error("eroarea de status trebuie propagată")
	}
	repo.getByID = func(int64) (*domain.Operator, error) { return nil, boom }
	if _, err := svc.GetOperatorByID(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteOperator(1); !errors.Is(err, boom) {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
}
