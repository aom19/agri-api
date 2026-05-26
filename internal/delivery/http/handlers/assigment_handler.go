package handlers

import (
	"net/http"
	"strconv"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// AssigmentHandler gestionează request-urile HTTP pentru resursa asigment
type AssigmentHandler struct {
	service *usecase.AssigmentService
}

// NewAssigmentHandler creează un handler nou cu serviciul injectat
func NewAssigmentHandler(service *usecase.AssigmentService) *AssigmentHandler {
	return &AssigmentHandler{service: service}
}

type CreateAssigmentRequest struct {
	MachineID  int64                  `json:"machine_id" binding:"required"`
	OperatorID int64                  `json:"operator_id" binding:"required"`
	StartDate  time.Time              `json:"start_date" binding:"required"`
	EndDate    *time.Time             `json:"end_date" binding:"required"`
	Status     domain.AssigmentStatus `json:"status" binding:"required"`
}

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
}

func (h *AssigmentHandler) GetAll(c *gin.Context) {
	assigments, err := h.service.GetAssigments()

	if err != nil {
		// return  empty array instead of error
		c.JSON(http.StatusOK, []domain.Assigment{})
		return
	}
	c.JSON(http.StatusOK, assigments)
}

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
}

func (h *AssigmentHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
	err = h.service.DeleteAssigment(assigmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Assigment deleted successfully"})
}

func (h *AssigmentHandler) Close(c *gin.Context) {
	id := c.Param("id")
	assigmentID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid assigment ID"})
		return
	}
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
}
