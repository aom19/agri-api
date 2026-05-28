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
	rows, err := operatorRepo.db.Query("SELECT id, name, status FROM operators 	WHERE deleted_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var operators []domain.Operator
	for rows.Next() {
		var o domain.Operator
		if err := rows.Scan(&o.ID, &o.Name, &o.Status); err != nil {
			return nil, err
		}
		operators = append(operators, o)
	}

	return operators, nil
}

// GetByID returnează un operator după ID; returnează nil, nil dacă nu există
func (operatorRepo *OperatorRepo) GetByID(id int64) (*domain.Operator, error) {
	var o domain.Operator
	err := operatorRepo.db.QueryRow("SELECT id, name, status FROM operators WHERE id = $1 AND deleted_at IS NULL", id).Scan(&o.ID, &o.Name, &o.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// Create inserează un operator nou și populează câmpul ID cu valoarea generată de DB
func (operatorRepo *OperatorRepo) Create(operator *domain.Operator) error {
	return operatorRepo.db.QueryRow(
		"INSERT INTO operators (name, status) VALUES ($1, $2) RETURNING id",
		operator.Name, operator.Status,
	).Scan(&operator.ID)
}

// Update modifică datele unui operator existent identificat prin ID
func (operatorRepo *OperatorRepo) Update(id int64, operator *domain.Operator) error {
	_, err := operatorRepo.db.Exec("UPDATE operators SET name = $1, status = $2 WHERE id = $3 AND deleted_at IS NULL", operator.Name, operator.Status, id)
	return err
}

func (operatorRepo *OperatorRepo) Delete(id int64) error {
	_, err := operatorRepo.db.Exec("UPDATE operators SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL", id)
	return err
}

func (operatorRepo *OperatorRepo) UpdateStatus(tx *sql.Tx, id int64, status domain.OperatorStatus) error {
	_, err := tx.Exec("UPDATE operators SET status = $1 WHERE id = $2 AND deleted_at IS NULL", status, id)
	return err
}
