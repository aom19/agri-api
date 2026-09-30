package usecase

import (
	"errors"
	"testing"

	"agri-api/internal/domain"
)

func validMachine() *domain.Machine {
	return &domain.Machine{Name: "Tractor 1", Code: "TR-1", Type: domain.MachineTypeTractor, Year: ptr(2020)}
}

func TestMachineService_CreateMachine(t *testing.T) {
	repo := &machineRepoMock{create: func(m *domain.Machine) error { m.ID = 7; return nil }}
	svc := NewMachineService(repo)

	created, err := svc.CreateMachine(validMachine())
	if err != nil {
		t.Fatalf("eroare neașteptată: %v", err)
	}
	if created.ID != 7 || created.Status != domain.MachineStatusActive {
		t.Errorf("mașina creată trebuie să aibă ID-ul din repo și statusul activ, am primit %+v", created)
	}

	invalid := map[string]*domain.Machine{
		"câmpuri lipsă":       {Name: "x"},
		"tip invalid":         {Name: "x", Code: "c", Type: "barca"},
		"combustibil invalid": {Name: "x", Code: "c", Type: domain.MachineTypeCar, FuelType: ptr(domain.FuelType("carbune"))},
		"an invalid":          {Name: "x", Code: "c", Type: domain.MachineTypeCar, Year: ptr(1800)},
	}
	for name, in := range invalid {
		if _, err := svc.CreateMachine(in); err == nil {
			t.Errorf("%s: mă așteptam la eroare de validare", name)
		}
	}

	repo.create = func(*domain.Machine) error { return errors.New("db down") }
	if _, err := svc.CreateMachine(validMachine()); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
}

func TestMachineService_UpdateMachine(t *testing.T) {
	existing := &domain.Machine{ID: 1, Name: "Vechi", Code: "OLD", Type: domain.MachineTypeCombine, Status: domain.MachineStatusActive}
	repo := &machineRepoMock{getByID: func(id int64) (*domain.Machine, error) {
		if id == 1 {
			return existing, nil
		}
		return nil, nil
	}}
	svc := NewMachineService(repo)

	if _, err := svc.UpdateMachine(99, validMachine()); err == nil {
		t.Error("mașina inexistentă trebuie să dea eroare")
	}

	input := validMachine()
	input.Status = domain.MachineStatusMaintenance
	updated, err := svc.UpdateMachine(1, input)
	if err != nil {
		t.Fatalf("eroare neașteptată: %v", err)
	}
	if updated.Name != "Tractor 1" || updated.Status != domain.MachineStatusMaintenance || updated.ID != 1 {
		t.Errorf("câmpurile nu au fost actualizate: %+v", updated)
	}

	invalid := map[string]*domain.Machine{
		"câmpuri lipsă":  {Name: "x", Status: domain.MachineStatusActive},
		"tip invalid":    {Name: "x", Code: "c", Type: "barca", Status: domain.MachineStatusActive},
		"status invalid": {Name: "x", Code: "c", Type: domain.MachineTypeCar, Status: "pierdut"},
		"an invalid":     {Name: "x", Code: "c", Type: domain.MachineTypeCar, Status: domain.MachineStatusActive, Year: ptr(2200)},
		"combustibil":    {Name: "x", Code: "c", Type: domain.MachineTypeCar, Status: domain.MachineStatusActive, FuelType: ptr(domain.FuelType("apa"))},
	}
	for name, in := range invalid {
		if _, err := svc.UpdateMachine(1, in); err == nil {
			t.Errorf("%s: mă așteptam la eroare de validare", name)
		}
	}

	repo.update = func(int64, *domain.Machine) error { return errors.New("db down") }
	if _, err := svc.UpdateMachine(1, input); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
	repo.getByID = func(int64) (*domain.Machine, error) { return nil, errors.New("db down") }
	if _, err := svc.UpdateMachine(1, input); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}

func TestMachineService_GetAndDelete(t *testing.T) {
	repo := &machineRepoMock{
		getAll:  func() ([]domain.Machine, error) { return []domain.Machine{{ID: 1}, {ID: 2}}, nil },
		getByID: func(id int64) (*domain.Machine, error) { return &domain.Machine{ID: id}, nil },
	}
	svc := NewMachineService(repo)

	all, err := svc.GetMachines()
	if err != nil || len(all) != 2 {
		t.Fatalf("GetMachines: %v, %d elemente", err, len(all))
	}
	m, err := svc.GetMachineByID(5)
	if err != nil || m.ID != 5 {
		t.Fatalf("GetMachineByID: %v, %+v", err, m)
	}
	if err := svc.DeleteMachine(5); err != nil {
		t.Fatalf("DeleteMachine: %v", err)
	}

	repo.getByID = func(int64) (*domain.Machine, error) { return nil, errors.New("db down") }
	if _, err := svc.GetMachineByID(5); err == nil {
		t.Error("eroarea din repo trebuie propagată")
	}
}
