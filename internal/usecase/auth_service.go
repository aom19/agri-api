package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/store"
	"errors"
	"fmt"
	"time"
)

// AuthService conține logica de business pentru autentificare
type AuthService struct {
	store       *store.Store
	jwtService  *auth.JWTService
	refreshRepo *auth.Repo
}

func NewAuthService(s *store.Store, jwt *auth.JWTService, refreshRepo *auth.Repo) *AuthService {
	return &AuthService{
		store:       s,
		jwtService:  jwt,
		refreshRepo: refreshRepo,
	}
}

func (service *AuthService) Login(email, password string) (string, string, error) {
	user, err := service.store.UserRepo.GetByEmail(email)
	if err != nil {
		return "", "", err
	}
	if user == nil || !auth.CheckPasswordHash(password, user.PasswordHash) {
		return "", "", errors.New("invalid credentials")
	}

	access, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.Role)
	if err != nil {
		return "", "", err
	}
	refresh := auth.GenerateRefreshToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := service.refreshRepo.StoreRefreshToken(user.ID, refresh, expiresAt); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (service *AuthService) Refresh(refreshToken string) (string, string, error) {
	userID, err := service.refreshRepo.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := service.store.UserRepo.GetByID(userID)
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

	newAccess, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.Role)
	if err != nil {
		return "", "", err
	}

	newRefresh := auth.GenerateRefreshToken()
	if err := service.refreshRepo.StoreRefreshToken(user.ID, newRefresh, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
}

func (service *AuthService) Logout(refreshToken string) error {
	return service.refreshRepo.RevokeRefreshToken(refreshToken)
}

func (service *AuthService) Register(email, password, role string) (string, string, error) {
	existing, err := service.store.UserRepo.GetByEmail(email)
	if err != nil {
		return "", "", err
	}
	if existing != nil {
		return "", "", errors.New("email already in use")
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", "", err
	}
	if role == "" {
		role = "operator"
	}

	user := &domain.User{Email: email, PasswordHash: hash, Role: role}
	if err := service.store.UserRepo.Create(user); err != nil {
		return "", "", err
	}

	access, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.Role)
	if err != nil {
		return "", "", err
	}
	refresh := auth.GenerateRefreshToken()
	if err := service.refreshRepo.StoreRefreshToken(user.ID, refresh, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (service *AuthService) ForgotPassword(email string) (string, error) {
	user, err := service.store.UserRepo.GetByEmail(email)
	if err != nil {
		return "", err
	}
	// Nu dezvăluim dacă email-ul există sau nu (securitate)
	if user == nil {
		return "", nil
	}

	token := auth.GenerateRefreshToken() // UUID random
	expiresAt := time.Now().Add(1 * time.Hour)

	if err := service.refreshRepo.StoreResetToken(user.ID, token, expiresAt); err != nil {
		return "", err
	}

	// În producție: trimite token-ul pe email. Aici îl returnăm direct.
	return token, nil
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

	if err := service.store.UserRepo.UpdatePassword(userID, hash); err != nil {
		return err
	}

	return service.refreshRepo.InvalidateResetToken(token)
}
