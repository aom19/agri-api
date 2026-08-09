package domain

type MachineType string
type FuelType string

const (
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
	ID                 int64       `json:"id"`
	Name               string      `json:"name"`
	Code               string      `json:"code"`
	Type               MachineType `json:"type"`
	Brand              string      `json:"brand,omitempty"`
	Model              string      `json:"model,omitempty"`
	Year               *int        `json:"year,omitempty"`
	RegistrationNumber string      `json:"registration_number,omitempty"`
	FuelType           *FuelType   `json:"fuel_type,omitempty"`
	OperatingHours     *float64    `json:"operating_hours,omitempty"`
	Status             AssetStatus `json:"status"`
	Notes              string      `json:"notes,omitempty"`
	WorkingHours       *float64    `json:"working_hours,omitempty"`
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

func (machineType MachineType) IsValid() bool {
	_, ok := validMachineTypes[machineType]
	return ok
}

func (fuelType FuelType) IsValid() bool {
	_, ok := validFuelTypes[fuelType]
	return ok
}
