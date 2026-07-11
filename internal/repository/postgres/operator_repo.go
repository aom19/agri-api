package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

// OperatorRepo implementează repository.OperatorRepository folosind PostgreSQL
type OperatorRepo struct {
	db *sql.DB
}

// NewOperatorRepo creează o nouă instanță a repository-ului pentru operatori
func NewOperatorRepo(db *sql.DB) *OperatorRepo {
	return &OperatorRepo{db: db}
}

// GetAll returnează toți operatorii din baza de date
func (operatorRepo *OperatorRepo) GetAll() ([]domain.Operator, error) {
	rows, err := operatorRepo.db.Query("SELECT id, name, phone, email, notes, status FROM operators WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var operators []domain.Operator
	for rows.Next() {
		var o domain.Operator
		var phone, email, notes sql.NullString
		if err := rows.Scan(&o.ID, &o.Name, &phone, &email, &notes, &o.Status); err != nil {
			return nil, err
		}
		if phone.Valid {
			o.Phone = phone.String
		}
		if email.Valid {
			o.Email = email.String
		}
		if notes.Valid {
			o.Notes = notes.String
		}
		operators = append(operators, o)
	}

	return operators, nil
}

// GetByID returnează un operator după ID; returnează nil, nil dacă nu există
func (operatorRepo *OperatorRepo) GetByID(id int64) (*domain.Operator, error) {
	var o domain.Operator
	var phone, email, notes sql.NullString
	err := operatorRepo.db.QueryRow(
		"SELECT id, name, phone, email, notes, status FROM operators WHERE id = $1 AND deleted_at IS NULL", id,
	).Scan(&o.ID, &o.Name, &phone, &email, &notes, &o.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if phone.Valid {
		o.Phone = phone.String
	}
	if email.Valid {
		o.Email = email.String
	}
	if notes.Valid {
		o.Notes = notes.String
	}
	return &o, nil
}

// Create inserează un operator nou și populează câmpul ID cu valoarea generată de DB
func (operatorRepo *OperatorRepo) Create(operator *domain.Operator) error {
	var phone, email, notes any
	if operator.Phone != "" {
		phone = operator.Phone
	}
	if operator.Email != "" {
		email = operator.Email
	}
	if operator.Notes != "" {
		notes = operator.Notes
	}
	return operatorRepo.db.QueryRow(
		"INSERT INTO operators (name, phone, email, notes, status) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		operator.Name, phone, email, notes, operator.Status,
	).Scan(&operator.ID)
}

// Update modifică datele unui operator existent identificat prin ID
func (operatorRepo *OperatorRepo) Update(id int64, operator *domain.Operator) error {
	var phone, email, notes any
	if operator.Phone != "" {
		phone = operator.Phone
	}
	if operator.Email != "" {
		email = operator.Email
	}
	if operator.Notes != "" {
		notes = operator.Notes
	}
	_, err := operatorRepo.db.Exec(
		"UPDATE operators SET name = $1, phone = $2, email = $3, notes = $4, updated_at = NOW() WHERE id = $5 AND deleted_at IS NULL",
		operator.Name, phone, email, notes, id,
	)
	return err
}

func (operatorRepo *OperatorRepo) Delete(id int64) error {
	_, err := operatorRepo.db.Exec("UPDATE operators SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

func (operatorRepo *OperatorRepo) UpdateStatus(tx *sql.Tx, id int64, status domain.OperatorStatus) error {
	_, err := tx.Exec("UPDATE operators SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL", status, id)
	return err
}

// UpdateStatusDirect actualizează statusul unui operator fără tranzacție
func (operatorRepo *OperatorRepo) UpdateStatusDirect(id int64, status domain.OperatorStatus) error {
	_, err := operatorRepo.db.Exec("UPDATE operators SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL", status, id)
	return err
}
