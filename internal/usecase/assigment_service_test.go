package usecase

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/store"

	"github.com/DATA-DOG/go-sqlmock"
)

func newAssigmentService(t *testing.T) (*AssigmentService, *store.Store, *assigmentRepoMock, *machineRepoMock, *operatorRepoMock, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newSQLMock(t)
	assignments := &assigmentRepoMock{
		create: func(_ *sql.Tx, a *domain.Assigment) error { a.ID = 10; return nil },
		getByID: func(id int64) (*domain.Assigment, error) {
			switch id {
			case 1:
				return &domain.Assigment{ID: 1, MachineID: 1, OperatorID: 1, Status: domain.AssigmentStatusActive}, nil
			case 2:
				return &domain.Assigment{ID: 2, MachineID: 1, OperatorID: 1, Status: domain.AssigmentStatusClosed}, nil
			}
			return nil, nil
		},
		isAssigmentActive: func(id int64) (bool, error) { return id == 1, nil },
	}
	machines := &machineRepoMock{getByID: func(id int64) (*domain.Machine, error) {
		if id == 1 {
			return &domain.Machine{ID: 1}, nil
		}
		return nil, nil
	}}
	operators := &operatorRepoMock{getByID: func(id int64) (*domain.Operator, error) {
		if id == 1 {
			return &domain.Operator{ID: 1}, nil
		}
		return nil, nil
	}}
	st := &store.Store{DB: db, AssigmentRepo: assignments, MachineRepo: machines, OperatorRepo: operators}
	return NewAssigmentService(st), st, assignments, machines, operators, mock
}

func TestAssigmentService_Read(t *testing.T) {
	svc, _, assignments, _, _, _ := newAssigmentService(t)
	if _, err := svc.GetAssigments(); err != nil {
		t.Errorf("GetAssigments: %v", err)
	}
	if _, err := svc.GetAll(dto.PaginationQuery{Page: 1, Limit: 10}); err != nil {
		t.Errorf("GetAll: %v", err)
	}
	if a, err := svc.GetAssigmentByID(1); err != nil || a.ID != 1 {
		t.Errorf("GetAssigmentByID: %v", err)
	}
	assignments.getByID = func(int64) (*domain.Assigment, error) { return nil, errors.New("db down") }
	if _, err := svc.GetAssigmentByID(1); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}

func TestAssigmentService_Create(t *testing.T) {
	svc, _, assignments, _, operators, mock := newAssigmentService(t)
	input := &domain.Assigment{MachineID: 1, OperatorID: 1, StartDate: time.Now()}

	var statusSet domain.OperatorStatus
	operators.updateStatus = func(_ *sql.Tx, _ int64, s domain.OperatorStatus) error { statusSet = s; return nil }
	mock.ExpectBegin()
	mock.ExpectCommit()
	created, err := svc.CreateAssigment(input)
	if err != nil || created.ID != 10 || created.Status != domain.AssigmentStatusActive || statusSet != domain.OperatorStatusActive {
		t.Fatalf("CreateAssigment: %v, %+v", err, created)
	}

	failures := map[string]func(){
		"mașină inexistentă":  func() { input.MachineID = 9 },
		"operator inexistent": func() { input.MachineID = 1; input.OperatorID = 9 },
		"mașină alocată": func() {
			input.OperatorID = 1
			assignments.getActiveByMachine = func(int64) (*domain.Assigment, error) { return &domain.Assigment{}, nil }
		},
		"operator alocat": func() {
			assignments.getActiveByMachine = nil
			assignments.getActiveByOperator = func(int64) (*domain.Assigment, error) { return &domain.Assigment{}, nil }
		},
		"eroare la creare": func() {
			assignments.getActiveByOperator = nil
			assignments.create = func(*sql.Tx, *domain.Assigment) error { return errors.New("db down") }
		},
	}
	for name, arrange := range failures {
		arrange()
		mock.ExpectBegin()
		mock.ExpectRollback()
		if _, err := svc.CreateAssigment(input); err == nil {
			t.Errorf("%s: mă așteptam la eroare", name)
		}
	}
	mock.ExpectBegin().WillReturnError(errors.New("db down"))
	if _, err := svc.CreateAssigment(input); err == nil {
		t.Error("eroarea la begin trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAssigmentService_UpdateDeleteClose(t *testing.T) {
	svc, _, assignments, _, operators, mock := newAssigmentService(t)

	if _, err := svc.UpdateAssigment(9, &domain.Assigment{}); err == nil {
		t.Error("update pe alocare inexistentă trebuie să dea eroare")
	}
	end := time.Now()
	updated, err := svc.UpdateAssigment(1, &domain.Assigment{MachineID: 2, OperatorID: 3, EndDate: &end, Status: domain.AssigmentStatusClosed})
	if err != nil || updated.MachineID != 2 || updated.OperatorID != 3 || updated.Status != domain.AssigmentStatusClosed {
		t.Fatalf("UpdateAssigment: %v, %+v", err, updated)
	}

	if err := svc.DeleteAssigment(9); err == nil {
		t.Error("ștergerea unei alocări inexistente trebuie să dea eroare")
	}
	if err := svc.DeleteAssigment(1); err == nil {
		t.Error("alocarea activă nu poate fi ștearsă")
	}
	if err := svc.DeleteAssigment(2); err != nil {
		t.Errorf("DeleteAssigment: %v", err)
	}
	assignments.isAssigmentActive = func(int64) (bool, error) { return false, errors.New("db down") }
	if err := svc.DeleteAssigment(2); err == nil {
		t.Error("eroarea de verificare trebuie propagată")
	}

	if _, err := svc.CloseAssigment(9); err == nil {
		t.Error("închiderea unei alocări inexistente trebuie să dea eroare")
	}
	if _, err := svc.CloseAssigment(2); err == nil {
		t.Error("alocarea deja închisă trebuie să dea eroare")
	}
	var statusSet domain.OperatorStatus
	operators.updateStatus = func(_ *sql.Tx, _ int64, s domain.OperatorStatus) error { statusSet = s; return nil }
	mock.ExpectBegin()
	mock.ExpectCommit()
	if ok, err := svc.CloseAssigment(1); err != nil || !ok || statusSet != domain.OperatorStatusInactive {
		t.Fatalf("CloseAssigment: %v, %v", err, ok)
	}
	assignments.update = func(int64, *domain.Assigment) error { return errors.New("db down") }
	mock.ExpectBegin()
	mock.ExpectRollback()
	if _, err := svc.CloseAssigment(1); err == nil {
		t.Error("eroarea de actualizare trebuie propagată")
	}
	if _, err := svc.UpdateAssigment(1, &domain.Assigment{}); err == nil {
		t.Error("eroarea de actualizare trebuie propagată la update")
	}
	assignments.getByID = func(int64) (*domain.Assigment, error) { return nil, errors.New("db down") }
	if _, err := svc.UpdateAssigment(1, &domain.Assigment{}); err == nil {
		t.Error("eroarea de citire trebuie propagată la update")
	}
	if err := svc.DeleteAssigment(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la delete")
	}
	if _, err := svc.CloseAssigment(1); err == nil {
		t.Error("eroarea de citire trebuie propagată la close")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}
