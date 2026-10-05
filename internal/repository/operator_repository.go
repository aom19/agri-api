package repository

import (
	"agri-api/internal/domain"
	"errors"
)

// ErrEmailTaken: adresa de e-mail aparține deja altui cont.
var ErrEmailTaken = errors.New("email already in use")

// OperatorRepository lucrează pe utilizatorii cu rolul `operator` (users + user_profiles).
type OperatorRepository interface {
	GetAll() ([]domain.Operator, error)
	GetByID(id int64) (*domain.Operator, error)
	// Create creează contul (rol operator, fără parolă utilizabilă) și profilul, într-o tranzacție.
	Create(operator *domain.Operator) error
	// Update schimbă e-mailul contului și profilul.
	Update(id int64, operator *domain.Operator) error
	// SetActive activează sau dezactivează contul.
	SetActive(id int64, active bool) error
}
