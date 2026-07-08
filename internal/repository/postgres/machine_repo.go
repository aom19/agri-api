package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

// MachineRepo implementează repository.MachineRepository folosind PostgreSQL
type MachineRepo struct {
	db *sql.DB
}

// NewMachineRepo creează o nouă instanță a repository-ului pentru mașini
func NewMachineRepo(db *sql.DB) *MachineRepo {
	return &MachineRepo{db: db}
}

// GetAll returnează toate mașinile din baza de date
func (machineRepo *MachineRepo) GetAll() ([]domain.Machine, error) {
	rows, err := machineRepo.db.Query("SELECT id, name, code, type, brand, model, year, registration_number, fuel_type, asset_status, working_hours, notes FROM machines WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var machines []domain.Machine
	for rows.Next() {
		var m domain.Machine
		var year sql.NullInt64
		var registrationNumber sql.NullString
		var fuelType sql.NullString
		var workingHours sql.NullFloat64
		if err := rows.Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &registrationNumber, &fuelType, &m.Status, &workingHours, &m.Notes); err != nil {
			return nil, err
		}
		if year.Valid {
			y := int(year.Int64)
			m.Year = &y
		}
		if registrationNumber.Valid {
			m.RegistrationNumber = registrationNumber.String
		}
		if fuelType.Valid {
			ft := domain.FuelType(fuelType.String)
			m.FuelType = &ft
		}
		if workingHours.Valid {
			w := float64(workingHours.Float64)
			m.WorkingHours = &w
		}
		machines = append(machines, m)
	}

	return machines, nil
}

// GetByID returnează o mașină după ID; returnează nil, nil dacă nu există
func (machineRepo *MachineRepo) GetByID(id int64) (*domain.Machine, error) {
	var m domain.Machine
	var year sql.NullInt64
	var registrationNumber sql.NullString
	var fuelType sql.NullString
	var workingHours sql.NullFloat64
	err := machineRepo.db.QueryRow("SELECT id, name, code, type, brand, model, year, registration_number, fuel_type, asset_status, working_hours, notes FROM machines WHERE id = $1 AND deleted_at IS NULL", id).Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &registrationNumber, &fuelType, &m.Status, &workingHours, &m.Notes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if year.Valid {
		y := int(year.Int64)
		m.Year = &y
	}
	if workingHours.Valid {
		w := float64(workingHours.Float64)
		m.WorkingHours = &w
	}
	if registrationNumber.Valid {
		m.RegistrationNumber = registrationNumber.String
	}
	if fuelType.Valid {
		ft := domain.FuelType(fuelType.String)
		m.FuelType = &ft
	}
	return &m, nil
}

// Create inserează o mașină nouă și populează câmpul ID cu valoarea generată de DB
func (machineRepo *MachineRepo) Create(machine *domain.Machine) error {
	var fuelType any
	if machine.FuelType != nil {
		fuelType = string(*machine.FuelType)
	}
	var registrationNumber any
	if machine.RegistrationNumber != "" {
		registrationNumber = machine.RegistrationNumber
	}
	var workingHours any
	if machine.WorkingHours != nil {
		workingHours = *machine.WorkingHours
	}

	return machineRepo.db.QueryRow(
		"INSERT INTO machines (name, code, type, brand, model, year, registration_number, fuel_type, asset_status, working_hours, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id",
		machine.Name, machine.Code, machine.Type, machine.Brand, machine.Model, machine.Year, registrationNumber, fuelType, machine.Status, workingHours, machine.Notes,
	).Scan(&machine.ID)
}

// Update modifică datele unei mașini existente identificate prin ID
func (machineRepo *MachineRepo) Update(id int64, machine *domain.Machine) error {
	var fuelType any
	if machine.FuelType != nil {
		fuelType = string(*machine.FuelType)
	}
	var registrationNumber any
	if machine.RegistrationNumber != "" {
		registrationNumber = machine.RegistrationNumber
	}
	var workingHours any
	if machine.WorkingHours != nil {
		workingHours = *machine.WorkingHours
	}

	_, err := machineRepo.db.Exec("UPDATE machines SET name = $1, code = $2, type = $3, brand = $4, model = $5, year = $6, registration_number = $7, fuel_type = $8, asset_status = $9, working_hours = $10, notes = $11, updated_at = NOW() WHERE id = $12 AND deleted_at IS NULL", machine.Name, machine.Code, machine.Type, machine.Brand, machine.Model, machine.Year, registrationNumber, fuelType, machine.Status, workingHours, machine.Notes, id)
	return err
}

func (machineRepo *MachineRepo) Delete(id int64) error {
	_, err := machineRepo.db.Exec("UPDATE machines SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

func (machineRepo *MachineRepo) UpdateStatus(tx *sql.Tx, id int64, status domain.MachineStatus) error {
	query := `UPDATE machines SET asset_status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := tx.Exec(query, status, id)
	return err
}

// DB returnează conexiunea la baza de date pentru inițierea tranzacțiilor
func (machineRepo *MachineRepo) DB() *sql.DB {
	return machineRepo.db
}
