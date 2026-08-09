package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"fmt"
)

type AuditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

func (s *AuditService) Log(entityType, entityID, action string, actorID *int64, changes map[string]interface{}) {
	entry := &domain.AuditEntry{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		ActorID:    actorID,
		Changes:    changes,
	}
	// fire-and-forget: don't block caller on audit failure
	go func() {
		_ = s.repo.Log(entry)
	}()
}

func (s *AuditService) LogSync(entityType, entityID, action string, actorID *int64, changes map[string]interface{}) error {
	entry := &domain.AuditEntry{
		EntityType: entityType,
		EntityID:   fmt.Sprintf("%s", entityID),
		Action:     action,
		ActorID:    actorID,
		Changes:    changes,
	}
	return s.repo.Log(entry)
}
