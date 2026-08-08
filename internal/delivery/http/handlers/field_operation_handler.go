package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FieldOperationHandler struct {
	service *usecase.FieldOperationService
}

func NewFieldOperationHandler(service *usecase.FieldOperationService) *FieldOperationHandler {
	return &FieldOperationHandler{service: service}
}

type createFieldOperationRequest struct {
	FieldID             string                      `json:"field_id" binding:"required"`
	OperationTypeID     int64                       `json:"operation_type_id" binding:"required"`
	OperationTemplateID *int64                      `json:"operation_template_id"`
	MachineID           *int64                      `json:"machine_id"`
	ImplementID         *int64                      `json:"implement_id"`
	OperatorID          *int64                      `json:"operator_id"`
	PlannedStartAt      *time.Time                  `json:"planned_start_at"`
	PlannedEndAt        *time.Time                  `json:"planned_end_at"`
	AreaPlannedHa       *float64                    `json:"area_planned_ha"`
	Notes               string                      `json:"notes"`
	Status              domain.FieldOperationStatus `json:"status"`
}

type updateFieldOperationRequest = createFieldOperationRequest

func toDomainFieldOperation(req createFieldOperationRequest) *domain.FieldOperation {
	return &domain.FieldOperation{
		FieldID:             req.FieldID,
		OperationTypeID:     req.OperationTypeID,
		OperationTemplateID: req.OperationTemplateID,
		MachineID:           req.MachineID,
		ImplementID:         req.ImplementID,
		OperatorID:          req.OperatorID,
		PlannedStartAt:      req.PlannedStartAt,
		PlannedEndAt:        req.PlannedEndAt,
		AreaPlannedHa:       req.AreaPlannedHa,
		Notes:               req.Notes,
		Status:              req.Status,
	}
}

func (h *FieldOperationHandler) GetAll(c *gin.Context) {
	filter := repository.FieldOperationFilter{
		Status:          c.Query("status"),
		FieldID:         c.Query("field_id"),
		OperationTypeID: c.Query("operation_type_id"),
		MachineID:       c.Query("machine_id"),
		OperatorID:      c.Query("operator_id"),
	}
	items, err := h.service.GetAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *FieldOperationHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, err := h.service.GetByID(id)
	if err != nil {
		if errors.Is(err, usecase.ErrFieldOperationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *FieldOperationHandler) Create(c *gin.Context) {
	var req createFieldOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Create(toDomainFieldOperation(req))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func (h *FieldOperationHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateFieldOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.Update(id, toDomainFieldOperation(req))
	if err != nil {
		if errors.Is(err, usecase.ErrFieldOperationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *FieldOperationHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.service.Delete(id); err != nil {
		if errors.Is(err, usecase.ErrFieldOperationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
