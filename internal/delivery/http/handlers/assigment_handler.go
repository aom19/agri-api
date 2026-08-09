package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AssigmentHandler gestionează request-urile HTTP pentru resursa asigment
type AssigmentHandler struct {
	service *usecase.AssigmentService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

// NewAssigmentHandler creează un handler nou cu serviciul injectat
func NewAssigmentHandler(service *usecase.AssigmentService, opts ...func(*AssigmentHandler)) *AssigmentHandler {
	h := &AssigmentHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithAssigmentAudit(a *usecase.AuditService) func(*AssigmentHandler) {
	return func(h *AssigmentHandler) { h.audit = a }
}

func WithAssigmentNotif(n *usecase.NotificationService) func(*AssigmentHandler) {
	return func(h *AssigmentHandler) { h.notif = n }
}

type CreateAssigmentRequest struct {
	MachineID  int64                  `json:"machine_id" binding:"required"`
	OperatorID int64                  `json:"operator_id" binding:"required"`
	StartDate  time.Time              `json:"start_date" binding:"required"`
	EndDate    *time.Time             `json:"end_date" binding:"required"`
	Status     domain.AssigmentStatus `json:"status" binding:"required"`
}

// Create creaza o asignare noua
// @Summary      Creare asignare
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateAssigmentRequest true "Date asignare"
// @Success      201 {object} domain.Assigment
// @Failure      400 {object} object{error=string}
// @Router       /assignments [post]
func (h *AssigmentHandler) Create(c *gin.Context) {
	var req CreateAssigmentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	assigment, err := h.service.CreateAssigment(&domain.Assigment{
		MachineID:  req.MachineID,
		OperatorID: req.OperatorID,
		StartDate:  req.StartDate,
		EndDate:    req.EndDate,
		Status:     req.Status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, assigment)
	auditAndNotify(c, h.audit, h.notif, "assignment", auditID(assigment.ID), "create", "Alocare creată", "Alocare activă", map[string]interface{}{
		"machine_id":  assigment.MachineID,
		"operator_id": assigment.OperatorID,
		"status":      string(assigment.Status),
	})
}

// GetAll returneaza asignarile cu paginare si filtrare
// @Summary      Listare asignari
// @Tags         assignments
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Numar pagina" default(1)
// @Param        limit query int false "Rezultate per pagina" default(10)
// @Param        status query string false "Filtru status"
// @Param        operator_id query int false "Filtru operator"
// @Param        machine_id query int false "Filtru masina"
// @Param        sort_by query string false "Camp sortare" default(start_date)
// @Param        order query string false "Ordine" Enums(asc,desc) default(asc)
// @Success      200 {object} dto.PaginatedAssignmentsResponse
// @Router       /assignments [get]
func (h *AssigmentHandler) GetAll(c *gin.Context) {
	// Extrage parametrii de paginare și filtrare din query string
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}
	// Validare pentru pagina și limită
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	//safe
	if limit > 100 {
		limit = 100
	}
	log.Printf("Received request for GetAll with page=%d, limit=%d", page, limit)

	status := c.Query("status")
	operatorID := c.Query("operator_id")
	machineID := c.Query("machine_id")
	sortBy := c.DefaultQuery("sort_by", "start_date")
	order := c.DefaultQuery("order", "asc")

	query := dto.PaginationQuery{
		Page:       page,
		Limit:      limit,
		Status:     status,
		OperatorID: operatorID,
		MachineID:  machineID,
		SortBy:     sortBy,
		Order:      order,
	}

	result, err := h.service.GetAll(query)
	if err != nil {
		c.JSON(500, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID returneaza o asignare dupa ID
// @Summary      Obtinere asignare
// @Tags         assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID asignare"
// @Success      200 {object} domain.Assigment
// @Failure      404 {object} object{error=string}
// @Router       /assignments/{id} [get]
func (h *AssigmentHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
	assigment, err := h.service.GetAssigmentByID(assigmentID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assigment not found"})
		return
	}
	c.JSON(http.StatusOK, assigment)
}

// Update actualizeaza o asignare
// @Summary      Actualizare asignare
// @Tags         assignments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID asignare"
// @Param        body body CreateAssigmentRequest true "Date actualizare"
// @Success      200 {object} domain.Assigment
// @Failure      400 {object} object{error=string}
// @Router       /assignments/{id} [patch]
func (h *AssigmentHandler) Update(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
	assigment, err := h.service.GetAssigmentByID(assigmentID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Assigment not found"})
		return
	}

	var req CreateAssigmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldStatus := assigment.Status
	assigment.MachineID = req.MachineID
	assigment.OperatorID = req.OperatorID
	assigment.StartDate = req.StartDate
	assigment.EndDate = req.EndDate
	assigment.Status = req.Status

	updatedAssigment, err := h.service.UpdateAssigment(assigmentID, assigment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update assigment"})
		return
	}
	c.JSON(http.StatusOK, updatedAssigment)
	changes := map[string]interface{}{
		"machine_id":  updatedAssigment.MachineID,
		"operator_id": updatedAssigment.OperatorID,
		"status":      string(updatedAssigment.Status),
	}
	if oldStatus != updatedAssigment.Status {
		changes["old_status"] = string(oldStatus)
	}
	auditAndNotify(c, h.audit, h.notif, "assignment", id, "update", "Alocare actualizată", "Alocare actualizată", changes)
}

// Delete sterge o asignare
// @Summary      Stergere asignare
// @Tags         assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID asignare"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Router       /assignments/{id} [delete]
func (h *AssigmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
	assigment, _ := h.service.GetAssigmentByID(assigmentID)
	err = h.service.DeleteAssigment(assigmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assigment deleted successfully"})
	changes := map[string]interface{}{}
	if assigment != nil {
		changes["machine_id"] = assigment.MachineID
		changes["operator_id"] = assigment.OperatorID
	}
	auditAndNotify(c, h.audit, h.notif, "assignment", id, "delete", "Alocare ștearsă", "Alocare ștearsă", changes)
}

// Close inchide o asignare activa
// @Summary      Inchidere asignare
// @Tags         assignments
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID asignare"
// @Success      200 {object} object{message=string}
// @Failure      400 {object} object{error=string}
// @Router       /assignments/{id}/close [patch]
func (h *AssigmentHandler) Close(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
	assigment, _ := h.service.GetAssigmentByID(assigmentID)
	success, err := h.service.CloseAssigment(assigmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to close assigment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assigment closed successfully"})
	changes := statusChange("", string(domain.AssigmentStatusClosed))
	if assigment != nil {
		changes["old_status"] = string(assigment.Status)
		changes["machine_id"] = assigment.MachineID
		changes["operator_id"] = assigment.OperatorID
	}
	auditAndNotify(c, h.audit, h.notif, "assignment", id, "close", "Alocare închisă", "Alocare închisă", changes)
}
