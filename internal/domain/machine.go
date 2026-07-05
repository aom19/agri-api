package domain

type MachineStatus string
type MachineType string
type FuelType string

const (
	MachineStatusActive      MachineStatus = "active"
	MachineStatusMaintenance MachineStatus = "maintenance"
	MachineStatusInactive    MachineStatus = "inactive"

	MachineTypeTractor    MachineType = "tractor"
	MachineTypeCombine    MachineType = "combine"
	MachineTypeDrone      MachineType = "drone"
	MachineTypeSprayer    MachineType = "sprayer"
	MachineTypeCar        MachineType = "car"
	MachineTypeSmallTruck MachineType = "small_truck"
	MachineTypeOther      MachineType = "other"

	FuelTypeDiesel   FuelType = "diesel"
	FuelTypeGasoline FuelType = "gasoline"
	FuelTypeElectric FuelType = "electric"
	FuelTypeHybrid   FuelType = "hybrid"
)

type Machine struct {
	ID                 int64         `json:"id"`
	Name               string        `json:"name"`
	Code               string        `json:"code"`
	Type               MachineType   `json:"type"`
	Brand              string        `json:"brand,omitempty"`
	Model              string        `json:"model,omitempty"`
	Year               *int          `json:"year,omitempty"`
	RegistrationNumber string        `json:"registration_number,omitempty"`
	FuelType           *FuelType     `json:"fuel_type,omitempty"`
	Status             MachineStatus `json:"status"`
	Notes              string        `json:"notes,omitempty"`

	Auditfields
}

var validMachineTypes = map[MachineType]struct{}{
	MachineTypeTractor:    {},
	MachineTypeCombine:    {},
	MachineTypeDrone:      {},
	MachineTypeSprayer:    {},
	MachineTypeCar:        {},
	MachineTypeSmallTruck: {},
	MachineTypeOther:      {},
}

var validFuelTypes = map[FuelType]struct{}{
	FuelTypeDiesel:   {},
	FuelTypeGasoline: {},
	FuelTypeElectric: {},
	FuelTypeHybrid:   {},
}

var validMachineStatuses = map[MachineStatus]struct{}{
	MachineStatusActive:      {},
	MachineStatusMaintenance: {},
	MachineStatusInactive:    {},
}

func (machineType MachineType) IsValid() bool {
	_, ok := validMachineTypes[machineType]
	return ok
}

func (fuelType FuelType) IsValid() bool {
	_, ok := validFuelTypes[fuelType]
	return ok
}

func (machineStatus MachineStatus) IsValid() bool {
	_, ok := validMachineStatuses[machineStatus]
	return ok
}
