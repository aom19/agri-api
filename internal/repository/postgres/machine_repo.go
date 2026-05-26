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
	rows, err := machineRepo.db.Query("SELECT id, name, type, status, description FROM machines")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var machines []domain.Machine
	for rows.Next() {
		var m domain.Machine
		if err := rows.Scan(&m.ID, &m.Name, &m.Type, &m.Status, &m.Description); err != nil {
			return nil, err
		}
		machines = append(machines, m)
	}

	return machines, nil
}

// GetByID returnează o mașină după ID; returnează nil, nil dacă nu există
func (machineRepo *MachineRepo) GetByID(id int64) (*domain.Machine, error) {
	var m domain.Machine
	err := machineRepo.db.QueryRow("SELECT id, name, type, status, description FROM machines WHERE id = $1", id).Scan(&m.ID, &m.Name, &m.Type, &m.Status, &m.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create inserează o mașină nouă și populează câmpul ID cu valoarea generată de DB
func (machineRepo *MachineRepo) Create(machine *domain.Machine) error {
	return machineRepo.db.QueryRow(
		"INSERT INTO machines (name, type, status, description) VALUES ($1, $2, $3, $4) RETURNING id",
		machine.Name, machine.Type, machine.Status, machine.Description,
	).Scan(&machine.ID)
}

// Update modifică datele unei mașini existente identificate prin ID
func (machineRepo *MachineRepo) Update(id int64, machine *domain.Machine) error {
	_, err := machineRepo.db.Exec("UPDATE machines SET name = $1, type = $2, status = $3, description = $4 WHERE id = $5", machine.Name, machine.Type, machine.Status, machine.Description, id)
	return err
}

func (machineRepo *MachineRepo) Delete(id int64) error {
	_, err := machineRepo.db.Exec("DELETE FROM machines WHERE id = $1", id)
	return err
}

func (machineRepo *MachineRepo) UpdateStatus(tx *sql.Tx, id int64, status domain.MachineStatus) error {
	query := `UPDATE machines SET status = $1 WHERE id = $2`
	_, err := tx.Exec(query, status, id)
	return err
}

// DB returnează conexiunea la baza de date pentru inițierea tranzacțiilor
func (machineRepo *MachineRepo) DB() *sql.DB {
	return machineRepo.db
}
