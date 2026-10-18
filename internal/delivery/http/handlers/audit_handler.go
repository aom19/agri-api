package handlers

import (
	"agri-api/internal/domain"
	"agri-api/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	repo repository.AuditRepository
}

func NewAuditHandler(repo repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// GetAll listează jurnalul de audit, cele mai noi intrări primele.
// @Summary      Jurnal de audit
// @Tags         audit
// @Produce      json
// @Security     BearerAuth
// @Param        entity_type query string false "Tipul entității"
// @Param        entity_id query string false "Id-ul entității"
// @Param        limit query int false "Fără page: numărul de intrări (implicit 100, maxim 500)"
// @Param        page query int false "Pagina (de la 1)"
// @Param        page_size query int false "Rânduri pe pagină (implicit 25, maxim 200)"
// @Success      200 {object} object{items=[]domain.AuditEntry,total=int,page=int,page_size=int} "Cu page: pagina; fără page: tabloul domain.AuditEntry"
// @Failure      400 {object} object{error=string}
// @Router       /audit-log [get]
func (h *AuditHandler) GetAll(c *gin.Context) {
	page, err := pageRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter := domain.AuditFilter{EntityType: c.Query("entity_type"), EntityID: c.Query("entity_id")}
	if page == nil {
		limit := 100
		if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 && l <= 500 {
			limit = l
		}
		entries, err := h.repo.GetAll(limit, filter.EntityType, filter.EntityID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		respondList(c, entries, len(entries), nil)
		return
	}
	entries, total, err := h.repo.List(filter, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	respondList(c, entries, total, page)
}
