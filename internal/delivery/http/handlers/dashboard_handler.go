package handlers

import (
	"net/http"
	"strconv"

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

// GetQuickStats returnează indicatorii rapizi pentru dashboard (terenuri active, lucrări întârziate etc.).
// @Summary      Statistici rapide dashboard
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} domain.DashboardQuickStats
// @Failure      500 {object} object{error=string}
// @Router       /dashboard/quick-stats [get]
func (h *DashboardHandler) GetQuickStats(c *gin.Context) {
	stats, err := h.service.GetQuickStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetActivity returnează activitatea recentă (din jurnalul de audit) pentru dashboard.
// @Summary      Activitate recentă dashboard
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        limit query int false "Număr maxim de intrări (implicit 5, maxim 50)"
// @Success      200 {array} domain.DashboardActivityItem
// @Failure      500 {object} object{error=string}
// @Router       /dashboard/activity [get]
func (h *DashboardHandler) GetActivity(c *gin.Context) {
	limit := 5
	if parsed, err := strconv.Atoi(c.Query("limit")); err == nil && parsed > 0 {
		limit = parsed
	}

	items, err := h.service.GetRecentActivity(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}
