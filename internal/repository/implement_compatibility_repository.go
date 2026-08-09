package repository

import "agri-api/internal/domain"

type ImplementCompatibilityRepository interface {
	GetAll() ([]domain.ImplementCompatibility, error)
}
