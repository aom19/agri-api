package repository

import "agri-api/internal/domain"

type AuditRepository interface {
	Log(entry *domain.AuditEntry) error
	GetAll(limit int, entityType, entityID string) ([]domain.AuditEntry, error)
	// List: filtre și paginare opțională, cele mai noi primele.
	List(filter domain.AuditFilter, page *domain.PageRequest) ([]domain.AuditEntry, int, error)
}
