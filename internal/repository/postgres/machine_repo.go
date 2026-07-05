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
	rows, err := machineRepo.db.Query("SELECT id, name, code, type, brand, model, year, registration_number, fuel_type, status, notes FROM machines WHERE deleted_at IS NULL")
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
		if err := rows.Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &registrationNumber, &fuelType, &m.Status, &m.Notes); err != nil {
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
	err := machineRepo.db.QueryRow("SELECT id, name, code, type, brand, model, year, registration_number, fuel_type, status, notes FROM machines WHERE id = $1 AND deleted_at IS NULL", id).Scan(&m.ID, &m.Name, &m.Code, &m.Type, &m.Brand, &m.Model, &year, &registrationNumber, &fuelType, &m.Status, &m.Notes)
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

	return machineRepo.db.QueryRow(
		"INSERT INTO machines (name, code, type, brand, model, year, registration_number, fuel_type, status, notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id",
		machine.Name, machine.Code, machine.Type, machine.Brand, machine.Model, machine.Year, registrationNumber, fuelType, machine.Status, machine.Notes,
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

	_, err := machineRepo.db.Exec("UPDATE machines SET name = $1, code = $2, type = $3, brand = $4, model = $5, year = $6, registration_number = $7, fuel_type = $8, status = $9, notes = $10, updated_at = NOW() WHERE id = $11 AND deleted_at IS NULL", machine.Name, machine.Code, machine.Type, machine.Brand, machine.Model, machine.Year, registrationNumber, fuelType, machine.Status, machine.Notes, id)
	return err
}

func (machineRepo *MachineRepo) Delete(id int64) error {
	_, err := machineRepo.db.Exec("UPDATE machines SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

func (machineRepo *MachineRepo) UpdateStatus(tx *sql.Tx, id int64, status domain.MachineStatus) error {
	query := `UPDATE machines SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	_, err := tx.Exec(query, status, id)
	return err
}

// DB returnează conexiunea la baza de date pentru inițierea tranzacțiilor
func (machineRepo *MachineRepo) DB() *sql.DB {
	return machineRepo.db
}
