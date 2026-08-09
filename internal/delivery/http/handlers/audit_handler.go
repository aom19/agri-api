package handlers

import (
	"agri-api/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	repo repository.AuditRepository
}

func NewAuditHandler(repo repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

func (h *AuditHandler) GetAll(c *gin.Context) {
	limit := 100
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
		limit = l
	}
	entityType := c.Query("entity_type")
	entityID := c.Query("entity_id")

	entries, err := h.repo.GetAll(limit, entityType, entityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if entries == nil {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	c.JSON(http.StatusOK, entries)
}
