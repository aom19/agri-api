package repository

import "agri-api/internal/domain"

type UserRepository interface {
	GetAll() ([]domain.User, error)
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int64) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	Delete(id int64) error
	Disable(id int64) error
	Enable(id int64) error
	MarkEmailConfirmed(id int64) error
	UpdatePassword(id int64, passwordHash string) error
	UpdateRole(userID int64, roleID int64) error
	GetProfile(userID int64) (*domain.UserProfile, error)
	UpsertProfile(profile *domain.UserProfile) error
}
