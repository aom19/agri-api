package repository

import "agri-api/internal/domain"

type AuditRepository interface {
	Log(entry *domain.AuditEntry) error
	GetAll(limit int, entityType, entityID string) ([]domain.AuditEntry, error)
}
