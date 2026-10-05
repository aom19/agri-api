package postgres

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/lib/pq"
)

// OperatorRepo implementează repository.OperatorRepository peste users + user_profiles.
type OperatorRepo struct {
	db *sql.DB
}

func NewOperatorRepo(db *sql.DB) *OperatorRepo {
	return &OperatorRepo{db: db}
}

// Operatorul unei lucrări (alias `fo`) este un utilizator: ou = users, oup = user_profiles.
const operatorJoin = `
	LEFT JOIN users ou          ON ou.id = fo.operator_id
	LEFT JOIN user_profiles oup ON oup.user_id = ou.id`

// operatorNameExpr este numele operatorului (alias-uri ou/oup): prenume + nume din profil sau,
// în lipsă, e-mailul contului.
const operatorNameExpr = `COALESCE(NULLIF(BTRIM(CONCAT_WS(' ', oup.first_name, oup.last_name)), ''), ou.email)`

// operatorUsersFrom sunt conturile cu rolul operator (alias u).
const operatorUsersFrom = `users u JOIN roles r ON r.id = u.role_id AND r.code = 'operator'`

// unusablePasswordHash nu e un hash bcrypt valid, deci contul nu se poate autentifica până
// când operatorul își setează parola (resetare prin e-mail).
const unusablePasswordHash = "!"

const operatorSelect = `
	SELECT
		u.id,
		COALESCE(up.first_name, ''),
		COALESCE(up.last_name, ''),
		COALESCE(up.phone, ''),
		u.email,
		COALESCE(up.notes, ''),
		u.deleted_at IS NULL,
		COALESCE(up.allowed_machine_types, '{}')
	FROM users u
	JOIN roles r ON r.id = u.role_id AND r.code = 'operator'
	LEFT JOIN user_profiles up ON up.user_id = u.id`

func scanOperator(row interface{ Scan(...interface{}) error }) (*domain.Operator, error) {
	var (
		o            domain.Operator
		active       bool
		allowedTypes pq.StringArray
	)
	if err := row.Scan(&o.ID, &o.FirstName, &o.LastName, &o.Phone, &o.Email, &o.Notes, &active, &allowedTypes); err != nil {
		return nil, err
	}
	o.Name = o.FullName()
	if o.Name == "" {
		o.Name = o.Email
	}
	if domain.IsPlaceholderEmail(o.Email) {
		o.Email = ""
	}
	o.Status = domain.OperatorStatusInactive
	if active {
		o.Status = domain.OperatorStatusActive
	}
	o.AllowedMachineTypes = make([]domain.MachineType, 0, len(allowedTypes))
	for _, t := range allowedTypes {
		o.AllowedMachineTypes = append(o.AllowedMachineTypes, domain.MachineType(t))
	}
	return &o, nil
}

// GetAll returnează toți operatorii, inclusiv cei dezactivați, ordonați după nume.
func (operatorRepo *OperatorRepo) GetAll() ([]domain.Operator, error) {
	rows, err := operatorRepo.db.Query(operatorSelect + " ORDER BY up.first_name, up.last_name, u.id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	operators := []domain.Operator{}
	for rows.Next() {
		o, err := scanOperator(rows)
		if err != nil {
			return nil, err
		}
		operators = append(operators, *o)
	}
	return operators, rows.Err()
}

// GetByID returnează operatorul după id-ul contului; nil, nil dacă nu există sau nu e operator.
func (operatorRepo *OperatorRepo) GetByID(id int64) (*domain.Operator, error) {
	o, err := scanOperator(operatorRepo.db.QueryRow(operatorSelect+" WHERE u.id = $1", id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return o, err
}

func (operatorRepo *OperatorRepo) Create(operator *domain.Operator) error {
	tx, err := operatorRepo.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	email := strings.TrimSpace(operator.Email)
	if email == "" {
		if email, err = placeholderEmail(); err != nil {
			return err
		}
	}
	err = tx.QueryRow(`
		INSERT INTO users (email, password_hash, role_id, email_confirmed)
		VALUES ($1, $2, (SELECT id FROM roles WHERE code = 'operator'), FALSE)
		RETURNING id`, email, unusablePasswordHash,
	).Scan(&operator.ID)
	if err != nil {
		return mapEmailTaken(err)
	}
	if err := upsertOperatorProfile(tx, operator.ID, operator); err != nil {
		return err
	}
	return tx.Commit()
}

func (operatorRepo *OperatorRepo) Update(id int64, operator *domain.Operator) error {
	tx, err := operatorRepo.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Fără e-mail, contul își păstrează adresa tehnică sau primește una nouă.
	email := strings.TrimSpace(operator.Email)
	if email == "" {
		var current string
		if err := tx.QueryRow(`SELECT email FROM users WHERE id = $1`, id).Scan(&current); err != nil {
			return err
		}
		email = current
		if !domain.IsPlaceholderEmail(current) {
			if email, err = placeholderEmail(); err != nil {
				return err
			}
		}
	}
	if _, err := tx.Exec(`UPDATE users SET email = $1, updated_at = NOW() WHERE id = $2`, email, id); err != nil {
		return mapEmailTaken(err)
	}
	if err := upsertOperatorProfile(tx, id, operator); err != nil {
		return err
	}
	return tx.Commit()
}

// SetActive dezactivează contul (deleted_at, ca în pagina Utilizatori) sau îl reactivează.
func (operatorRepo *OperatorRepo) SetActive(id int64, active bool) error {
	query := `UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	if active {
		query = `UPDATE users SET deleted_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NOT NULL`
	}
	_, err := operatorRepo.db.Exec(query, id)
	return err
}

func upsertOperatorProfile(tx *sql.Tx, userID int64, operator *domain.Operator) error {
	allowedTypes := make([]string, len(operator.AllowedMachineTypes))
	for i, t := range operator.AllowedMachineTypes {
		allowedTypes[i] = string(t)
	}
	_, err := tx.Exec(`
		INSERT INTO user_profiles (user_id, first_name, last_name, phone, notes, allowed_machine_types)
		VALUES ($1, $2, NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6)
		ON CONFLICT (user_id) DO UPDATE SET
			first_name            = EXCLUDED.first_name,
			last_name             = EXCLUDED.last_name,
			phone                 = EXCLUDED.phone,
			notes                 = EXCLUDED.notes,
			allowed_machine_types = EXCLUDED.allowed_machine_types,
			updated_at            = NOW()`,
		userID,
		strings.TrimSpace(operator.FirstName),
		strings.TrimSpace(operator.LastName),
		strings.TrimSpace(operator.Phone),
		strings.TrimSpace(operator.Notes),
		pq.Array(allowedTypes),
	)
	return err
}

// placeholderEmail generează o adresă tehnică unică pentru un operator fără e-mail.
func placeholderEmail() (string, error) {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "operator-" + hex.EncodeToString(buf) + "@" + domain.PlaceholderEmailDomain, nil
}

func mapEmailTaken(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return repository.ErrEmailTaken
	}
	return err
}
