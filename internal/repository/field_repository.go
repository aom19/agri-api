package repository

import "agri-api/internal/domain"

type FieldRepository interface {
	Create(field *domain.Field) error
	GetAll() ([]domain.Field, error)
	// List: filtre, ordine și paginare opțională; întoarce și totalul și suprafața totală filtrată.
	List(filter domain.FieldFilter, page *domain.PageRequest) ([]domain.Field, int, float64, error)
	GetByID(id string) (*domain.Field, error)
	Update(id string, field *domain.Field) error
	Delete(id string) error
}
