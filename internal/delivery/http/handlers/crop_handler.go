package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

type CropHandler struct {
	service *usecase.CropService
	audit   *usecase.AuditService
}

func NewCropHandler(service *usecase.CropService, audit *usecase.AuditService) *CropHandler {
	return &CropHandler{service: service, audit: audit}
}

type seasonRequest struct {
	Name      string `json:"name" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	IsActive  bool   `json:"is_active"`
	Notes     string `json:"notes"`
}

type cropRequest struct {
	Name      string  `json:"name" binding:"required"`
	Code      *string `json:"code"`
	Category  string  `json:"category"`
	YieldUnit string  `json:"yield_unit"`
	Notes     string  `json:"notes"`
}

type fieldCropRequest struct {
	FieldID            string   `json:"field_id" binding:"required"`
	SeasonID           int64    `json:"season_id" binding:"required"`
	CropID             int64    `json:"crop_id" binding:"required"`
	PlantedAreaHa      *float64 `json:"planted_area_ha"`
	PlantedAt          *string  `json:"planted_at"`
	HarvestedAt        *string  `json:"harvested_at"`
	ProductionTotal    *float64 `json:"production_total"`
	ExpectedYieldPerHa *float64 `json:"expected_yield_per_ha"`
	Notes              string   `json:"notes"`
}

func respondCropError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, usecase.ErrCropNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, usecase.ErrCropDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, usecase.ErrCropInvalid), errors.Is(err, usecase.ErrHarvestNotRecordable):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func pathID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func (h *CropHandler) log(c *gin.Context, entityType string, id int64, action string, changes map[string]interface{}) {
	if h.audit != nil {
		h.audit.Log(entityType, auditID(id), action, currentActorID(c), changes)
	}
}

// ─── Seasons ─────────────────────────────────────────────────────────────────

// GetSeasons listează sezoanele agricole.
// @Summary      Sezoane
// @Tags         crops
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Season
// @Router       /seasons [get]
func (h *CropHandler) GetSeasons(c *gin.Context) {
	items, err := h.service.GetSeasons()
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateSeason creează un sezon; dacă este marcat activ, celelalte devin inactive.
// @Summary      Creare sezon
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body seasonRequest true "Sezonul"
// @Success      201 {object} domain.Season
// @Failure      400 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Router       /seasons [post]
func (h *CropHandler) CreateSeason(c *gin.Context) {
	var req seasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	season, err := h.service.CreateSeason(&domain.Season{Name: req.Name, StartDate: req.StartDate, EndDate: req.EndDate, IsActive: req.IsActive, Notes: req.Notes})
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusCreated, season)
	h.log(c, "season", season.ID, "create", map[string]interface{}{"name": season.Name, "start_date": season.StartDate, "end_date": season.EndDate})
}

// UpdateSeason actualizează un sezon.
// @Summary      Actualizare sezon
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID sezon"
// @Param        payload body seasonRequest true "Sezonul"
// @Success      200 {object} domain.Season
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /seasons/{id} [patch]
func (h *CropHandler) UpdateSeason(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req seasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	season, err := h.service.UpdateSeason(id, &domain.Season{Name: req.Name, StartDate: req.StartDate, EndDate: req.EndDate, IsActive: req.IsActive, Notes: req.Notes})
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, season)
	h.log(c, "season", id, "update", map[string]interface{}{"name": season.Name, "is_active": season.IsActive})
}

// DeleteSeason șterge un sezon și culturile pe terenuri asociate.
// @Summary      Ștergere sezon
// @Tags         crops
// @Security     BearerAuth
// @Param        id path int true "ID sezon"
// @Success      204
// @Failure      404 {object} object{error=string}
// @Router       /seasons/{id} [delete]
func (h *CropHandler) DeleteSeason(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteSeason(id); err != nil {
		respondCropError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	h.log(c, "season", id, "delete", nil)
}

// ─── Crops ───────────────────────────────────────────────────────────────────

// GetCrops listează catalogul de culturi.
// @Summary      Culturi
// @Tags         crops
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} domain.Crop
// @Router       /crops [get]
func (h *CropHandler) GetCrops(c *gin.Context) {
	items, err := h.service.GetCrops()
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateCrop adaugă o cultură în catalog.
// @Summary      Creare cultură
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body cropRequest true "Cultura"
// @Success      201 {object} domain.Crop
// @Failure      400 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Router       /crops [post]
func (h *CropHandler) CreateCrop(c *gin.Context) {
	var req cropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	crop, err := h.service.CreateCrop(&domain.Crop{Name: req.Name, Code: req.Code, Category: req.Category, YieldUnit: req.YieldUnit, Notes: req.Notes})
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusCreated, crop)
	h.log(c, "crop", crop.ID, "create", map[string]interface{}{"name": crop.Name})
}

// UpdateCrop actualizează o cultură.
// @Summary      Actualizare cultură
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID cultură"
// @Param        payload body cropRequest true "Cultura"
// @Success      200 {object} domain.Crop
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /crops/{id} [patch]
func (h *CropHandler) UpdateCrop(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req cropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	crop, err := h.service.UpdateCrop(id, &domain.Crop{Name: req.Name, Code: req.Code, Category: req.Category, YieldUnit: req.YieldUnit, Notes: req.Notes})
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, crop)
	h.log(c, "crop", id, "update", map[string]interface{}{"name": crop.Name})
}

// DeleteCrop șterge o cultură nefolosită.
// @Summary      Ștergere cultură
// @Tags         crops
// @Security     BearerAuth
// @Param        id path int true "ID cultură"
// @Success      204
// @Failure      404 {object} object{error=string}
// @Router       /crops/{id} [delete]
func (h *CropHandler) DeleteCrop(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteCrop(id); err != nil {
		if errors.Is(err, usecase.ErrCropNotFound) {
			respondCropError(c, err)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
	h.log(c, "crop", id, "delete", nil)
}

// ─── Field crops ─────────────────────────────────────────────────────────────

// ListFieldCrops listează culturile pe terenuri, filtrabil pe sezon, cultură și teren.
// @Summary      Culturi pe terenuri
// @Tags         crops
// @Produce      json
// @Security     BearerAuth
// @Param        season_id query int false "Filtru sezon"
// @Param        crop_id query int false "Filtru cultură"
// @Param        field_id query string false "Filtru teren"
// @Success      200 {array} domain.FieldCrop
// @Router       /field-crops [get]
func (h *CropHandler) ListFieldCrops(c *gin.Context) {
	filter := domain.FieldCropFilter{FieldID: c.Query("field_id")}
	var err error
	if filter.SeasonID, err = optionalInt64Query(c.Query("season_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "season_id invalid"})
		return
	}
	if filter.CropID, err = optionalInt64Query(c.Query("crop_id")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "crop_id invalid"})
		return
	}
	items, err := h.service.ListFieldCrops(filter)
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, items)
}

func toDomainFieldCrop(req fieldCropRequest) *domain.FieldCrop {
	return &domain.FieldCrop{
		FieldID:            req.FieldID,
		SeasonID:           req.SeasonID,
		CropID:             req.CropID,
		PlantedAreaHa:      req.PlantedAreaHa,
		PlantedAt:          req.PlantedAt,
		HarvestedAt:        req.HarvestedAt,
		ProductionTotal:    req.ProductionTotal,
		ExpectedYieldPerHa: req.ExpectedYieldPerHa,
		Notes:              req.Notes,
	}
}

// CreateFieldCrop atribuie o cultură unui teren într-un sezon.
// @Summary      Creare cultură pe teren
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        payload body fieldCropRequest true "Cultura pe teren"
// @Success      201 {object} domain.FieldCrop
// @Failure      400 {object} object{error=string}
// @Failure      409 {object} object{error=string}
// @Router       /field-crops [post]
func (h *CropHandler) CreateFieldCrop(c *gin.Context) {
	var req fieldCropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.CreateFieldCrop(toDomainFieldCrop(req))
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusCreated, item)
	h.log(c, "field_crop", item.ID, "create", map[string]interface{}{"field": item.FieldName, "crop": item.CropName, "season": item.SeasonName})
}

// UpdateFieldCrop actualizează o cultură pe teren (inclusiv producția obținută).
// @Summary      Actualizare cultură pe teren
// @Tags         crops
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID"
// @Param        payload body fieldCropRequest true "Cultura pe teren"
// @Success      200 {object} domain.FieldCrop
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /field-crops/{id} [patch]
func (h *CropHandler) UpdateFieldCrop(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req fieldCropRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	item, err := h.service.UpdateFieldCrop(id, toDomainFieldCrop(req))
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
	h.log(c, "field_crop", id, "update", map[string]interface{}{"field": item.FieldName, "crop": item.CropName, "production_total": item.ProductionTotal})
}

// DeleteFieldCrop șterge o cultură pe teren.
// @Summary      Ștergere cultură pe teren
// @Tags         crops
// @Security     BearerAuth
// @Param        id path int true "ID"
// @Success      204
// @Failure      404 {object} object{error=string}
// @Router       /field-crops/{id} [delete]
func (h *CropHandler) DeleteFieldCrop(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteFieldCrop(id); err != nil {
		respondCropError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
	h.log(c, "field_crop", id, "delete", nil)
}

type harvestResponse struct {
	FieldCrop *domain.FieldCrop     `json:"field_crop"`
	Movement  *domain.StockMovement `json:"movement,omitempty"`
}

// RecordHarvest înregistrează producția obținută ca intrare în stoc (resursa de recoltă a culturii).
// @Summary      Recoltă în stoc
// @Tags         crops
// @Produce      json
// @Security     BearerAuth
// @Param        id path int true "ID cultură pe teren"
// @Success      200 {object} harvestResponse
// @Failure      400 {object} object{error=string}
// @Failure      404 {object} object{error=string}
// @Router       /field-crops/{id}/harvest [post]
func (h *CropHandler) RecordHarvest(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	result, err := h.service.RecordHarvest(id, currentActorID(c))
	if err != nil {
		respondCropError(c, err)
		return
	}
	c.JSON(http.StatusOK, harvestResponse{FieldCrop: result.FieldCrop, Movement: result.Movement})
	if result.Movement != nil {
		h.log(c, "field_crop", id, "harvest", map[string]interface{}{
			"crop":           result.FieldCrop.CropName,
			"field":          result.FieldCrop.FieldName,
			"quantity_delta": result.Movement.QuantityDelta,
			"stock_id":       result.Movement.StockID,
		})
	}
}
