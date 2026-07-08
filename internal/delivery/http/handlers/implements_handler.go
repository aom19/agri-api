package handlers 

import (
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// ImplementHandler gestionează request-urile HTTP pentru resursa utilaje agricole
type ImplementHandler struct {
	service *usecase.ImplementService
}

// NewImplementHandler creează un handler nou cu serviciul injectat
func NewImplementHandler(service *usecase.ImplementService) *ImplementHandler {
	return &ImplementHandler{service: service}
}

type CreateImplementRequest struct {
	Name               string             `json:"name" binding:"required"`
	Code               string             `json:"code" binding:"required"`
	Type               domain.ImplementType `json:"type" binding:"required"`
	Brand              string             `json:"brand"`
	Model              string             `json:"model"`
	Year               *int               `json:"year"`
	RegistrationNumber string             `json:"registration_number"`
	FuelType           *domain.FuelType   `json:"fuel_type"`
	Notes              string             `json:"notes"`
	WorkingWidth	  *float64           `json:"working_width"`
	Capacity          *float64           `json:"capacity"`
}
type UpdateImplementRequest struct {
	Name               string               `json:"name" binding:"required"`
	Code               string               `json:"code" binding:"required"`
	Type               domain.ImplementType `json:"type" binding:"required"`
	Brand              string               `json:"brand"`
	Model              string               `json:"model"`
	Year               *int                 `json:"year"`
	RegistrationNumber string               `json:"registration_number"`
	FuelType           *domain.FuelType     `json:"fuel_type"`
	Status             domain.AssetStatus   `json:"status" binding:"required"`
	Notes              string               `json:"notes"`
	WorkingWidth       *float64             `json:"working_width"`
	Capacity           *float64             `json:"capacity"`	
}



// Create creează un utilaj agricol nou
// @Summary      Creare utilaj agricol
// @Tags         implements
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body CreateImplementRequest true "Date utilaj agricol"
// @Success      201 {object} domain.Implement
// @Failure      400 {object} object{error=string}
// @Router       /implements [post]
func (h *ImplementHandler) Create(c *gin.Context) {
	var req CreateImplementRequest	
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	implement, err := h.service.CreateImplement(&domain.Implement{
		Name:         req.Name,
		Code:         req.Code,
		Type:         req.Type,
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		WorkingWidth: req.WorkingWidth,
		Capacity:     req.Capacity,
		Status:       domain.AssetStatusActive, //default active
		Notes:        req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, implement)
}

func (h *ImplementHandler) Update(c *gin.Context) {
	var req UpdateImplementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id tehnicii agricole invalid"})
		return
	}

	implement, err := h.service.UpdateImplement(id, &domain.Implement{
		Name:         req.Name,
		Code:         req.Code,
		Type:         req.Type,
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		Status:       req.Status,
		Notes:        req.Notes,
		WorkingWidth: req.WorkingWidth,
		Capacity:     req.Capacity,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, implement)
}	



// GetByID returnează un utilaj agricol după ID
// @Summary      Obținere utilaj agricol
// @Tags         implements
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID utilaj agricol"
// @Success      200 {object} domain.Implement
// @Failure      404 {object} object{error=string}
// @Router       /implements/{id} [get]
func (h *ImplementHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id tehnicii agricole invalid"})
		return
	}

	implement, err := h.service.GetImplementByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if implement == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tehnica agricola nu a fost gasita"})
		return
	}		

	c.JSON(http.StatusOK, implement)
}

// GetAll returnează toate utilajele agricole
// @Summary      Listare utilaje agricole
// @Tags         implements
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Implement
// @Router       /implements [get]
func (h *ImplementHandler) GetAll(c *gin.Context) {
	implements, err := h.service.GetImplements()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, implements)
}	

// Delete șterge un utilaj agricol existent
// @Summary      Ștergere utilaj agricol
// @Tags         implements	
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID utilaj agricol"
// @Success      204 {object} object{}
// @Failure      404 {object} object{error=string}
// @Router       /implements/{id} [delete]
func (h *ImplementHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id tehnicii agricole invalid"})
		return
	}

	err = h.service.DeleteImplement(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.Status(http.StatusNoContent)
}


// Activate activează un utilaj agricol existent
// @Summary      Activare utilaj agricol
// @Tags         implements
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID utilaj agricol"
// @Success      200 {object} domain.Implement
// @Failure      404 {object} object{error=string}
// @Router       /implements/{id}/activate [patch]
func (h *ImplementHandler) Activate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id tehnicii agricole invalid"})
		return
	}

	err = h.service.ActivateImplement(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tehnica agricola activata cu succes"})
}

// Deactivate dezactivează un utilaj agricol existent
// @Summary      Dezactivare utilaj agricol
// @Tags         implements
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID utilaj agricol"
// @Success      200 {object} domain.Implement
// @Failure      404 {object} object{error=string}
// @Router       /implements/{id}/deactivate [patch]
func (h *ImplementHandler) Deactivate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Id tehnicii agricole invalid"})
		return
	}

	err = h.service.DeactivateImplement(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tehnica agricola dezactivata cu succes"})
}