package usecase_test

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/email"
	"agri-api/internal/usecase"

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

func newAuthService(t *testing.T, emailSvc *email.EmailService) (*usecase.AuthService, *userRepoMock, *roleRepoMock, sqlmock.Sqlmock) {
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
	svc := usecase.NewAuthService(users, roles, jwt, auth.NewRepo(db), deadBlacklist(), emailSvc, "http://front/")
	return svc, users, roles, mock
}

func TestNewAuthService_ClientOrigin(t *testing.T) {
	svc := usecase.NewAuthService(nil, nil, nil, nil, nil, nil, "")
	if svc.ClientOrigin() != "http://localhost:3000" {
		t.Errorf("origin implicit: %s", svc.ClientOrigin())
	}
	svc = usecase.NewAuthService(nil, nil, nil, nil, nil, nil, "http://app/")
	if svc.ClientOrigin() != "http://app" {
		t.Errorf("origin fără slash final: %s", svc.ClientOrigin())
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
	if _, _, err := svc.Login("nou@x.ro", "Parola1!"); !errors.Is(err, usecase.ErrEmailNotConfirmed) {
		t.Errorf("email neconfirmat: %v", err)
	}

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"access_jti"}).AddRow("jti-vechi"))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	access, refresh, err := svc.Login("ana@x.ro", "Parola1!")
	if err != nil || access == "" || refresh == "" {
		t.Fatalf("Login: %v", err)
	}
	claims, err := svc.JWTService().Parse(access)
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

	oldAccess, _ := svc.JWTService().GenerateAccess("1", 1, "admin", "Admin")
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

	if err := svc.Register("ana@x.ro", "Parola1!"); err == nil {
		t.Error("emailul deja confirmat trebuie să dea eroare")
	}
	// utilizator neconfirmat → retrimite confirmarea; fără serviciu de e-mail configurat
	if err := svc.Register("nou@x.ro", "Parola1!"); err == nil || !strings.Contains(err.Error(), "email service") {
		t.Errorf("fără serviciu de e-mail: %v", err)
	}
	roles.getByCode = func(string) (*domain.Role, error) { return nil, nil }
	if err := svc.Register("alt@x.ro", "Parola1!"); err == nil {
		t.Error("lipsa rolului viewer trebuie să dea eroare")
	}
	roles.getByCode = func(code string) (*domain.Role, error) {
		if code != "viewer" {
			t.Errorf("înregistrarea trebuie să ceară mereu rolul viewer, nu %q", code)
		}
		return &domain.Role{ID: 2, Code: "viewer", Name: "Vizitator"}, nil
	}

	// utilizator nou, rol implicit „viewer”; SMTP-ul e inaccesibil în teste, deci verificăm
	// că token-ul de confirmare a fost salvat și că trimiterea a fost încercată
	svc.SetEmailService(deadEmailService())
	var created *domain.User
	users.create = func(u *domain.User) error { u.ID = 3; created = u; return nil }
	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_confirmation_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	err := svc.Register("alt@x.ro", "Parola1!")
	if err == nil || strings.Contains(err.Error(), "email service") {
		t.Errorf("mă așteptam la eroare de SMTP, nu: %v", err)
	}
	if created == nil || created.RoleCode != "viewer" || created.PasswordHash == "" {
		t.Errorf("utilizatorul nu a fost creat corect: %+v", created)
	}

	mock.ExpectExec("UPDATE email_confirmation_tokens SET used = TRUE WHERE user_id").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_confirmation_tokens").WillReturnError(errors.New("db down"))
	if err := svc.Register("alt@x.ro", "Parola1!"); err == nil || strings.Contains(err.Error(), "connect") {
		t.Errorf("eroarea la salvarea token-ului trebuie propagată: %v", err)
	}

	users.create = func(*domain.User) error { return errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!"); err == nil {
		t.Error("eroarea de creare trebuie propagată")
	}
	roles.getByCode = func(string) (*domain.Role, error) { return nil, errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!"); err == nil {
		t.Error("eroarea de citire a rolului trebuie propagată")
	}
	users.getByEmail = func(string) (*domain.User, error) { return nil, errors.New("db down") }
	if err := svc.Register("alt@x.ro", "Parola1!"); err == nil {
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
	svc.SetEmailService(deadEmailService())
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
	svc, users, _, mock := newAuthService(t, deadEmailService())

	if err := svc.ChangePassword(9, "x", "y"); err == nil {
		t.Error("utilizator inexistent trebuie să dea eroare")
	}
	if err := svc.ChangePassword(1, "gresit", "Parola2!"); err == nil {
		t.Error("parola curentă greșită trebuie să dea eroare")
	}
	updated := false
	users.updatePassword = func(int64, string) error { updated = true; return nil }
	// după schimbare, sesiunile vechi sunt revocate și se emite una nouă
	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").
		WillReturnRows(sqlmock.NewRows([]string{"access_jti"}).AddRow("jti-vechi"))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	if err := svc.ChangePassword(1, "Parola1!", "Parola2!"); err != nil || !updated {
		t.Errorf("ChangePassword: %v", err)
	}
	if err := svc.ChangePassword(1, "Parola1!", "Parola1!"); err == nil {
		t.Error("parola nouă identică cu cea veche trebuie refuzată")
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

// Parola provizorie (ex. administratorul creat la pornire): autentificarea o semnalează și pune
// revendicarea în token; schimbarea parolei o ridică și întoarce o sesiune nouă, fără revendicare.
func TestAuthService_ProvisionalPassword(t *testing.T) {
	svc, users, _, mock := newAuthService(t, deadEmailService())
	hash := passwordHash(t)
	admin := &domain.User{ID: 5, Email: "admin@x.ro", PasswordHash: hash, EmailConfirmed: true, RoleID: 1, RoleCode: "admin", RoleName: "Admin", MustChangePassword: true}
	users.getByEmail = func(e string) (*domain.User, error) {
		if e == admin.Email {
			return admin, nil
		}
		return nil, nil
	}
	users.getByID = func(id int64) (*domain.User, error) {
		if id == 5 {
			return admin, nil
		}
		return nil, nil
	}
	cleared := false
	users.setMustChange = func(id int64, value bool) error { cleared = id == 5 && !value; return nil }

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").WillReturnRows(sqlmock.NewRows([]string{"access_jti"}))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	login, err := svc.LoginWithResult(admin.Email, "Parola1!")
	if err != nil || !login.MustChangePassword {
		t.Fatalf("autentificarea trebuie să semnaleze parola provizorie: %v, %+v", err, login)
	}
	claims, _ := svc.JWTService().Parse(login.AccessToken)
	if claims[auth.ClaimMustChangePassword] != true {
		t.Errorf("token-ul trebuie să poarte revendicarea: %v", claims)
	}

	mock.ExpectQuery("UPDATE refresh_tokens SET revoked = TRUE").WillReturnRows(sqlmock.NewRows([]string{"access_jti"}))
	mock.ExpectExec("INSERT INTO refresh_tokens").WillReturnResult(sqlmock.NewResult(1, 1))
	result, err := svc.ChangePasswordWithSession(5, "Parola1!", "Parola2!")
	if err != nil || !cleared || result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatalf("schimbarea trebuie să ridice obligația și să emită sesiune nouă: %v, %v, %+v", err, cleared, result)
	}
	claims, _ = svc.JWTService().Parse(result.AccessToken)
	if _, still := claims[auth.ClaimMustChangePassword]; still {
		t.Errorf("sesiunea nouă nu mai poartă revendicarea: %v", claims)
	}
}

// BootstrapAdmin: nimic fără ADMIN_EMAIL sau când există deja un admin; promovează contul
// existent; altfel creează contul confirmat, cu parolă provizorie.
func TestBootstrapAdmin(t *testing.T) {
	admins := 0
	var created *domain.User
	promoted := map[int64]int64{}
	users := &userRepoMock{
		countByRole: func(code string) (int, error) { return admins, nil },
		getByEmail: func(e string) (*domain.User, error) {
			if e == "existent@x.ro" {
				return &domain.User{ID: 4, Email: e, RoleID: 2}, nil
			}
			return nil, nil
		},
		create:     func(u *domain.User) error { u.ID = 10; created = u; return nil },
		updateRole: func(userID, roleID int64) error { promoted[userID] = roleID; return nil },
	}
	roles := &roleRepoMock{getByCode: func(code string) (*domain.Role, error) {
		if code == "admin" {
			return &domain.Role{ID: 1, Code: "admin", Name: "Administrator"}, nil
		}
		return nil, nil
	}}

	if res, err := usecase.BootstrapAdmin(users, roles, "", "x"); err != nil || res != nil {
		t.Errorf("fără ADMIN_EMAIL nu face nimic: %v, %+v", err, res)
	}
	admins = 1
	if res, err := usecase.BootstrapAdmin(users, roles, "admin@x.ro", "parola-lunga"); err != nil || res != nil {
		t.Errorf("cu un admin existent nu face nimic: %v, %+v", err, res)
	}
	admins = 0
	if _, err := usecase.BootstrapAdmin(users, roles, "admin@x.ro", "scurta"); err == nil {
		t.Error("parola prea scurtă trebuie refuzată")
	}
	res, err := usecase.BootstrapAdmin(users, roles, " Admin@X.ro ", "parola-lunga")
	if err != nil || res == nil || !res.Created || created == nil {
		t.Fatalf("crearea adminului: %v, %+v", err, res)
	}
	if created.Email != "admin@x.ro" || created.RoleCode != "admin" || !created.EmailConfirmed || !created.MustChangePassword || created.PasswordHash == "parola-lunga" {
		t.Errorf("contul creat: %+v", created)
	}
	res, err = usecase.BootstrapAdmin(users, roles, "existent@x.ro", "")
	if err != nil || res == nil || !res.Promoted || promoted[4] != 1 {
		t.Errorf("promovarea contului existent: %v, %+v, %v", err, res, promoted)
	}
}
