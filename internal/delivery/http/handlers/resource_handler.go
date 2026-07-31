package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ResourceHandler struct {
	service *usecase.ResourceService
}

func NewResourceHandler(service *usecase.ResourceService) *ResourceHandler {
	return &ResourceHandler{service: service}
}

type createResourceTypeRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Category    domain.ResourceCategory `json:"category" binding:"required"`
	DefaultUnit string                  `json:"default_unit" binding:"required"`
}

type updateResourceTypeRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Category    domain.ResourceCategory `json:"category" binding:"required"`
	DefaultUnit string                  `json:"default_unit" binding:"required"`
}

type createResourceRequest struct {
	Name           string  `json:"name" binding:"required"`
	ResourceTypeID int64   `json:"resource_type_id" binding:"required"`
	PricePerUnit   float64 `json:"price_per_unit" binding:"required"`
	Notes          string  `json:"notes"`
}

type updateResourceRequest struct {
	Name           string  `json:"name" binding:"required"`
	ResourceTypeID int64   `json:"resource_type_id" binding:"required"`
	PricePerUnit   float64 `json:"price_per_unit" binding:"required"`
	Notes          string  `json:"notes"`
}

func (h *ResourceHandler) GetAllResourceTypes(c *gin.Context) {
	items, err := h.service.GetResourceTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ResourceHandler) GetResourceTypeByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type id"})
		return
	}

	item, err := h.service.GetResourceTypeByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource type not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ResourceHandler) CreateResourceType(c *gin.Context) {
	var req createResourceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.CreateResourceType(&domain.ResourceType{
		Name:        req.Name,
		Category:    req.Category,
		DefaultUnit: req.DefaultUnit,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *ResourceHandler) UpdateResourceType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type id"})
		return
	}

	var req updateResourceTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateResourceType(id, &domain.ResourceType{
		Name:        req.Name,
		Category:    req.Category,
		DefaultUnit: req.DefaultUnit,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrResourceTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *ResourceHandler) DeleteResourceType(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource type id"})
		return
	}

	if err := h.service.DeleteResourceType(id); err != nil {
		if errors.Is(err, usecase.ErrResourceTypeNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ResourceHandler) GetAllResources(c *gin.Context) {
	items, err := h.service.GetResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *ResourceHandler) GetResourceByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	item, err := h.service.GetResourceByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *ResourceHandler) CreateResource(c *gin.Context) {
	var req createResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.service.CreateResource(&domain.Resource{
		Name:           req.Name,
		ResourceTypeID: req.ResourceTypeID,
		PricePerUnit:   req.PricePerUnit,
		Notes:          req.Notes,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	var req updateResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateResource(id, &domain.Resource{
		Name:           req.Name,
		ResourceTypeID: req.ResourceTypeID,
		PricePerUnit:   req.PricePerUnit,
		Notes:          req.Notes,
	})
	if err != nil {
		if errors.Is(err, usecase.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource id"})
		return
	}

	if err := h.service.DeleteResource(id); err != nil {
		if errors.Is(err, usecase.ErrResourceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
