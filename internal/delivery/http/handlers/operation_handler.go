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
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

func NewOperationHandler(service *usecase.OperationService, opts ...func(*OperationHandler)) *OperationHandler {
	h := &OperationHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithOperationAudit(a *usecase.AuditService) func(*OperationHandler) {
	return func(h *OperationHandler) { h.audit = a }
}

func WithOperationNotif(n *usecase.NotificationService) func(*OperationHandler) {
	return func(h *OperationHandler) { h.notif = n }
}

// ─── Request DTOs ────────────────────────────────────────────────────────────

type templateResourceRequest struct {
	ResourceID      int64   `json:"resource_id" binding:"required"`
	QuantityPerUnit float64 `json:"quantity_per_unit" binding:"required"`
	Notes           string  `json:"notes"`
}

type createTemplateRequest struct {
	OperationType  domain.OperationType      `json:"operation_type" binding:"required"`
	Name           string                    `json:"name" binding:"required"`
	Description    string                    `json:"description"`
	Unit           string                    `json:"unit" binding:"required"`
	CropID         *int64                    `json:"crop_id"`
	Resources      []templateResourceRequest `json:"resources"`
	MachineTypes   []string                  `json:"machine_types"`
	ImplementTypes []string                  `json:"implement_types"`
}

type updateTemplateRequest struct {
	OperationType  domain.OperationType      `json:"operation_type" binding:"required"`
	Name           string                    `json:"name" binding:"required"`
	Description    string                    `json:"description"`
	Unit           string                    `json:"unit" binding:"required"`
	CropID         *int64                    `json:"crop_id"`
	Resources      []templateResourceRequest `json:"resources"`
	MachineTypes   []string                  `json:"machine_types"`
	ImplementTypes []string                  `json:"implement_types"`
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
		OperationType:  req.OperationType,
		Name:           req.Name,
		Description:    req.Description,
		Unit:           req.Unit,
		CropID:         req.CropID,
		Resources:      resources,
		MachineTypes:   req.MachineTypes,
		ImplementTypes: req.ImplementTypes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
	auditAndNotify(c, h.audit, h.notif, "operation_template", auditID(result.ID), "create", "Template creat", result.Name, map[string]interface{}{
		"name":           result.Name,
		"operation_type": result.OperationType,
		"unit":           result.Unit,
	})
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
		OperationType:  req.OperationType,
		Name:           req.Name,
		Description:    req.Description,
		Unit:           req.Unit,
		CropID:         req.CropID,
		Resources:      resources,
		MachineTypes:   req.MachineTypes,
		ImplementTypes: req.ImplementTypes,
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
	auditAndNotify(c, h.audit, h.notif, "operation_template", auditID(result.ID), "update", "Template actualizat", result.Name, map[string]interface{}{
		"name":           result.Name,
		"operation_type": result.OperationType,
		"unit":           result.Unit,
	})
}

func (h *OperationHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	item, _ := h.service.GetTemplateByID(id)
	if err := h.service.DeleteTemplate(id); err != nil {
		if errors.Is(err, usecase.ErrOperationTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
	message := "template"
	if item != nil {
		message = item.Name
	}
	auditAndNotify(c, h.audit, h.notif, "operation_template", auditID(id), "delete", "Template șters", message, nil)
}
