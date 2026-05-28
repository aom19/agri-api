package postgres

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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
			machine_id, operator_id, start_date, end_date, status
		)
		VALUES ($1,$2,$3,$4,$5)
		RETURNING id
	`

	return tx.QueryRow(query,
		assigment.MachineID,
		assigment.OperatorID,
		assigment.StartDate,
		assigment.EndDate,
		assigment.Status,
	).Scan(&assigment.ID)
}

func (repo *AssignmentRepo) GetAll() ([]domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE deleted_at IS NULL`
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

func (repo *AssignmentRepo) GetAllWithPagination(query dto.PaginationQuery) (*dto.PaginatedAssignmentsResponse, error) {
	// offset - calculat pe baza paginii și limită pentru a sări peste înregistrările anterioare

	baseQuery := `SELECT a.id, a.machine_id, m.name, a.operator_id, o.name, a.start_date, a.end_date, a.status FROM assignments a JOIN machines m ON a.machine_id = m.id JOIN operators o ON a.operator_id = o.id WHERE 1=1 AND a.deleted_at IS NULL`
	countQuery := `SELECT COUNT(*) FROM assignments a WHERE 1=1 AND a.deleted_at IS NULL`
	var filterArgs []interface{}
	argIndex := 1

	if query.Status != "" {
		cond := fmt.Sprintf(` AND a.status = $%d`, argIndex)
		baseQuery += cond
		countQuery += cond
		filterArgs = append(filterArgs, query.Status)
		argIndex++
	}
	if query.OperatorID != "" {
		cond := fmt.Sprintf(` AND a.operator_id = $%d`, argIndex)
		baseQuery += cond
		countQuery += cond
		filterArgs = append(filterArgs, query.OperatorID)
		argIndex++
	}
	if query.MachineID != "" {
		cond := fmt.Sprintf(` AND a.machine_id = $%d`, argIndex)
		baseQuery += cond
		countQuery += cond
		filterArgs = append(filterArgs, query.MachineID)
		argIndex++
	}
	// Validare pentru sortare - doar anumite câmpuri sunt permise pentru sortare
	allowedSortFields := map[string]string{
		"id":            "a.id",
		"start_date":    "a.start_date",
		"end_date":      "a.end_date",
		"status":        "a.status",
		"machine_name":  "m.name",
		"operator_name": "o.name",
	}

	sortField, ok := allowedSortFields[query.SortBy]
	// Dacă câmpul de sortare nu este valid, se folosește un câmp implicit pentru sortare (de exemplu, data de început)
	if !ok {
		sortField = "a.start_date"
	}
	// Validare pentru ordinea de sortare - doar "ASC" sau "DESC" sunt permise
	order := "ASC"
	// Dacă ordinea de sortare este specificată și este "DESC", se setează ordinea la "DESC"
	if strings.ToUpper(query.Order) == "DESC" {
		order = "DESC"
	}

	offset := (query.Page - 1) * query.Limit
	baseQuery += fmt.Sprintf(` ORDER BY %s %s LIMIT $%d OFFSET $%d`, sortField, order, argIndex, argIndex+1)
	pageArgs := append(filterArgs, query.Limit, offset)

	rows, err := repo.db.Query(baseQuery, pageArgs...)
	if err != nil {
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	var assigments []dto.AssigmentResponse
	for rows.Next() {
		var a dto.AssigmentResponse
		if err := rows.Scan(&a.ID, &a.MachineID, &a.MachineName, &a.OperatorID, &a.OperatorName, &a.StartDate, &a.EndDate, &a.Status); err != nil {
			return nil, err
		}
		assigments = append(assigments, a)
	}

	var total int
	err = repo.db.QueryRow(countQuery, filterArgs...).Scan(&total)
	if err != nil {
		return nil, err
	}

	response := &dto.PaginatedAssignmentsResponse{
		Data: assigments,
		Meta: dto.PaginationMeta{
			Page:       query.Page,
			Limit:      query.Limit,
			Total:      total,
			TotalPages: (total + query.Limit - 1) / query.Limit,
		},
	}

	return response, nil
}

func (repo *AssignmentRepo) GetByID(id int64) (*domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE id = $1 AND deleted_at IS NULL`
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
	query := `UPDATE assignments SET machine_id = $1, operator_id = $2, start_date = $3, end_date = $4, status = $5, updated_at = NOW() WHERE id = $6 AND deleted_at IS NULL`
	_, err := repo.db.Exec(query, assigment.MachineID, assigment.OperatorID, assigment.StartDate, assigment.EndDate, assigment.Status, id)
	return err
}

func (repo *AssignmentRepo) Delete(id int64) error {
	query := `UPDATE assignments SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	_, err := repo.db.Exec(query, id)
	return err
}

func (repo *AssignmentRepo) GetActiveByMachine(machineID int64) (*domain.Assigment, error) {
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE machine_id = $1 AND status = 'active' AND deleted_at IS NULL`
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
	query := `SELECT id, machine_id, operator_id, start_date, end_date, status FROM assignments WHERE operator_id = $1 AND status = 'active' AND deleted_at IS NULL`
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
