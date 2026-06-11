package repository

import "agri-api/internal/domain"

type FieldRepository interface {
	Create(field *domain.Field) error
	GetAll() ([]domain.Field, error)
	GetByID(id string) (*domain.Field, error)
	Update(id string, field *domain.Field) error
	Delete(id string) error
}
