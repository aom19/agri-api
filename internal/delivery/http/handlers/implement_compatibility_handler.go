package handlers

import (
	"net/http"

	"agri-api/internal/repository"

	"github.com/gin-gonic/gin"
)

type ImplementCompatibilityHandler struct {
	repo repository.ImplementCompatibilityRepository
}

func NewImplementCompatibilityHandler(repo repository.ImplementCompatibilityRepository) *ImplementCompatibilityHandler {
	return &ImplementCompatibilityHandler{repo: repo}
}

func (h *ImplementCompatibilityHandler) GetAll(c *gin.Context) {
	items, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}
