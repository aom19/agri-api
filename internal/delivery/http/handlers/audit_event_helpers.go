package handlers

import (
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func auditID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func auditAndNotify(
	c *gin.Context,
	audit *usecase.AuditService,
	notif *usecase.NotificationService,
	entityType string,
	entityID string,
	action string,
	title string,
	message string,
	changes map[string]interface{},
) {
	if audit != nil {
		audit.Log(entityType, entityID, action, currentActorID(c), changes)
	}
	if notif != nil {
		notif.Emit(domain.NotifResourceIssue, title, message, entityType, entityID)
	}
}

func statusChange(oldStatus, newStatus string) map[string]interface{} {
	return map[string]interface{}{
		"old_status": oldStatus,
		"status":     newStatus,
	}
}
