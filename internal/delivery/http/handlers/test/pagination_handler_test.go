package handlers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"agri-api/internal/domain"
)

// Paginarea pe server: fără `page`, lista rămâne tabloul simplu (dropdown-uri, hărți); cu `page`,
// răspunsul e pagina {items, total, page, page_size}, iar căutarea `q` ajunge la repository.
func TestListHandlers_Pagination(t *testing.T) {
	var gotFilter domain.ListFilter
	var gotPage *domain.PageRequest
	repo := &machineRepoMock{list: func(f domain.ListFilter, p *domain.PageRequest) ([]domain.Machine, int, error) {
		gotFilter, gotPage = f, p
		return []domain.Machine{{ID: 3, Name: "T3"}}, 41, nil
	}}
	router := newMachineRouter(repo, &auditRepoMock{}, &notificationRepoMock{})

	rec := do(t, router, http.MethodGet, "/machines?q=fendt", "")
	var plain []domain.Machine
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &plain) != nil || len(plain) != 1 {
		t.Fatalf("fără page, tabloul simplu: %d %s", rec.Code, rec.Body.String())
	}
	if gotPage != nil || gotFilter.Query != "fendt" {
		t.Errorf("fără page: filtru %+v, pagină %+v", gotFilter, gotPage)
	}

	rec = do(t, router, http.MethodGet, "/machines?page=3&page_size=10&q=fendt", "")
	var page domain.Page[domain.Machine]
	if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &page) != nil {
		t.Fatalf("pagina: %d %s", rec.Code, rec.Body.String())
	}
	if page.Total != 41 || page.Page != 3 || page.PageSize != 10 || len(page.Items) != 1 {
		t.Errorf("plicul paginii: %+v", page)
	}
	if gotPage == nil || gotPage.Offset() != 20 {
		t.Errorf("pagina 3 × 10 sare 20 de rânduri: %+v", gotPage)
	}

	if rec = do(t, router, http.MethodGet, "/machines?page=2", ""); rec.Code != http.StatusOK || gotPage.PageSize != domain.DefaultPageSize {
		t.Errorf("mărimea implicită: %d %+v", rec.Code, gotPage)
	}
	for _, query := range []string{"page=0", "page=x", "page=1&page_size=0", "page=1&page_size=201"} {
		if rec = do(t, router, http.MethodGet, "/machines?"+query, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s trebuie refuzat: %d", query, rec.Code)
		}
	}
}
