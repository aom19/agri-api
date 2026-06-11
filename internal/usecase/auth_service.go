package usecase

import (
	"agri-api/internal/auth"
	"agri-api/internal/domain"
	"agri-api/internal/store"
	"context"
	"errors"
	"fmt"
	"time"
)

// AuthService conține logica de business pentru autentificare
type AuthService struct {
	store       *store.Store
	jwtService  *auth.JWTService
	refreshRepo *auth.Repo
	blacklist   *auth.Blacklist
}

func NewAuthService(s *store.Store, jwt *auth.JWTService, refreshRepo *auth.Repo, blacklist *auth.Blacklist) *AuthService {
	return &AuthService{
		store:       s,
		jwtService:  jwt,
		refreshRepo: refreshRepo,
		blacklist:   blacklist,
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

	// Revocă toate sesiunile active și blacklistează access token-urile lor
	oldJTIs, _ := service.refreshRepo.RevokeAllUserSessions(user.ID)
	for _, jti := range oldJTIs {
		_ = service.blacklist.Add(context.Background(), jti, service.jwtService.AccessTokenTTL())
	}

	access, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.RoleID, user.RoleCode, user.RoleName)
	if err != nil {
		return "", "", err
	}

	accessJTI, _, _ := service.jwtService.ExtractJTIAndTTL(access)
	refresh := auth.GenerateRefreshToken()
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := service.refreshRepo.StoreRefreshToken(user.ID, refresh, accessJTI, expiresAt); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (service *AuthService) Refresh(refreshToken, oldAccessToken string) (string, string, error) {
	userID, oldJTI, err := service.refreshRepo.ValidateRefreshToken(refreshToken)
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

	// Blacklist access token-ul vechi — JTI din DB (nu depinde de client)
	if oldJTI != "" {
		if _, ttl, err := service.jwtService.ExtractJTIAndTTL(oldAccessToken); err == nil && ttl > 0 {
			_ = service.blacklist.Add(context.Background(), oldJTI, ttl)
		} else {
			// fallback: TTL complet dacă nu putem calcula ce a rămas
			_ = service.blacklist.Add(context.Background(), oldJTI, service.jwtService.AccessTokenTTL())
		}
	}

	newAccess, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.RoleID, user.RoleCode, user.RoleName)
	if err != nil {
		return "", "", err
	}

	newJTI, _, _ := service.jwtService.ExtractJTIAndTTL(newAccess)
	newRefresh := auth.GenerateRefreshToken()
	if err := service.refreshRepo.StoreRefreshToken(user.ID, newRefresh, newJTI, time.Now().Add(7*24*time.Hour)); err != nil {
		return "", "", err
	}

	return newAccess, newRefresh, nil
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
		role = "viewer"
	}

	// Găsește role_id după nume
	roleEntity, err := service.store.RoleRepo.GetByCode(role)
	if err != nil {
		return "", "", err
	}
	if roleEntity == nil {
		return "", "", errors.New("role not found: " + role)
	}

	user := &domain.User{Email: email, PasswordHash: hash, RoleID: roleEntity.ID, RoleCode: roleEntity.Code, RoleName: roleEntity.Name}
	if err := service.store.UserRepo.Create(user); err != nil {
		return "", "", err
	}

	access, err := service.jwtService.GenerateAccess(fmt.Sprintf("%d", user.ID), user.RoleID, user.RoleCode, user.RoleName)
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
