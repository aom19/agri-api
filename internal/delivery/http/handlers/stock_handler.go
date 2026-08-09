package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	service *usecase.StockService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

func NewStockHandler(service *usecase.StockService, opts ...func(*StockHandler)) *StockHandler {
	h := &StockHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithStockAudit(a *usecase.AuditService) func(*StockHandler) {
	return func(h *StockHandler) { h.audit = a }
}

func WithStockNotif(n *usecase.NotificationService) func(*StockHandler) {
	return func(h *StockHandler) { h.notif = n }
}

type stockRequest struct {
	ResourceID      int64   `json:"resource_id" binding:"gt=0"`
	Quantity        float64 `json:"quantity" binding:"gte=0"`
	MinimumQuantity float64 `json:"minimum_quantity" binding:"gte=0"`
}

func (h *StockHandler) GetAll(c *gin.Context) {
	stocks, err := h.service.GetStocks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stocks)
}

func (h *StockHandler) GetByID(c *gin.Context) {
	id, err := parseStockID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock id"})
		return
	}

	stock, err := h.service.GetStockByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if stock == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "stock not found"})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *StockHandler) Create(c *gin.Context) {
	var req stockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stock, err := h.service.CreateStock(&domain.Stock{
		ResourceID:      req.ResourceID,
		Quantity:        req.Quantity,
		MinimumQuantity: req.MinimumQuantity,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, stock)

	auditAndNotify(c, h.audit, h.notif, "stock", auditID(stock.ID), "create", "Stoc creat", fmt.Sprintf("Stoc #%d", stock.ID), map[string]interface{}{
		"resource_id":      stock.ResourceID,
		"quantity":         stock.Quantity,
		"minimum_quantity": stock.MinimumQuantity,
	})
}

func (h *StockHandler) Update(c *gin.Context) {
	id, err := parseStockID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock id"})
		return
	}

	var req stockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	oldStock, _ := h.service.GetStockByID(id)
	stock, err := h.service.UpdateStock(id, &domain.Stock{
		ResourceID:      req.ResourceID,
		Quantity:        req.Quantity,
		MinimumQuantity: req.MinimumQuantity,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrStockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)

	changes := map[string]interface{}{
		"resource_id":      stock.ResourceID,
		"quantity":         stock.Quantity,
		"minimum_quantity": stock.MinimumQuantity,
	}
	if oldStock != nil {
		changes["old_quantity"] = oldStock.Quantity
		changes["old_minimum_quantity"] = oldStock.MinimumQuantity
	}
	auditAndNotify(c, h.audit, h.notif, "stock", auditID(id), "update", "Stoc actualizat", fmt.Sprintf("Stoc #%d", id), changes)
	if h.notif != nil && stock.Quantity <= stock.MinimumQuantity {
		h.notif.Emit(
			domain.NotifStockLow,
			"Stoc scăzut",
			fmt.Sprintf("Stocul #%d a atins nivelul minim (%.2f / %.2f)", id, stock.Quantity, stock.MinimumQuantity),
			"stock", fmt.Sprintf("%d", id),
		)
	}
}

func (h *StockHandler) Delete(c *gin.Context) {
	id, err := parseStockID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock id"})
		return
	}

	stock, _ := h.service.GetStockByID(id)
	if err := h.service.DeleteStock(id); err != nil {
		if errors.Is(err, usecase.ErrStockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
	message := fmt.Sprintf("Stoc #%d", id)
	if stock != nil && stock.Resource != nil {
		message = stock.Resource.Name
	}
	auditAndNotify(c, h.audit, h.notif, "stock", auditID(id), "delete", "Stoc șters", message, nil)
}

func parseStockID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
