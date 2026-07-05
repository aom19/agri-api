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
	if machine.Name == "" || machine.Code == "" || machine.Type == "" {
		return nil, errors.New("name, code and type are required")
	}
	if !machine.Type.IsValid() {
		return nil, errors.New("invalid machine type")
	}
	if machine.FuelType != nil && !machine.FuelType.IsValid() {
		return nil, errors.New("invalid fuel type")
	}
	if machine.Year != nil && (*machine.Year < 1900 || *machine.Year > 2100) {
		return nil, errors.New("invalid year")
	}
	m := &domain.Machine{
		Name:               machine.Name,
		Code:               machine.Code,
		Type:               machine.Type,
		Brand:              machine.Brand,
		Model:              machine.Model,
		Year:               machine.Year,
		RegistrationNumber: machine.RegistrationNumber,
		FuelType:           machine.FuelType,
		Status:             domain.MachineStatusActive,
		Notes:              machine.Notes,
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
	if machine.Name == "" || machine.Code == "" || machine.Type == "" {
		return nil, errors.New("name, code and type are required")
	}
	if !machine.Type.IsValid() {
		return nil, errors.New("invalid machine type")
	}
	if machine.FuelType != nil && !machine.FuelType.IsValid() {
		return nil, errors.New("invalid fuel type")
	}
	if !machine.Status.IsValid() {
		return nil, errors.New("invalid machine status")
	}
	if machine.Year != nil && (*machine.Year < 1900 || *machine.Year > 2100) {
		return nil, errors.New("invalid year")
	}

	existing.Name = machine.Name
	existing.Code = machine.Code
	existing.Type = machine.Type
	existing.Brand = machine.Brand
	existing.Model = machine.Model
	existing.Year = machine.Year
	existing.RegistrationNumber = machine.RegistrationNumber
	existing.FuelType = machine.FuelType
	existing.Status = machine.Status
	existing.Notes = machine.Notes

	if err := machineService.machineRepo.Update(id, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (machineService *MachineService) DeleteMachine(id int64) error {
	return machineService.machineRepo.Delete(id)
}
