package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
	"errors"
)

type AssignmentRepo struct {
	db *sql.DB
}

func NewAssignmentRepo(db *sql.DB) *AssignmentRepo {
	return &AssignmentRepo{
		db: db,
	}
}

func (repo *AssignmentRepo) Create(tx *sql.Tx, assigment *domain.Assigment) error {
	query := `
		INSERT INTO assignments (
			id, machine_id, operator_id, start_date, end_date, status
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`

	_, err := tx.Exec(query,
		assigment.ID,
		assigment.MachineID,
		assigment.OperatorID,
		assigment.StartDate,
		assigment.EndDate,
		assigment.Status,
	)

	return err
}

func (repo *AssignmentRepo) GetAll() ([]domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments`
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var assigments []domain.Assigment
	for rows.Next() {
		var a domain.Assigment
		if err := rows.Scan(&a.ID, &a.MachineID, &a.OperatorID, &a.StartDate, &a.EndDate, &a.Status); err != nil {
			return nil, err
		}
		assigments = append(assigments, a)
	}

	return assigments, nil
}

func (repo *AssignmentRepo) GetByID(id int64) (*domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE id = $1`
	row := repo.db.QueryRow(query, id)

	var a domain.Assigment
	if err := row.Scan(&a.ID, &a.MachineID, &a.OperatorID, &a.StartDate, &a.EndDate, &a.Status); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Nu s-a găsit nicio înregistrare
		}
		return nil, err
	}

	return &a, nil
}

func (repo *AssignmentRepo) Update(id int64, assigment *domain.Assigment) error {
	query := `UPDATE assignments SET machine_id = $1, operator_id = $2, start_date = $3, end_date = $4, status = $5 WHERE id = $6`
	_, err := repo.db.Exec(query, assigment.MachineID, assigment.OperatorID, assigment.StartDate, assigment.EndDate, assigment.Status, id)
	return err
}

func (repo *AssignmentRepo) Delete(id int64) error {
	query := `DELETE FROM assignments WHERE id = $1`
	_, err := repo.db.Exec(query, id)
	return err
}

func (repo *AssignmentRepo) GetActiveByMachine(machineID int64) (*domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE machine_id = $1 AND status = 'active'`
	var a domain.Assigment
	err := repo.db.QueryRow(query, machineID).Scan(&a.ID, &a.MachineID, &a.OperatorID, &a.StartDate, &a.EndDate, &a.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &a, nil
}

func (repo *AssignmentRepo) GetActiveByOperator(operatorID int64) (*domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE operator_id = $1 AND status = 'active'`
	var a domain.Assigment

	err := repo.db.QueryRow(query, operatorID).Scan(
		&a.ID,
		&a.MachineID,
		&a.OperatorID,
		&a.StartDate,
		&a.EndDate,
		&a.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil // Nu s-a găsit nicio înregistrare
	} else if err != nil {
		return nil, err
	}

	return &a, nil
}

func (repo *AssignmentRepo) IsAssigmentActive(id int64) (bool, error) {
	assigment, err := repo.GetByID(id)
	if err != nil {
		return false, err
	}
	if assigment == nil {
		return false, errors.New("assigment not found")
	}
	return assigment.Status == domain.AssigmentStatusActive, nil
}
