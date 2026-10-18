package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"agri-api/internal/domain"

	"github.com/gin-gonic/gin"
)

// pageRequest citește `page` (de la 1) și `page_size` din query. Fără `page`, lista nu e
// paginată și răspunsul rămâne tabloul simplu (dropdown-uri, hărți).
func pageRequest(c *gin.Context) (*domain.PageRequest, error) {
	rawPage := strings.TrimSpace(c.Query("page"))
	if rawPage == "" {
		return nil, nil
	}
	page, err := strconv.Atoi(rawPage)
	if err != nil || page < 1 {
		return nil, fmt.Errorf("page trebuie să fie un număr întreg de la 1")
	}
	size := domain.DefaultPageSize
	if rawSize := strings.TrimSpace(c.Query("page_size")); rawSize != "" {
		size, err = strconv.Atoi(rawSize)
		if err != nil || size < 1 || size > domain.MaxPageSize {
			return nil, fmt.Errorf("page_size trebuie să fie între 1 și %d", domain.MaxPageSize)
		}
	}
	return &domain.PageRequest{Page: page, PageSize: size}, nil
}

// respondList scrie lista: tabloul simplu, sau pagina {items, total, page, page_size} când
// clientul a cerut paginare.
func respondList[T any](c *gin.Context, items []T, total int, page *domain.PageRequest) {
	if page == nil {
		if items == nil {
			items = []T{}
		}
		c.JSON(http.StatusOK, items)
		return
	}
	c.JSON(http.StatusOK, domain.NewPage(items, total, *page))
}

// listRequest citește paginarea și căutarea `q`; la o paginare invalidă răspunde 400 și întoarce ok=false.
func listRequest(c *gin.Context) (domain.ListFilter, *domain.PageRequest, bool) {
	page, err := pageRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return domain.ListFilter{}, nil, false
	}
	return domain.ListFilter{Query: c.Query("q")}, page, true
}
