package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// OperatorHandler gestionează request-urile HTTP pentru resursa operatori
type OperatorHandler struct {
	service *usecase.OperatorService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

// CreateOperatorRequest creează contul (rol operator) și profilul. Fără e-mail, contul primește
// o adresă tehnică și nu se poate folosi până nu se completează e-mailul real.
type CreateOperatorRequest struct {
	FirstName           string               `json:"first_name" binding:"required"`
	LastName            string               `json:"last_name"`
	Phone               string               `json:"phone"`
	Email               string               `json:"email" binding:"omitempty,email"`
	Notes               string               `json:"notes"`
	AllowedMachineTypes []domain.MachineType `json:"allowed_machine_types"`
}

type UpdateOperatorRequest = CreateOperatorRequest

func (req CreateOperatorRequest) toOperator() *domain.Operator {
	return &domain.Operator{
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Phone:               req.Phone,
		Email:               req.Email,
		Notes:               req.Notes,
		AllowedMachineTypes: req.AllowedMachineTypes,
	}
}

// operatorErrorStatus alege codul HTTP pentru erorile serviciului de operatori.
func operatorErrorStatus(err error) int {
	switch {
	case errors.Is(err, usecase.ErrOperatorNotFound):
		return http.StatusNotFound
	case errors.Is(err, usecase.ErrOperatorInvalid):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// NewOperatorHandler creează un handler nou cu serviciul injectat
func NewOperatorHandler(service *usecase.OperatorService, opts ...func(*OperatorHandler)) *OperatorHandler {
	h := &OperatorHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithOperatorAudit(a *usecase.AuditService) func(*OperatorHandler) {
	return func(h *OperatorHandler) { h.audit = a }
}

func WithOperatorNotif(n *usecase.NotificationService) func(*OperatorHandler) {
	return func(h *OperatorHandler) { h.notif = n }
}

// Create creează un operator nou
// @Summary      Creare operator
// @Tags         operators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateOperatorRequest true "Date operator"
// @Success      201 {object} domain.Operator
// @Failure      400 {object} object{error=string}
// @Router       /operators [post]
func (h *OperatorHandler) Create(c *gin.Context) {
	var req CreateOperatorRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	operator, err := h.service.CreateOperator(req.toOperator())
	if err != nil {
		c.JSON(operatorErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, operator)
	auditAndNotify(c, h.audit, h.notif, "operator", auditID(operator.ID), "create", "Operator creat", operator.Name, map[string]interface{}{
		"name":   operator.Name,
		"email":  operator.Email,
		"status": string(operator.Status),
	})
}

// GetAll returnează toți operatorii
// @Summary      Listare operatori
// @Tags         operators
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Operator
// @Router       /operators [get]
func (h *OperatorHandler) GetAll(c *gin.Context) {
	operators, err := h.service.GetOperators()

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"operators": []domain.Operator{}})
		return
	}
	c.JSON(http.StatusOK, operators)
}

// GetByID returnează un operator după ID
// @Summary      Obținere operator
// @Tags         operators
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID operator"
// @Success      200 {object} domain.Operator
// @Failure      404 {object} object{error=string}
// @Router       /operators/{id} [get]
func (h *OperatorHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	operator, err := h.service.GetOperatorByID(operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if operator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Operator not found"})
		return
	}
	c.JSON(http.StatusOK, operator)
}

// Update actualizează un operator
// @Summary      Actualizare operator
// @Tags         operators
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID operator"
// @Param        body body UpdateOperatorRequest true "Date actualizare"
// @Success      200 {object} domain.Operator
// @Failure      400 {object} object{error=string}
// @Router       /operators/{id} [patch]
func (h *OperatorHandler) Update(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	var req UpdateOperatorRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	operator, err := h.service.UpdateOperator(operatorID, req.toOperator())
	if err != nil {
		c.JSON(operatorErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, operator)
	auditAndNotify(c, h.audit, h.notif, "operator", auditID(operator.ID), "update", "Operator actualizat", operator.Name, map[string]interface{}{
		"name":  operator.Name,
		"email": operator.Email,
	})
}

// Delete șterge operatorul, adică îi dezactivează contul (ca în pagina Utilizatori).
// @Summary      Ștergere operator
// @Tags         operators
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID operator"
// @Success      204
// @Failure      400 {object} object{error=string}
// @Router       /operators/{id} [delete]
func (h *OperatorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	operator, err := h.service.GetOperatorByID(operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.DeleteOperator(operatorID); err != nil {
		c.JSON(operatorErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
	name := "operator"
	if operator != nil {
		name = operator.Name
	}
	auditAndNotify(c, h.audit, h.notif, "operator", id, "delete", "Operator șters", name, nil)
}

// Disable dezactivează un operator (setează status la inactive)
// @Summary      Dezactivare operator
// @Tags         operators
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID operator"
// @Success      200 {object} domain.Operator
// @Failure      400 {object} object{error=string}
// @Router       /operators/{id}/disable [patch]
func (h *OperatorHandler) Disable(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	oldOperator, _ := h.service.GetOperatorByID(operatorID)
	operator, err := h.service.DisableOperator(operatorID)
	if err != nil {
		c.JSON(operatorErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, operator)
	changes := statusChange("", string(operator.Status))
	if oldOperator != nil {
		changes["old_status"] = string(oldOperator.Status)
	}
	auditAndNotify(c, h.audit, h.notif, "operator", id, "disable", "Operator dezactivat", operator.Name, changes)
}

// Enable reactivează un operator (setează status la active)
// @Summary      Reactivare operator
// @Tags         operators
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID operator"
// @Success      200 {object} domain.Operator
// @Failure      400 {object} object{error=string}
// @Router       /operators/{id}/enable [patch]
func (h *OperatorHandler) Enable(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	oldOperator, _ := h.service.GetOperatorByID(operatorID)
	operator, err := h.service.EnableOperator(operatorID)
	if err != nil {
		c.JSON(operatorErrorStatus(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, operator)
	changes := statusChange("", string(operator.Status))
	if oldOperator != nil {
		changes["old_status"] = string(oldOperator.Status)
	}
	auditAndNotify(c, h.audit, h.notif, "operator", id, "enable", "Operator activat", operator.Name, changes)
}
