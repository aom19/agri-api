package postgres

import (
	"agri-api/internal/domain"
	"database/sql"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role FROM users WHERE email = $1 AND deleted_at IS NULL`
	var u domain.User
	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	query := `SELECT id, email, password_hash, role FROM users WHERE id = $1 AND deleted_at IS NULL`
	var u domain.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(user *domain.User) error {
	query := `INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3) RETURNING id`
	return r.db.QueryRow(query, user.Email, user.PasswordHash, user.Role).Scan(&user.ID)
}

func (r *UserRepo) UpdatePassword(id int64, passwordHash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, id)
	return err
}
