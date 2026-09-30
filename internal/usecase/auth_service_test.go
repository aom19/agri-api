package usecase

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/email"
	"agri-api/internal/store"

	"github.com/DATA-DOG/go-sqlmock"
)

var cachedPasswordHash string

// passwordHash generează o singură dată hash-ul bcrypt (cost 12 e lent) pentru parola „Parola1!”.
func passwordHash(t *testing.T) string {
	t.Helper()
	if cachedPasswordHash == "" {
		hash, err := auth.HashPassword("Parola1!")
		if err != nil {
			t.Fatal(err)
		}
		cachedPasswordHash = hash
	}
	return cachedPasswordHash
}

// deadEmailService indică un server SMTP inaccesibil: trimiterea eșuează imediat.
func deadEmailService() *email.EmailService {
	return email.NewEmailService("127.0.0.1", 1, "", "", "noreply@test.local", nil)
}

func newAuthService(t *testing.T, emailSvc *email.EmailService) (*AuthService, *userRepoMock, *roleRepoMock, sqlmock.Sqlmock) {
	t.Helper()
	db, mock := newSQLMock(t)
	hash := passwordHash(t)
	users := &userRepoMock{
		getByEmail: func(e string) (*domain.User, error) {
			switch e {
			case "ana@x.ro":
				return &domain.User{ID: 1, Email: e, PasswordHash: hash, EmailConfirmed: true, RoleID: 1, RoleCode: "admin", RoleName: "Admin"}, nil
			case "nou@x.ro":
				return &domain.User{ID: 2, Email: e, PasswordHash: hash, EmailConfirmed: false}, nil
			}
			return nil, nil
		},
		getByID: func(id int64) (*domain.User, error) {
			switch id {
			case 1:
				return &domain.User{ID: 1, Email: "ana@x.ro", PasswordHash: hash, EmailConfirmed: true, RoleID: 1}, nil
			case 2:
				return &domain.User{ID: 2, Email: "nou@x.ro", PasswordHash: hash, EmailConfirmed: false}, nil
			}
			return nil, nil
		},
		create: func(u *domain.User) error { u.ID = 3; return nil },
	}
	roles := &roleRepoMock{getByCode: func(code string) (*domain.Role, error) {
		if code == "viewer" {
			return &domain.Role{ID: 2, Code: "viewer", Name: "Vizitator"}, nil
		}
		return nil, nil
	}}
	jwt := auth.NewJWTService("secret", 15*time.Minute, time.Hour)
	svc := NewAuthService(&store.Store{UserRepo: users, RoleRepo: roles}, jwt, auth.NewRepo(db), deadBlacklist(), emailSvc, "http://front/")
	return svc, users, roles, mock
}

func TestNewAuthService_ClientOrigin(t *testing.T) {
	svc := NewAuthService(nil, nil, nil, nil, nil, "")
	if svc.clientOrigin != "http://localhost:3000" {
		t.Errorf("origin implicit: %s", svc.clientOrigin)
	}
	svc = NewAuthService(nil, nil, nil, nil, nil, "http://app/")
	if svc.clientOrigin != "http://app" {
		t.Errorf("origin fără slash final: %s", svc.clientOrigin)
	}
}

func TestAuthService_Login(t *testing.T) {
	svc, users, _, mock := newAuthService(t, nil)

	if _, _, err := svc.Login("necunoscut@x.ro", "x"); err == nil {
		t.Error("email necunoscut trebuie să dea eroare")
	}
	if _, _, err := svc.Login("ana@x.ro", "gresit"); err == nil {
		t.Error("parola greșită trebuie să dea eroare")
	}
	if _, _, err := svc.Login("nou@x.ro", "Parola1!"); !errors.Is(err, ErrEmailNotConfirmed) {
		t.Errorf("email neconfirmat: %v", err)
	}

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"access_jti"}).AddRow("jti-vechi"))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	access, refresh, err := svc.Login("ana@x.ro", "Parola1!")
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("Login: %v", err)
	}
	claims, err := svc.jwtService.Parse(access)
	if err != nil || claims["user_id"] != "1" || claims["role_code"] != "admin" {
		t.Errorf("claims greșite: %v, %v", err, claims)
	}

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").WillReturnRows(sqlmock.NewRows([]string{"access_jti"}))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnError(errors.New("db down"))
	if _, _, err := svc.Login("ana@x.ro", "Parola1!"); err == nil {
		t.Error("eroarea la salvarea refresh token-ului trebuie propagată")
	}

	users.getByEmail = func(string) (*domain.User, error) { return nil, errors.New("db down") }
	if _, _, err := svc.Login("ana@x.ro", "Parola1!"); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAuthService_RefreshAndLogout(t *testing.T) {
	svc, users, _, mock := newAuthService(t, nil)
	refreshColumns := []string{"user_id", "expires_at", "revoked", "access_jti"}
	future := time.Now().Add(time.Hour)

	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").WillReturnError(sql.ErrNoRows)
	if _, _, err := svc.Refresh("invalid", ""); err == nil {
		t.Error("refresh token invalid trebuie să dea eroare")
	}

	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(refreshColumns).AddRow(int64(9), future, false, "jti"))
	if _, _, err := svc.Refresh("token", ""); err == nil {
		t.Error("utilizator inexistent trebuie să dea eroare")
	}

	oldAccess, _ := svc.jwtService.GenerateAccess("1", 1, "admin", "Admin")
	for _, accessToken := range []string{oldAccess, ""} {
		mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
			WillReturnRows(sqlmock.NewRows(refreshColumns).AddRow(int64(1), future, false, "jti-vechi"))
		mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
		newAccess, newRefresh, err := svc.Refresh("token", accessToken)
		if err != nil || newAccess == "" || newRefresh == "" {
			t.Fatalf("Refresh: %v", err)
		}
	}

	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(refreshColumns).AddRow(int64(1), future, false, "jti-vechi"))
	mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnError(errors.New("db down"))
	if _, _, err := svc.Refresh("token", ""); err == nil {
		t.Error("eroarea la revocare trebuie propagată")
	}

	mock.ExpectQuery("SELECT user_id, expires_at, revoked, access_jti FROM refresh_tokens").
		WillReturnRows(sqlmock.NewRows(refreshColumns).AddRow(int64(1), future, false, "jti-vechi"))
	users.getByID = func(int64) (*domain.User, error) { return nil, errors.New("db down") }
	if _, _, err := svc.Refresh("token", ""); err == nil {
		t.Error("eroarea de citire a utilizatorului trebuie propagată")
	}

	mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.Logout("token", oldAccess); err != nil {
		t.Errorf("Logout: %v", err)
	}
	mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.Logout("token", ""); err != nil {
		t.Errorf("Logout fără access token: %v", err)
	}
	mock.ExpectExec("UPDATE refresh_tokens SET revoked = TRUE WHERE token").WillReturnError(errors.New("db down"))
	if err := svc.Logout("token", ""); err == nil {
		t.Error("eroarea la revocare trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAuthService_Register(t *testing.T) {
	svc, users, roles, mock := newAuthService(t, nil)

	if err := svc.Register("ana@x.ro", "Parola1!", ""); err == nil {
		t.Error("emailul deja confirmat trebuie să dea eroare")
	}
	// utilizator neconfirmat → retrimite confirmarea; fără serviciu de e-mail configurat
	if err := svc.Register("nou@x.ro", "Parola1!", ""); err == nil || !strings.Contains(err.Error(), "email service") {
		t.Errorf("fără serviciu de e-mail: %v", err)
	}
	if err := svc.Register("alt@x.ro", "Parola1!", "inexistent"); err == nil {
		t.Error("rolul inexistent trebuie să dea eroare")
	}

	// utilizator nou, rol implicit „viewer”; SMTP-ul e inaccesibil în teste, deci verificăm
	// că token-ul de confirmare a fost salvat și că trimiterea a fost încercată
	svc.emailService = deadEmailService()
	var created *domain.User
	users.create = func(u *domain.User) error { u.ID = 3; created = u; return nil }
	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_confirmation_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	err := svc.Register("alt@x.ro", "Parola1!", "")
	if err == nil || strings.Contains(err.Error(), "email service") {
		t.Errorf("mă așteptam la eroare de SMTP, nu: %v", err)
	}
	if created == nil || created.RoleCode != "viewer" || created.PasswordHash == "" {
		t.Errorf("utilizatorul nu a fost creat corect: %+v", created)
	}

	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_confirmation_tokens").WillReturnError(errors.New("db down"))
	if err := svc.Register("alt@x.ro", "Parola1!", ""); err == nil || strings.Contains(err.Error(), "connect") {
		t.Errorf("eroarea la salvarea token-ului trebuie propagată: %v", err)
	}

	users.create = func(*domain.User) error { return errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!", ""); err == nil {
		t.Error("eroarea de creare trebuie propagată")
	}
	roles.getByCode = func(string) (*domain.Role, error) { return nil, errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!", ""); err == nil {
		t.Error("eroarea de citire a rolului trebuie propagată")
	}
	users.getByEmail = func(string) (*domain.User, error) { return nil, errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!", ""); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAuthService_PasswordReset(t *testing.T) {
	svc, users, _, mock := newAuthService(t, nil)

	if err := svc.ForgotPassword("necunoscut@x.ro"); err != nil {
		t.Errorf("emailul necunoscut nu trebuie să dea eroare (nu dezvăluim existența): %v", err)
	}
	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO password_reset_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := svc.ForgotPassword("ana@x.ro"); err == nil || !strings.Contains(err.Error(), "email service") {
		t.Errorf("fără serviciu de e-mail: %v", err)
	}
	svc.emailService = deadEmailService()
	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO password_reset_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := svc.ForgotPassword("ana@x.ro"); err == nil || strings.Contains(err.Error(), "email service") {
		t.Errorf("mă așteptam la eroare de SMTP: %v", err)
	}
	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO password_reset_tokens").WillReturnError(errors.New("db down"))
	if err := svc.ForgotPassword("ana@x.ro"); err == nil {
		t.Error("eroarea la salvarea token-ului trebuie propagată")
	}

	resetColumns := []string{"user_id", "expires_at", "used"}
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").WillReturnError(sql.ErrNoRows)
	if err := svc.ResetPassword("bad", "Parola2!"); err == nil {
		t.Error("token invalid trebuie să dea eroare")
	}

	var newHash string
	users.updatePassword = func(_ int64, hash string) error { newHash = hash; return nil }
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").
		WillReturnRows(sqlmock.NewRows(resetColumns).AddRow(int64(1), time.Now().Add(time.Hour), false))
	mock.ExpectExec("UPDATE password_reset_tokens SET used = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.ResetPassword("token", "Parola2!"); err != nil {
		t.Fatalf("ResetPassword: %v", err)
	}
	if !auth.CheckPasswordHash("Parola2!", newHash) {
		t.Error("parola nouă trebuie salvată ca hash")
	}

	users.updatePassword = func(int64, string) error { return errors.New("db down") }
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM password_reset_tokens").
		WillReturnRows(sqlmock.NewRows(resetColumns).AddRow(int64(1), time.Now().Add(time.Hour), false))
	if err := svc.ResetPassword("token", "Parola2!"); err == nil {
		t.Error("eroarea la actualizarea parolei trebuie propagată")
	}

	users.getByEmail = func(string) (*domain.User, error) { return nil, errors.New("db down") }
	if err := svc.ForgotPassword("ana@x.ro"); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAuthService_ConfirmEmail(t *testing.T) {
	svc, users, _, mock := newAuthService(t, nil)
	columns := []string{"user_id", "expires_at", "used"}

	confirmed := false
	users.markEmailConfirmed = func(int64) error { confirmed = true; return nil }
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(2), time.Now().Add(time.Hour), false))
	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE token").WillReturnResult(sqlmock.NewResult(0, 1))
	if err := svc.ConfirmEmail("token"); err != nil || !confirmed {
		t.Fatalf("ConfirmEmail: %v", err)
	}

	// token deja folosit, dar utilizatorul e confirmat → idempotent
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(1), time.Now().Add(time.Hour), true))
	mock.ExpectQuery("SELECT user_id FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(1)))
	if err := svc.ConfirmEmail("token"); err != nil {
		t.Errorf("confirmarea repetată trebuie să fie idempotentă: %v", err)
	}
	// token folosit, utilizator neconfirmat → eroare
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(2), time.Now().Add(time.Hour), true))
	mock.ExpectQuery("SELECT user_id FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(int64(2)))
	if err := svc.ConfirmEmail("token"); err == nil {
		t.Error("token folosit pentru utilizator neconfirmat trebuie să dea eroare")
	}
	// token inexistent
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery("SELECT user_id FROM email_confirmation_tokens").WillReturnError(sql.ErrNoRows)
	if err := svc.ConfirmEmail("bad"); err == nil {
		t.Error("token inexistent trebuie să dea eroare")
	}

	users.markEmailConfirmed = func(int64) error { return errors.New("db down") }
	mock.ExpectQuery("SELECT user_id, expires_at, used FROM email_confirmation_tokens").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(int64(2), time.Now().Add(time.Hour), false))
	if err := svc.ConfirmEmail("token"); err == nil {
		t.Error("eroarea la marcare trebuie propagată")
	}

	if err := svc.ResendConfirmation("necunoscut@x.ro"); err != nil {
		t.Errorf("email necunoscut: %v", err)
	}
	if err := svc.ResendConfirmation("ana@x.ro"); err != nil {
		t.Errorf("email deja confirmat: %v", err)
	}
	if err := svc.ResendConfirmation("nou@x.ro"); err == nil {
		t.Error("fără serviciu de e-mail trebuie să dea eroare")
	}
	users.getByEmail = func(string) (*domain.User, error) { return nil, errors.New("db down") }
	if err := svc.ResendConfirmation("nou@x.ro"); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("așteptări sqlmock: %v", err)
	}
}

func TestAuthService_ChangePassword(t *testing.T) {
	svc, users, _, _ := newAuthService(t, deadEmailService())

	if err := svc.ChangePassword(9, "x", "y"); err == nil {
		t.Error("utilizator inexistent trebuie să dea eroare")
	}
	if err := svc.ChangePassword(1, "gresit", "Parola2!"); err == nil {
		t.Error("parola curentă greșită trebuie să dea eroare")
	}
	updated := false
	users.updatePassword = func(int64, string) error { updated = true; return nil }
	if err := svc.ChangePassword(1, "Parola1!", "Parola2!"); err != nil || !updated {
		t.Errorf("ChangePassword: %v", err)
	}
	users.updatePassword = func(int64, string) error { return errors.New("db down") }
	if err := svc.ChangePassword(1, "Parola1!", "Parola2!"); err == nil {
		t.Error("eroarea la actualizare trebuie propagată")
	}
	users.getByID = func(int64) (*domain.User, error) { return nil, errors.New("db down") }
	if err := svc.ChangePassword(1, "Parola1!", "Parola2!"); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}
