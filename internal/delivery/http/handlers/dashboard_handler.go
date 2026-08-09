package handlers

import (
	"net/http"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	service *usecase.DashboardService
}

func NewDashboardHandler(service *usecase.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

type DashboardCardsResponse struct {
	Cards []domain.DashboardCard `json:"cards"`
}

// GetCards returnează cardurile agregate pentru dashboard.
// @Summary      Carduri dashboard
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} DashboardCardsResponse
// @Failure      500 {object} object{error=string}
// @Router       /dashboard/cards [get]
func (h *DashboardHandler) GetCards(c *gin.Context) {
	cards, err := h.service.GetCards()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DashboardCardsResponse{Cards: cards})
}
