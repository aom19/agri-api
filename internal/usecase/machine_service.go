package usecase

import (
	"errors"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

// MachineService conține logica de business pentru gestionarea mașinilor agricole
type MachineService struct {
	machineRepo repository.MachineRepository
}

// NewMachineService creează o nouă instanță a serviciului cu repository-ul injectat
func NewMachineService(machineRepo repository.MachineRepository) *MachineService {
	return &MachineService{machineRepo: machineRepo}
}

// GetMachines returnează lista completă a mașinilor
func (machineService *MachineService) GetMachines() ([]domain.Machine, error) {
	return machineService.machineRepo.GetAll()
}

// GetMachineByID returnează o mașină după ID sau eroare dacă nu există
func (machineService *MachineService) GetMachineByID(id int64) (*domain.Machine, error) {
	machine, err := machineService.machineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return machine, nil
}

// CreateMachine validează și creează o mașină nouă cu statusul implicit "available"
func (machineService *MachineService) CreateMachine(machine *domain.Machine) (*domain.Machine, error) {
	if machine.Name == "" || machine.Type == "" {
		return nil, errors.New("name and type are required")
	}
	m := &domain.Machine{
		Name:        machine.Name,
		Type:        machine.Type,
		Status:      domain.MachineStatusAvailable,
		Description: machine.Description,
	}
	if err := machineService.machineRepo.Create(m); err != nil {
		return nil, err
	}
	return m, nil
}

// UpdateMachine actualizează câmpurile unei mașini existente
func (machineService *MachineService) UpdateMachine(id int64, machine *domain.Machine) (*domain.Machine, error) {
	existing, err := machineService.machineRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("machine not found")
	}
	existing.Name = machine.Name
	existing.Type = machine.Type
	existing.Status = machine.Status
	existing.Description = machine.Description

	if err := machineService.machineRepo.Update(id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (machineService *MachineService) DeleteMachine(id int64) error {
	return machineService.machineRepo.Delete(id)
}
