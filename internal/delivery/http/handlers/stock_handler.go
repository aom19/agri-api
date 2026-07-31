package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StockHandler struct {
	service *usecase.StockService
}

func NewStockHandler(service *usecase.StockService) *StockHandler {
	return &StockHandler{service: service}
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
}

func (h *StockHandler) Delete(c *gin.Context) {
	id, err := parseStockID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock id"})
		return
	}

	if err := h.service.DeleteStock(id); err != nil {
		if errors.Is(err, usecase.ErrStockNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseStockID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
