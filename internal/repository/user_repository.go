package repository

import "agri-api/internal/domain"

type UserRepository interface {
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
	Create(user *domain.User) error
	UpdatePassword(id int64, passwordHash string) error
	UpdateRole(userID int64, roleID int64) error
	GetProfile(userID int64) (*domain.UserProfile, error)
	UpsertProfile(profile *domain.UserProfile) error
}
