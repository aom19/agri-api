package usecase

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
)

func TestOperatorService_Create(t *testing.T) {
	repo := &operatorRepoMock{create: func(o *domain.Operator) error { o.ID = 3; return nil }}
	svc := NewOperatorService(repo)

	if _, err := svc.CreateOperator(&domain.Operator{}); err == nil {
		t.Error("numele lipsă trebuie să dea eroare")
	}

	created, err := svc.CreateOperator(&domain.Operator{Name: "Ion", Phone: "0700", AllowedMachineTypes: []domain.MachineType{domain.MachineTypeTractor}})
	if err != nil {
		t.Fatalf("eroare neașteptată: %v", err)
	}
	if created.ID != 3 || created.Status != domain.OperatorStatusActive || len(created.AllowedMachineTypes) != 1 {
		t.Errorf("operator creat greșit: %+v", created)
	}

	repo.create = func(*domain.Operator) error { return errors.New("db down") }
	if _, err := svc.CreateOperator(&domain.Operator{Name: "Ion"}); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
}

func TestOperatorService_UpdateDeleteAndStatus(t *testing.T) {
	var stored *domain.Operator
	repo := &operatorRepoMock{
		getByID: func(id int64) (*domain.Operator, error) {
			if id == 1 {
				return &domain.Operator{ID: 1, Name: "Vechi", Status: domain.OperatorStatusActive}, nil
			}
			return nil, nil
		},
		update: func(_ int64, o *domain.Operator) error { stored = o; return nil },
		getAll: func() ([]domain.Operator, error) { return []domain.Operator{{ID: 1}}, nil },
	}
	svc := NewOperatorService(repo)

	if _, err := svc.UpdateOperator(2, &domain.Operator{Name: "X"}); err == nil {
		t.Error("operatorul inexistent trebuie să dea eroare")
	}
	updated, err := svc.UpdateOperator(1, &domain.Operator{Name: "Nou", Email: "nou@x.ro"})
	if err != nil || updated.Name != "Nou" || stored == nil || stored.Email != "nou@x.ro" {
		t.Fatalf("UpdateOperator: %v, %+v", err, updated)
	}

	if err := svc.DeleteOperator(2); err == nil {
		t.Error("ștergerea unui operator inexistent trebuie să dea eroare")
	}
	if err := svc.DeleteOperator(1); err != nil {
		t.Errorf("DeleteOperator: %v", err)
	}

	disabled, err := svc.DisableOperator(1)
	if err != nil || disabled.Status != domain.OperatorStatusInactive {
		t.Fatalf("DisableOperator: %v, %+v", err, disabled)
	}
	enabled, err := svc.EnableOperator(1)
	if err != nil || enabled.Status != domain.OperatorStatusActive {
		t.Fatalf("EnableOperator: %v, %+v", err, enabled)
	}
	if _, err := svc.DisableOperator(2); err == nil {
		t.Error("dezactivarea unui operator inexistent trebuie să dea eroare")
	}
	if _, err := svc.EnableOperator(2); err == nil {
		t.Error("activarea unui operator inexistent trebuie să dea eroare")
	}

	all, err := svc.GetOperators()
	if err != nil || len(all) != 1 {
		t.Errorf("GetOperators: %v, %d", err, len(all))
	}
	one, err := svc.GetOperatorByID(1)
	if err != nil || one == nil {
		t.Errorf("GetOperatorByID: %v", err)
	}

	repo.updateStatusDirect = func(int64, domain.OperatorStatus) error { return errors.New("db down") }
	if _, err := svc.DisableOperator(1); err == nil {
		t.Error("eroarea de status trebuie propagată")
	}
	if _, err := svc.EnableOperator(1); err == nil {
		t.Error("eroarea de status trebuie propagată")
	}
	repo.getByID = func(int64) (*domain.Operator, error) { return nil, errors.New("db down") }
	if _, err := svc.GetOperatorByID(1); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := svc.DeleteOperator(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la ștergere")
	}
}
