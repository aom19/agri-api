package handlers

import (
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// OperatorHandler gestionează request-urile HTTP pentru resursa operatori
type OperatorHandler struct {
	service *usecase.OperatorService
}

type CreateOperatorRequest struct {
	Name   string                `json:"name" binding:"required"`
	Status domain.OperatorStatus `json:"status"`
}

type UpdateOperatorRequest struct {
	Name   string                `json:"name" binding:"required"`
	Status domain.OperatorStatus `json:"status"`
}

// NewOperatorHandler creează un handler nou cu serviciul injectat
func NewOperatorHandler(service *usecase.OperatorService) *OperatorHandler {
	return &OperatorHandler{service: service}
}

// Create procesează POST /api/operators și creează un operator nou
func (h *OperatorHandler) Create(c *gin.Context) {
	var req CreateOperatorRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	operator, err := h.service.CreateOperator(&domain.Operator{
		Name:   req.Name,
		Status: req.Status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, operator)
}

// GetAll procesează GET /api/operators și returnează toți operatorii
func (h *OperatorHandler) GetAll(c *gin.Context) {
	operators, err := h.service.GetOperators()

	if err != nil {
		// return  empty array instead of error
		c.JSON(http.StatusNotFound, gin.H{"operators": []domain.Operator{}})
		return
	}
	c.JSON(http.StatusOK, operators)
}

// GetByID procesează GET /api/operators/:id și returnează operatorul cu ID-ul specificat
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

// Update procesează PUT /api/operators/:id și actualizează operatorul cu ID-ul specificat
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
	operator, err := h.service.UpdateOperator(operatorID, &domain.Operator{
		Name:   req.Name,
		Status: req.Status,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, operator)
}

// Delete procesează DELETE /api/operators/:id și șterge operatorul cu ID-ul specificat
func (h *OperatorHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	operatorID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operator ID"})
		return
	}
	_, err = h.service.GetOperatorByID(operatorID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "operator not found"})
		return
	}
	if err := h.service.DeleteOperator(operatorID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
