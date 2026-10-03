package auth_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"

	"agri-api/internal/auth"
)

func newRepo(t *testing.T) (*auth.Repo, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return auth.NewRepo(db), mock
}

func TestRepo_RefreshTokens(t *testing.T) {
	repo, mock := newRepo(t)
	columns := []string{"user_id", "expires_at", "revoked", "access_jti"}
	future := time.Now().Add(time.Hour)

	mock.ExpectExec("INSERT INTO refresh_tokens").WithArgs(int64(1), "rt", "jti", future).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.StoreRefreshToken(1, "rt", "jti", future); err != nil {
		t.Errorf("StoreRefreshToken: %v", err)
	}

	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), future, false, "jti"))
	userID, jti, err := repo.ValidateRefreshToken("rt")
	if err != nil || userID != 1 || jti != "jti" {
		t.Errorf("ValidateRefreshToken: %v, %d, %q", err, userID, jti)
	}
	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), future, true, nil))
	if _, _, err := repo.ValidateRefreshToken("rt"); err == nil {
		t.Error("token revocat trebuie respins")
	}
	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), time.Now().Add(-time.Hour), false, "jti"))
	if _, _, err := repo.ValidateRefreshToken("rt"); err == nil {
		t.Error("token expirat trebuie respins")
	}
	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").WillReturnError(sql.ErrNoRows)
	if _, _, err := repo.ValidateRefreshToken("rt"); err == nil {
		t.Error("token inexistent trebuie respins")
	}

	mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.RevokeRefreshToken("rt"); err != nil {
		t.Errorf("RevokeRefreshToken: %v", err)
	}

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"access_jti"}).AddRow("a").AddRow(nil).AddRow("").AddRow("b"))
	jtis, err := repo.RevokeAllUserSessions(1)
	if err != nil || len(jtis) != 2 || jtis[0] != "a" || jtis[1] != "b" {
		t.Errorf("RevokeAllUserSessions: %v, %v", err, jtis)
	}
	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").WillReturnError(errors.New("db down"))
	if _, err := repo.RevokeAllUserSessions(1); err == nil {
		t.Error("eroarea de revocare trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestRepo_ResetAndConfirmationTokens(t *testing.T) {
	repo, mock := newRepo(t)
	columns := []string{"user_id", "expires_at", "used"}
	future := time.Now().Add(time.Hour)

	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO password_reset_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.StoreResetToken(1, "t", future); err != nil {
		t.Errorf("StoreResetToken: %v", err)
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), future, false))
	if userID, err := repo.ValidateResetToken("t"); err != nil || userID != 1 {
		t.Errorf("ValidateResetToken: %v, %d", err, userID)
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), future, true))
	if _, err := repo.ValidateResetToken("t"); err == nil {
		t.Error("token folosit trebuie respins")
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").WillReturnError(sql.ErrNoRows)
	if _, err := repo.ValidateResetToken("t"); err == nil {
		t.Error("token inexistent trebuie respins")
	}
	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.InvalidateResetToken("t"); err != nil {
		t.Errorf("InvalidateResetToken: %v", err)
	}

	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_confirmation_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.StoreEmailConfirmationToken(1, "c", future); err != nil {
		t.Errorf("StoreEmailConfirmationToken: %v", err)
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), future, false))
	if userID, err := repo.ValidateEmailConfirmationToken("c"); err != nil || userID != 1 {
		t.Errorf("ValidateEmailConfirmationToken: %v, %d", err, userID)
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), time.Now().Add(-time.Hour), false))
	if _, err := repo.ValidateEmailConfirmationToken("c"); err == nil {
		t.Error("token expirat trebuie respins")
	}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").WillReturnError(sql.ErrNoRows)
	if _, err := repo.ValidateEmailConfirmationToken("c"); err == nil {
		t.Error("token inexistent trebuie respins")
	}
	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.InvalidateEmailConfirmationToken("c"); err != nil {
		t.Errorf("InvalidateEmailConfirmationToken: %v", err)
	}
	mock.ExpectQuery("SELECT user_id FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(7)))
	if userID, err := repo.GetEmailConfirmationTokenUserID("c"); err != nil || userID != 7 {
		t.Errorf("GetEmailConfirmationTokenUserID: %v, %d", err, userID)
	}
	mock.ExpectQuery("SELECT user_id FROM email_confirmation_tokens").WillReturnError(sql.ErrNoRows)
	if _, err := repo.GetEmailConfirmationTokenUserID("c"); err == nil {
		t.Error("token inexistent trebuie respins")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}
