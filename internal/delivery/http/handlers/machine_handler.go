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

// Create creează o mașină nouă
// @Summary      Creare mașină
// @Tags         machines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateMachineRequest true "Date mașină"
// @Success      201 {object} domain.Machine
// @Failure      400 {object} object{error=string}
// @Router       /machines [post]
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

// GetAll returnează toate mașinile
// @Summary      Listare mașini
// @Tags         machines
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Machine
// @Router       /machines [get]
func (h *MachineHandler) GetAll(c *gin.Context) {
	machines, err := h.service.GetMachines()

	if err != nil {
		// return  empty array instead of error
		c.JSON(http.StatusNotFound, gin.H{"machines": []domain.Machine{}})
		return
	}
	c.JSON(http.StatusOK, machines)
}

// GetByID returnează o mașină după ID
// @Summary      Obținere mașină
// @Tags         machines
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID mașină"
// @Success      200 {object} domain.Machine
// @Failure      404 {object} object{error=string}
// @Router       /machines/{id} [get]
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

// Update actualizează o mașină
// @Summary      Actualizare mașină
// @Tags         machines
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID mașină"
// @Param        body body UpdateMachineRequest true "Date actualizare"
// @Success      200 {object} domain.Machine
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /machines/{id} [patch]
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

// Delete șterge o mașină
// @Summary      Ștergere mașină
// @Tags         machines
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID mașină"
// @Success      200 {object} object{message=string}
// @Failure      500 {object} object{error=string}
// @Router       /machines/{id} [delete]
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
