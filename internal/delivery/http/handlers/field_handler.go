package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/usecase"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// FieldHandler gestioneaza request-urile HTTP pentru terenuri.
type FieldHandler struct {
	service *usecase.FieldService
	audit   *usecase.AuditService
	notif   *usecase.NotificationService
}

func NewFieldHandler(service *usecase.FieldService, opts ...func(*FieldHandler)) *FieldHandler {
	h := &FieldHandler{service: service}
	for _, o := range opts {
		o(h)
	}
	return h
}

func WithFieldAudit(a *usecase.AuditService) func(*FieldHandler) {
	return func(h *FieldHandler) { h.audit = a }
}

func WithFieldNotif(n *usecase.NotificationService) func(*FieldHandler) {
	return func(h *FieldHandler) { h.notif = n }
}

type createFieldRequest struct {
	Name            string          `json:"name" binding:"required"`
	CadastralNumber *string         `json:"cadastral_number"`
	AreaHa          *float64        `json:"area_ha"`
	Geometry        json.RawMessage `json:"geometry" binding:"required"`
}

type updateFieldRequest struct {
	Name            string          `json:"name" binding:"required"`
	CadastralNumber *string         `json:"cadastral_number"`
	AreaHa          *float64        `json:"area_ha"`
	Geometry        json.RawMessage `json:"geometry" binding:"required"`
}

func (h *FieldHandler) Create(c *gin.Context) {
	var req createFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field, err := h.service.CreateField(&domain.Field{
		Name:            req.Name,
		CadastralNumber: req.CadastralNumber,
		AreaHa:          req.AreaHa,
		Geometry:        req.Geometry,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, field)
	auditAndNotify(c, h.audit, h.notif, "field", field.ID, "create", "Teren creat", field.Name, map[string]interface{}{
		"name":    field.Name,
		"area_ha": field.AreaHa,
	})
}

// GetAll listează terenurile; cu page, o pagină cu totalul și suprafața totală filtrată.
// @Summary      Listare terenuri
// @Tags         fields
// @Produce      json
// @Security     BearerAuth
// @Param        q query string false "Nume sau număr cadastral"
// @Param        min_area query number false "Suprafață minimă (ha)"
// @Param        max_area query number false "Suprafață maximă (ha)"
// @Param        sort query string false "name (implicit) sau area"
// @Param        order query string false "asc (implicit) sau desc"
// @Param        page query int false "Pagina (de la 1); fără ea, lista întreagă ca tablou"
// @Param        page_size query int false "Rânduri pe pagină (implicit 25, maxim 200)"
// @Success      200 {object} object{items=[]domain.Field,total=int,page=int,page_size=int,total_area_ha=number} "Cu page: pagina; fără page: tabloul domain.Field"
// @Failure      400 {object} object{error=string}
// @Router       /fields [get]
func (h *FieldHandler) GetAll(c *gin.Context) {
	search, page, ok := listRequest(c)
	if !ok {
		return
	}
	filter := domain.FieldFilter{Query: search.Query, SortBy: c.Query("sort"), SortDesc: c.Query("order") == "desc"}
	for name, target := range map[string]**float64{"min_area": &filter.MinArea, "max_area": &filter.MaxArea} {
		raw := strings.TrimSpace(c.Query(name))
		if raw == "" {
			continue
		}
		value, err := strconv.ParseFloat(raw, 64)
		if err != nil || value < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": name + " trebuie să fie un număr pozitiv"})
			return
		}
		*target = &value
	}
	fields, total, totalArea, err := h.service.ListFields(filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if page == nil {
		respondList(c, fields, total, nil)
		return
	}
	c.JSON(http.StatusOK, domain.FieldPage{Page: domain.NewPage(fields, total, *page), TotalAreaHa: totalArea})
}

func (h *FieldHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	field, err := h.service.GetFieldByID(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if field == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "field not found"})
		return
	}
	c.JSON(http.StatusOK, field)
}

func (h *FieldHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req updateFieldRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateField(id, &domain.Field{
		Name:            req.Name,
		CadastralNumber: req.CadastralNumber,
		AreaHa:          req.AreaHa,
		Geometry:        req.Geometry,
	})
	if err != nil {
		if err.Error() == "field not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
	auditAndNotify(c, h.audit, h.notif, "field", updated.ID, "update", "Teren actualizat", updated.Name, map[string]interface{}{
		"name":    updated.Name,
		"area_ha": updated.AreaHa,
	})
}

func (h *FieldHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	field, _ := h.service.GetFieldByID(id)
	if err := h.service.DeleteField(id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
	name := "teren"
	if field != nil {
		name = field.Name
	}
	auditAndNotify(c, h.audit, h.notif, "field", id, "delete", "Teren șters", name, nil)
}
