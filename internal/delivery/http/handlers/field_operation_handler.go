package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FieldOperationHandler struct {
	service *usecase.FieldOperationService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

func NewFieldOperationHandler(service *usecase.FieldOperationService, opts ...func(*FieldOperationHandler)) *FieldOperationHandler {
	h := &FieldOperationHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithFieldOpAudit(a *usecase.AuditService) func(*FieldOperationHandler) {
	return func(h *FieldOperationHandler) { h.audit = a }
}

func WithFieldOpNotif(n *usecase.NotificationService) func(*FieldOperationHandler) {
	return func(h *FieldOperationHandler) { h.notif = n }
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

type updateFieldOperationChecklistRequest struct {
	MachineStatus   bool `json:"machine_status"`
	ImplementStatus bool `json:"implement_status"`
	FieldArea       bool `json:"field_area"`
	NotesConfirmed  bool `json:"notes_confirmed"`
}

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

func toDomainFieldOperationChecklist(req updateFieldOperationChecklistRequest) domain.FieldOperationChecklist {
	return domain.FieldOperationChecklist{
		MachineStatus:   req.MachineStatus,
		ImplementStatus: req.ImplementStatus,
		FieldArea:       req.FieldArea,
		NotesConfirmed:  req.NotesConfirmed,
	}
}

func isOperatorRequest(c *gin.Context) bool {
	roleCode, _ := c.Get("role_code")
	return roleCode == "operator"
}

func currentUserID(c *gin.Context) (int64, bool) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	switch value := userIDRaw.(type) {
	case string:
		userID, err := strconv.ParseInt(value, 10, 64)
		if err != nil || userID <= 0 {
			return 0, false
		}
		return userID, true
	case float64:
		if value <= 0 || value > math.MaxInt64 {
			return 0, false
		}
		return int64(value), true
	case int64:
		return value, value > 0
	case int:
		return int64(value), value > 0
	default:
		return 0, false
	}
}

func currentActorID(c *gin.Context) *int64 {
	userID, ok := currentUserID(c)
	if !ok {
		return nil
	}
	return &userID
}

func (h *FieldOperationHandler) GetAll(c *gin.Context) {
	filter := repository.FieldOperationFilter{
		Status:          c.Query("status"),
		FieldID:         c.Query("field_id"),
		OperationTypeID: c.Query("operation_type_id"),
		MachineID:       c.Query("machine_id"),
		OperatorID:      c.Query("operator_id"),
	}
	if isOperatorRequest(c) {
		userID, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		filter.OperatorID = ""
		filter.AssignedUserID = userID
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
	var item interface{}
	if isOperatorRequest(c) {
		userID, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		item, err = h.service.GetByIDForAssignedUser(id, userID)
	} else {
		item, err = h.service.GetByID(id)
	}
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

	if h.audit != nil {
		h.audit.Log("field_operation", strconv.FormatInt(item.ID, 10), "create", currentActorID(c), map[string]interface{}{
			"status": string(item.Status),
		})
	}
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
	oldItem, oldErr := h.service.GetByID(id)
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

	if h.audit != nil {
		changes := map[string]interface{}{"status": string(req.Status)}
		if oldErr == nil && oldItem != nil {
			changes["old_status"] = string(oldItem.Status)
		}
		h.audit.Log("field_operation", strconv.FormatInt(id, 10), "update", currentActorID(c), changes)
	}
}

func (h *FieldOperationHandler) UpdateChecklist(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateFieldOperationChecklistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	checklist := toDomainFieldOperationChecklist(req)
	var item interface{}
	if isOperatorRequest(c) {
		userID, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		item, err = h.service.UpdateChecklistForAssignedUser(id, userID, checklist)
	} else {
		item, err = h.service.UpdateChecklist(id, checklist)
	}
	if err != nil {
		if errors.Is(err, usecase.ErrFieldOperationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)

	if h.audit != nil {
		h.audit.Log("field_operation", strconv.FormatInt(id, 10), "checklist", currentActorID(c), map[string]interface{}{
			"machine_status":   checklist.MachineStatus,
			"implement_status": checklist.ImplementStatus,
			"field_area":       checklist.FieldArea,
			"notes_confirmed":  checklist.NotesConfirmed,
		})
	}
}

func (h *FieldOperationHandler) Start(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	oldItem, oldErr := h.service.GetByID(id)
	var item interface{}
	if isOperatorRequest(c) {
		userID, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid token"})
			return
		}
		item, err = h.service.StartForAssignedUser(id, userID)
	} else {
		item, err = h.service.Start(id)
	}
	if err != nil {
		if errors.Is(err, usecase.ErrFieldOperationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)

	if h.audit != nil {
		changes := map[string]interface{}{"status": string(domain.FieldOperationStatusInProgress)}
		if oldErr == nil && oldItem != nil {
			changes["old_status"] = string(oldItem.Status)
		}
		h.audit.Log("field_operation", strconv.FormatInt(id, 10), "start", currentActorID(c), changes)
	}
	if h.notif != nil {
		h.notif.Emit(
			domain.NotifOperationStarted,
			"Lucrare pornită",
			fmt.Sprintf("Lucrarea #%d a fost pornită", id),
			"field_operation", strconv.FormatInt(id, 10),
		)
	}
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
