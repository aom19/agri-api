package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type StockMovementHandler struct {
	service *usecase.StockMovementService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

func NewStockMovementHandler(service *usecase.StockMovementService, audit *usecase.AuditService, notif *usecase.NotificationService) *StockMovementHandler {
	return &StockMovementHandler{service: service, audit: audit, notif: notif}
}

type createStockMovementRequest struct {
	StockID          int64    `json:"stock_id" binding:"required"`
	MovementType     string   `json:"movement_type" binding:"required,oneof=in out adjustment"`
	Quantity         float64  `json:"quantity"`
	UnitCost         *float64 `json:"unit_cost"`
	FieldOperationID *int64   `json:"field_operation_id"`
	Notes            string   `json:"notes"`
}

// List returnează istoricul mișcărilor de stoc.
// @Summary      Mișcări de stoc
// @Tags         stocks
// @Produce      json
// @Security     BearerAuth
// @Param        stock_id query int false "Filtru stoc"
// @Param        resource_id query int false "Filtru resursă"
// @Param        field_operation_id query int false "Filtru operațiune pe teren"
// @Param        from query string false "De la (YYYY-MM-DD)"
// @Param        to query string false "Până la (YYYY-MM-DD, inclusiv)"
// @Param        limit query int false "Număr maxim (implicit 200, maxim 500)"
// @Success      200 {array} domain.StockMovement
// @Failure      400 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /stock-movements [get]
func (h *StockMovementHandler) List(c *gin.Context) {
	filter := domain.StockMovementFilter{}
	var err error
	if filter.StockID, err = optionalInt64Query(c.Query("stock_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stock_id invalid"})
		return
	}
	if filter.ResourceID, err = optionalInt64Query(c.Query("resource_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resource_id invalid"})
		return
	}
	if filter.FieldOperationID, err = optionalInt64Query(c.Query("field_operation_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field_operation_id invalid"})
		return
	}
	if limit, err := optionalInt64Query(c.Query("limit")); err == nil && limit > 0 {
		filter.Limit = int(limit)
	}
	if raw := c.Query("from"); raw != "" {
		parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from invalid (YYYY-MM-DD)"})
			return
		}
		filter.From = &parsed
	}
	if raw := c.Query("to"); raw != "" {
		parsed, err := time.ParseInLocation("2006-01-02", raw, time.UTC)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to invalid (YYYY-MM-DD)"})
			return
		}
		exclusive := parsed.AddDate(0, 0, 1)
		filter.To = &exclusive
	}

	items, err := h.service.List(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Create înregistrează o mișcare manuală de stoc (recepție, consum, ajustare).
// @Summary      Înregistrare mișcare de stoc
// @Tags         stocks
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body createStockMovementRequest true "Mișcarea de stoc"
// @Success      201 {object} domain.StockMovement
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Failure      500 {object} object{error=string}
// @Router       /stock-movements [post]
func (h *StockMovementHandler) Create(c *gin.Context) {
	var req createStockMovementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.service.Create(domain.StockMovementInput{
		StockID:          req.StockID,
		FieldOperationID: req.FieldOperationID,
		MovementType:     domain.StockMovementType(req.MovementType),
		Quantity:         req.Quantity,
		UnitCost:         req.UnitCost,
		Notes:            req.Notes,
		ActorID:          currentActorID(c),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrStockNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, usecase.ErrInvalidStockMovement):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	movement := result.Movement
	c.JSON(http.StatusCreated, movement)

	if h.audit != nil {
		h.audit.Log("stock", auditID(movement.StockID), "movement", currentActorID(c), map[string]interface{}{
			"movement_type":      string(movement.MovementType),
			"quantity_delta":     movement.QuantityDelta,
			"resulting_quantity": movement.ResultingQuantity,
			"notes":              movement.Notes,
		})
	}
	if h.notif != nil && result.BelowMinimum && movement.MovementType != domain.StockMovementIn {
		h.notif.Emit(
			domain.NotifStockLow,
			"Stoc scăzut",
			fmt.Sprintf("Stocul #%d a atins nivelul minim (%.2f / %.2f) după o mișcare de stoc", movement.StockID, movement.ResultingQuantity, result.Minimum),
			"stock", strconv.FormatInt(movement.StockID, 10),
		)
	}
}
