package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OperationHandler struct {
	service *usecase.OperationService
}

func NewOperationHandler(service *usecase.OperationService) *OperationHandler {
	return &OperationHandler{service: service}
}

// ─── Request DTOs ────────────────────────────────────────────────────────────

type createOperationTypeRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type updateOperationTypeRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type templateResourceRequest struct {
	ResourceID      int64   `json:"resource_id" binding:"required"`
	QuantityPerUnit float64 `json:"quantity_per_unit" binding:"required"`
	Notes           string  `json:"notes"`
}

type createTemplateRequest struct {
	OperationTypeID int64                     `json:"operation_type_id" binding:"required"`
	Name            string                    `json:"name" binding:"required"`
	Description     string                    `json:"description"`
	Unit            string                    `json:"unit" binding:"required"`
	Resources       []templateResourceRequest `json:"resources"`
	MachineTypes    []string                  `json:"machine_types"`
	ImplementTypes  []string                  `json:"implement_types"`
}

type updateTemplateRequest struct {
	OperationTypeID int64                     `json:"operation_type_id" binding:"required"`
	Name            string                    `json:"name" binding:"required"`
	Description     string                    `json:"description"`
	Unit            string                    `json:"unit" binding:"required"`
	Resources       []templateResourceRequest `json:"resources"`
	MachineTypes    []string                  `json:"machine_types"`
	ImplementTypes  []string                  `json:"implement_types"`
}

// ─── OperationType Handlers ──────────────────────────────────────────────────

func (h *OperationHandler) GetAllTypes(c *gin.Context) {
	items, err := h.service.GetAllTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OperationHandler) GetTypeByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.service.GetTypeByID(id)
	if err != nil {
		if errors.Is(err, usecase.ErrOperationTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *OperationHandler) CreateType(c *gin.Context) {
	var req createOperationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.CreateType(&domain.OperationType{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *OperationHandler) UpdateType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateOperationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.UpdateType(id, &domain.OperationType{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrOperationTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *OperationHandler) DeleteType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteType(id); err != nil {
		if errors.Is(err, usecase.ErrOperationTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ─── OperationTemplate Handlers ──────────────────────────────────────────────

func (h *OperationHandler) GetAllTemplates(c *gin.Context) {
	items, err := h.service.GetAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OperationHandler) GetTemplateByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, err := h.service.GetTemplateByID(id)
	if err != nil {
		if errors.Is(err, usecase.ErrOperationTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *OperationHandler) GetTemplatesByType(c *gin.Context) {
	typeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid operation type id"})
		return
	}

	items, err := h.service.GetTemplatesByOperationType(typeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *OperationHandler) CreateTemplate(c *gin.Context) {
	var req createTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resources := make([]domain.TemplateResource, len(req.Resources))
	for i, r := range req.Resources {
		resources[i] = domain.TemplateResource{
			ResourceID:      r.ResourceID,
			QuantityPerUnit: r.QuantityPerUnit,
			Notes:           r.Notes,
		}
	}

	result, err := h.service.CreateTemplate(&domain.OperationTemplate{
		OperationTypeID: req.OperationTypeID,
		Name:            req.Name,
		Description:     req.Description,
		Unit:            req.Unit,
		Resources:       resources,
		MachineTypes:    req.MachineTypes,
		ImplementTypes:  req.ImplementTypes,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrOperationTypeNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "operation type not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *OperationHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req updateTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resources := make([]domain.TemplateResource, len(req.Resources))
	for i, r := range req.Resources {
		resources[i] = domain.TemplateResource{
			ResourceID:      r.ResourceID,
			QuantityPerUnit: r.QuantityPerUnit,
			Notes:           r.Notes,
		}
	}

	result, err := h.service.UpdateTemplate(id, &domain.OperationTemplate{
		OperationTypeID: req.OperationTypeID,
		Name:            req.Name,
		Description:     req.Description,
		Unit:            req.Unit,
		Resources:       resources,
		MachineTypes:    req.MachineTypes,
		ImplementTypes:  req.ImplementTypes,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrOperationTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *OperationHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteTemplate(id); err != nil {
		if errors.Is(err, usecase.ErrOperationTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
