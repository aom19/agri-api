package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)


type PermissionHandler struct {
	service *usecase.PermissionService
}

func NewPermissionHandler(service *usecase.PermissionService) *PermissionHandler {
	return &PermissionHandler{service: service}
}

func (h *PermissionHandler) GetAll(c *gin.Context) {
	permissions, err := h.service.GetPermissions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, permissions)
}

func (h *PermissionHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	permission, err := h.service.GetPermissionByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if permission == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Permission nu a fost găsita"})
		return
	}
	c.JSON(http.StatusOK, permission)
}