package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/email"
	"agri-api/internal/repository"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

var ErrEmailNotConfirmed = errors.New("email not confirmed")

// defaultRegisterRole e rolul atribuit oricărui cont creat prin înregistrare publică
const defaultRegisterRole = "viewer"

// AuthService conține logica de business pentru autentificare
type AuthService struct {
	users        repository.UserRepository
	roles        repository.RoleRepository
	jwtService   *auth.JWTService
	refreshRepo  *auth.Repo
	blacklist    *auth.Blacklist
	emailService *email.EmailService
	clientOrigin string
}

func NewAuthService(users repository.UserRepository, roles repository.RoleRepository, jwt *auth.JWTService, refreshRepo *auth.Repo, blacklist *auth.Blacklist, emailService *email.EmailService, clientOrigin string) *AuthService {
	if clientOrigin == "" {
		clientOrigin = "http://localhost:3000"
	}
	return &AuthService{
		users:        users,
		roles:        roles,
		jwtService:   jwt,
		refreshRepo:  refreshRepo,
		blacklist:    blacklist,
		emailService: emailService,
		clientOrigin: strings.TrimRight(clientOrigin, "/"),
	}
}

// accessClaims sunt revendicările suplimentare din access token, derivate din cont.
func accessClaims(user *domain.User) map[string]interface{} {
	if !user.MustChangePassword {
		return nil
	}
	return map[string]interface{}{auth.ClaimMustChangePassword: true}
}

// issueSession emite o pereche nouă de token-uri pentru cont și o înregistrează.
func (service *AuthService) issueSession(user *domain.User) (string, string, error) {
	access, err := service.jwtService.GenerateAccessWithClaims(fmt.Sprintf("%d", user.ID), user.RoleID, user.RoleCode, user.RoleName, accessClaims(user))
	if err != nil {
		return "", "", err
	}
	accessJTI, _, _ := service.jwtService.ExtractJTIAndTTL(access)
	refresh := auth.GenerateRefreshToken()
	if err := service.refreshRepo.StoreRefreshToken(user.ID, refresh, accessJTI, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

// revokeAllSessions revocă toate sesiunile contului și blacklistează access token-urile lor.
func (service *AuthService) revokeAllSessions(userID int64) {
	oldJTIs, _ := service.refreshRepo.RevokeAllUserSessions(userID)
	for _, jti := range oldJTIs {
		_ = service.blacklist.Add(context.Background(), jti, service.jwtService.AccessTokenTTL())
	}
}

// LoginResult este rezultatul autentificării: token-urile și dacă parola trebuie schimbată acum.
type LoginResult struct {
	AccessToken        string
	RefreshToken       string
	MustChangePassword bool
}

func (service *AuthService) Login(email, password string) (string, string, error) {
	result, err := service.LoginWithResult(email, password)
	if err != nil {
		return "", "", err
	}
	return result.AccessToken, result.RefreshToken, nil
}

// LoginWithResult autentifică utilizatorul; o sesiune nouă le revocă pe cele vechi.
func (service *AuthService) LoginWithResult(email, password string) (*LoginResult, error) {
	user, err := service.users.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}
	if !user.EmailConfirmed {
		return nil, ErrEmailNotConfirmed
	}

	service.revokeAllSessions(user.ID)

	access, refresh, err := service.issueSession(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: access, RefreshToken: refresh, MustChangePassword: user.MustChangePassword}, nil
}

func (service *AuthService) Refresh(refreshToken, oldAccessToken string) (string, string, error) {
	userID, oldJTI, err := service.refreshRepo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := service.users.GetByID(userID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", errors.New("user not found")
	}

	// Revocă refresh token-ul vechi (rotație)
	if err := service.refreshRepo.RevokeRefreshToken(refreshToken); err != nil {
		return "", "", err
	}

	// Blacklist access token-ul vechi — JTI din DB (nu depinde de client)
	if oldJTI != "" {
		if _, ttl, err := service.jwtService.ExtractJTIAndTTL(oldAccessToken); err == nil && ttl > 0 {
			_ = service.blacklist.Add(context.Background(), oldJTI, ttl)
		} else {
			// fallback: TTL complet dacă nu putem calcula ce a rămas
			_ = service.blacklist.Add(context.Background(), oldJTI, service.jwtService.AccessTokenTTL())
		}
	}

	// revendicarea de schimbare a parolei vine din cont, deci dispare de la sine după schimbare
	return service.issueSession(user)
}

func (service *AuthService) Logout(refreshToken, accessToken string) error {
	if err := service.refreshRepo.RevokeRefreshToken(refreshToken); err != nil {
		return err
	}
	// Blacklist access token-ul în Redis
	if accessToken != "" {
		if jti, ttl, err := service.jwtService.ExtractJTIAndTTL(accessToken); err == nil {
			_ = service.blacklist.Add(context.Background(), jti, ttl)
		}
	}
	return nil
}

// Register creează mereu un cont cu rolul viewer; rolurile se schimbă doar din administrarea utilizatorilor.
func (service *AuthService) Register(email, password string) error {
	existing, err := service.users.GetByEmail(email)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.EmailConfirmed {
			return errors.New("email already in use")
		}
		return service.sendAccountConfirmation(existing.ID, existing.Email)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	roleEntity, err := service.roles.GetByCode(defaultRegisterRole)
	if err != nil {
		return err
	}
	if roleEntity == nil {
		return errors.New("role not found: " + defaultRegisterRole)
	}

	user := &domain.User{Email: email, PasswordHash: hash, RoleID: roleEntity.ID, RoleCode: roleEntity.Code, RoleName: roleEntity.Name}
	if err := service.users.Create(user); err != nil {
		return err
	}

	return service.sendAccountConfirmation(user.ID, user.Email)
}

func (service *AuthService) ForgotPassword(email string) error {
	user, err := service.users.GetByEmail(email)
	if err != nil {
		return err
	}
	// Nu dezvăluim dacă email-ul există sau nu (securitate)
	if user == nil {
		return nil
	}

	token := auth.GenerateRefreshToken() // UUID random
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := service.refreshRepo.StoreResetToken(user.ID, token, expiresAt); err != nil {
		return err
	}
	if service.emailService == nil {
		return errors.New("email service is not configured")
	}

	resetURL := fmt.Sprintf("%s/reset-password/%s", service.clientOrigin, url.QueryEscape(token))
	if err := service.emailService.SendPasswordReset(user.Email, resetURL); err != nil {
		return err
	}

	return nil
}

func (service *AuthService) ResetPassword(token, newPassword string) error {
	userID, err := service.refreshRepo.ValidateResetToken(token)
	if err != nil {
		return err
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := service.users.UpdatePassword(userID, hash); err != nil {
		return err
	}

	return service.refreshRepo.InvalidateResetToken(token)
}

func (service *AuthService) ConfirmEmail(token string) error {
	userID, err := service.refreshRepo.ValidateEmailConfirmationToken(token)
	if err != nil {
		// Token deja folosit — verifică dacă userul e deja confirmat (idempotent)
		if uid, lookupErr := service.refreshRepo.GetEmailConfirmationTokenUserID(token); lookupErr == nil {
			if user, userErr := service.users.GetByID(uid); userErr == nil && user != nil && user.EmailConfirmed {
				return nil
			}
		}
		return err
	}

	if err := service.users.MarkEmailConfirmed(userID); err != nil {
		return err
	}

	return service.refreshRepo.InvalidateEmailConfirmationToken(token)
}

func (service *AuthService) ResendConfirmation(emailAddr string) error {
	user, err := service.users.GetByEmail(emailAddr)
	if err != nil {
		return err
	}
	// Răspuns generic — nu dezvăluim dacă emailul există
	if user == nil {
		return nil
	}
	if user.EmailConfirmed {
		return nil
	}
	return service.sendAccountConfirmation(user.ID, user.Email)
}

func (service *AuthService) sendAccountConfirmation(userID int64, emailAddr string) error {
	if service.emailService == nil {
		return errors.New("email service is not configured")
	}

	token := auth.GenerateRefreshToken()
	expiresAt := time.Now().Add(24 * time.Hour)
	if err := service.refreshRepo.StoreEmailConfirmationToken(userID, token, expiresAt); err != nil {
		return err
	}

	confirmURL := fmt.Sprintf("%s/confirm-email?token=%s", service.clientOrigin, url.QueryEscape(token))
	return service.emailService.SendAccountConfirmation(emailAddr, confirmURL)
}

// ChangePassword verifică parola veche și o înlocuiește cu una nouă
func (service *AuthService) ChangePassword(userID int64, oldPassword, newPassword string) error {
	_, err := service.ChangePasswordWithSession(userID, oldPassword, newPassword)
	return err
}

// ChangePasswordWithSession schimbă parola, ridică obligația de schimbare, revocă toate sesiunile
// (inclusiv pe cea curentă, care nu mai poartă revendicarea veche) și emite una nouă.
func (service *AuthService) ChangePasswordWithSession(userID int64, oldPassword, newPassword string) (*LoginResult, error) {
	user, err := service.users.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	if !auth.CheckPasswordHash(oldPassword, user.PasswordHash) {
		return nil, errors.New("parola curentă este incorectă")
	}
	if oldPassword == newPassword {
		return nil, errors.New("parola nouă trebuie să fie diferită de cea curentă")
	}
	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	if err := service.users.UpdatePassword(userID, hash); err != nil {
		return nil, err
	}
	if user.MustChangePassword {
		if err := service.users.SetMustChangePassword(userID, false); err != nil {
			return nil, err
		}
		user.MustChangePassword = false
	}

	result := &LoginResult{}
	if service.refreshRepo != nil && service.jwtService != nil {
		service.revokeAllSessions(userID)
		access, refresh, err := service.issueSession(user)
		if err != nil {
			return nil, err
		}
		result.AccessToken, result.RefreshToken = access, refresh
	}

	// Trimite email de confirmare (non-blocant față de răspunsul HTTP)
	if service.emailService != nil {
		go func() { _ = service.emailService.SendPasswordChanged(user.Email) }()
	}
	return result, nil
}
