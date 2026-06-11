package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// FieldHandler gestioneaza request-urile HTTP pentru terenuri.
type FieldHandler struct {
	service *usecase.FieldService
}

func NewFieldHandler(service *usecase.FieldService) *FieldHandler {
	return &FieldHandler{service: service}
}

type createFieldRequest struct {
	Name     string          `json:"name" binding:"required"`
	AreaHa   *float64        `json:"area_ha"`
	Geometry json.RawMessage `json:"geometry" binding:"required"`
}

type updateFieldRequest struct {
	Name     string          `json:"name" binding:"required"`
	AreaHa   *float64        `json:"area_ha"`
	Geometry json.RawMessage `json:"geometry" binding:"required"`
}

func (h *FieldHandler) Create(c *gin.Context) {
	var req createFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field, err := h.service.CreateField(&domain.Field{
		Name:     req.Name,
		AreaHa:   req.AreaHa,
		Geometry: req.Geometry,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, field)
}

func (h *FieldHandler) GetAll(c *gin.Context) {
	fields, err := h.service.GetFields()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func (h *FieldHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	field, err := h.service.GetFieldByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "field not found"})
		return
	}
	c.JSON(http.StatusOK, field)
}

func (h *FieldHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req updateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateField(id, &domain.Field{
		Name:     req.Name,
		AreaHa:   req.AreaHa,
		Geometry: req.Geometry,
	})
	if err != nil {
		if err.Error() == "field not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *FieldHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteField(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
