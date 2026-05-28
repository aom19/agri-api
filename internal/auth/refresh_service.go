package auth

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func GenerateRefreshToken() string {
	// Implement refresh token generation logic here (e.g., using UUID or random string
	return uuid.NewString()
}

func (r *Repo) StoreRefreshToken(userID int64, refreshToken string, expiresAt time.Time) error {
	// Implement logic to store refresh token in the database
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, userID, refreshToken, expiresAt)
	return err
}

func (r *Repo) ValidateRefreshToken(refreshToken string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	var revoked bool

	err := r.db.QueryRow(`SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token = $1`, refreshToken).Scan(&userID, &expiresAt, &revoked)
	if err != nil {
		return 0, errors.New("invalid or expired refresh token")
	}
	if revoked {
		return 0, errors.New("refresh token has been revoked")
	}
	if time.Now().After(expiresAt) {
		return 0, errors.New("refresh token has expired")
	}
	return userID, nil
}

func (r *Repo) RevokeRefreshToken(refreshToken string) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`, refreshToken)
	return err
}

func (r *Repo) StoreResetToken(userID int64, token string, expiresAt time.Time) error {
	// Invalidează orice token anterior pentru același user
	_, _ = r.db.Exec(`UPDATE password_reset_tokens SET used = TRUE WHERE user_id = $1`, userID)
	query := `INSERT INTO password_reset_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, userID, token, expiresAt)
	return err
}

func (r *Repo) ValidateResetToken(token string) (int64, error) {
	var userID int64
	var expiresAt time.Time
	var used bool

	err := r.db.QueryRow(
		`SELECT user_id, expires_at, used FROM password_reset_tokens WHERE token = $1`,
		token,
	).Scan(&userID, &expiresAt, &used)
	if err != nil {
		return 0, errors.New("invalid reset token")
	}
	if used || time.Now().After(expiresAt) {
		return 0, errors.New("reset token expired or already used")
	}
	return userID, nil
}

func (r *Repo) InvalidateResetToken(token string) error {
	_, err := r.db.Exec(`UPDATE password_reset_tokens SET used = TRUE WHERE token = $1`, token)
	return err
}
