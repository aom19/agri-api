package handlers

import (
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// MachineHandler gestionează request-urile HTTP pentru resursa mașini
type MachineHandler struct {
	service *usecase.MachineService
}

// NewMachineHandler creează un handler nou cu serviciul injectat
func NewMachineHandler(service *usecase.MachineService) *MachineHandler {
	return &MachineHandler{service: service}
}

type CreateMachineRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
}
type UpdateMachineRequest struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" `
	Description string `json:"description"`
	Status      string `json:"status" `
}

// Create procesează POST /api/machines și creează o mașină nouă
func (h *MachineHandler) Create(c *gin.Context) {
	var req CreateMachineRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	machine, err := h.service.CreateMachine(&domain.Machine{
		Name:        req.Name,
		Type:        req.Type,
		Description: req.Description,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, machine)
}

// GetAll procesează GET /api/machines și returnează toate mașinile
func (h *MachineHandler) GetAll(c *gin.Context) {
	machines, err := h.service.GetMachines()

	if err != nil {
		// return  empty array instead of error
		c.JSON(http.StatusNotFound, gin.H{"machines": []domain.Machine{}})
		return
	}
	c.JSON(http.StatusOK, machines)
}

// GetByID procesează GET /api/machines/:id și returnează o mașină după ID
func (h *MachineHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	machineID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid machine ID"})
		return
	}
	machine, err := h.service.GetMachineByID(machineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if machine == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Machine not found"})
		return
	}
	c.JSON(http.StatusOK, machine)
}

// Update procesează PATCH /api/machines/:id și actualizează datele mașinii
func (h *MachineHandler) Update(c *gin.Context) {
	id := c.Param("id")
	machineID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid machine ID"})
		return
	}
	machine, err := h.service.GetMachineByID(machineID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Machine not found"})
		return
	}

	var req UpdateMachineRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	machine.Name = req.Name
	machine.Type = req.Type
	machine.Description = req.Description

	updatedMachine, err := h.service.UpdateMachine(machineID, machine)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update machine"})
		return
	}
	c.JSON(http.StatusOK, updatedMachine)
}

func (h *MachineHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	machineID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid machine ID"})
		return
	}
	err = h.service.DeleteMachine(machineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete machine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Machine deleted successfully"})
}
