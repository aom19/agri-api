package handlers_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

// fieldOpRepoFake ține lucrările în memorie; err, când e setat, e întors de toate citirile.
type fieldOpRepoFake struct {
	repository.FieldOperationRepository
	items      map[int64]*dto.FieldOperationResponse
	operatorOf map[int64]int64
	lastFilter repository.FieldOperationFilter
	err        error
}

func newFieldOpRepoFake() *fieldOpRepoFake {
	active := string(domain.AssetStatusActive)
	machine := int64(4)
	return &fieldOpRepoFake{
		items: map[int64]*dto.FieldOperationResponse{
			1: {ID: 1, FieldID: "f1", FieldName: "Lot 1", Status: "planned", MachineID: &machine, MachineStatus: &active},
			2: {ID: 2, FieldID: "f1", FieldName: "Lot 1", Status: "in_progress"},
			3: {ID: 3, FieldID: "f1", FieldName: "Lot 1", Status: "completed"},
		},
		operatorOf: map[int64]int64{1: 7, 2: 7},
	}
}

func (r *fieldOpRepoFake) List(f repository.FieldOperationFilter, p *domain.PageRequest) ([]dto.FieldOperationResponse, int, error) {
	r.lastFilter = f
	if r.err != nil {
		return nil, 0, r.err
	}
	return []dto.FieldOperationResponse{*r.items[1]}, 1, nil
}
func (r *fieldOpRepoFake) CountByStatus(f repository.FieldOperationFilter) (map[string]int, error) {
	r.lastFilter = f
	if r.err != nil {
		return nil, r.err
	}
	return map[string]int{"planned": 2, "completed": 5}, nil
}
func (r *fieldOpRepoFake) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.items[id], nil
}
func (r *fieldOpRepoFake) GetByIDForAssignedUser(id, userID int64) (*dto.FieldOperationResponse, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.operatorOf[id] != userID {
		return nil, nil
	}
	return r.items[id], nil
}
func (r *fieldOpRepoFake) Create(op *domain.FieldOperation) error {
	op.ID = 9
	r.items[9] = &dto.FieldOperationResponse{ID: 9, FieldID: op.FieldID, Status: string(op.Status)}
	return nil
}
func (r *fieldOpRepoFake) Update(id int64, op *domain.FieldOperation) error {
	r.items[id].Status = string(op.Status)
	return nil
}
func (r *fieldOpRepoFake) MarkStarted(id int64, _ time.Time) error {
	r.items[id].Status = "in_progress"
	return nil
}
func (r *fieldOpRepoFake) Delete(id int64) error {
	delete(r.items, id)
	return nil
}
func (r *fieldOpRepoFake) ArchivedReferences(*domain.FieldOperation) ([]string, error) {
	return nil, nil
}
func (r *fieldOpRepoFake) FindConflicts(repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	return []dto.FieldOperationConflict{{OperationID: 2, Resources: []string{"machine"}}}, nil
}
func (r *fieldOpRepoFake) InProgressConflicts(repository.FieldOperationConflictQuery) ([]dto.FieldOperationConflict, error) {
	return nil, nil
}

// asUser pune în context contul și rolul, ca AuthMiddleware.
func asUser(userID, role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Set("role_code", role)
	}
}

func fieldOpRouter(repo *fieldOpRepoFake, mw gin.HandlerFunc) (*gin.Engine, *notificationRepoMock) {
	notifs := &notificationRepoMock{}
	completion := usecase.NewFieldOperationCompletionService(nil, repo, stockMovementRepoStub{})
	h := handlers.NewFieldOperationHandler(
		usecase.NewFieldOperationService(repo),
		handlers.WithFieldOpAudit(usecase.NewAuditService(&auditRepoMock{})),
		handlers.WithFieldOpNotif(usecase.NewNotificationService(notifs)),
		handlers.WithFieldOpCompletion(completion),
	)
	r := gin.New()
	r.Use(mw)
	r.GET("/ops", h.GetAll)
	r.GET("/ops/conflicts", h.Conflicts)
	r.GET("/ops/status-counts", h.StatusCounts)
	r.GET("/ops/:id", h.GetByID)
	r.POST("/ops", h.Create)
	r.PATCH("/ops/:id", h.Update)
	r.PATCH("/ops/:id/start", h.Start)
	r.PATCH("/ops/:id/complete", h.Complete)
	r.DELETE("/ops/:id", h.Delete)
	return r, notifs
}

func TestFieldOperationHandler_ListFilters(t *testing.T) {
	repo := newFieldOpRepoFake()
	r, _ := fieldOpRouter(repo, asUser("1", "admin"))

	rec := do(t, r, http.MethodGet, "/ops?page=1&q=lot&status=planned,%20in_progress&operation_type=seeding&field_id=6f1c2b8e-4a1d-4d6e-9c2a-1b2c3d4e5f60&machine_id=4&operator_id=7&planned_from=2026-01-01T00:00:00Z&planned_to=2026-02-01T00:00:00Z&sort=planned_asc", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"total":1`) {
		t.Fatalf("listă: %d %s", rec.Code, rec.Body)
	}
	f := repo.lastFilter
	if f.Search != "lot" || len(f.Statuses) != 2 || f.Statuses[1] != "in_progress" || f.OperatorID != "7" ||
		f.PlannedFrom == nil || f.PlannedTo == nil || f.Sort != repository.FieldOperationSortPlannedAsc || f.AssignedUserID != 0 {
		t.Errorf("filtrele trebuie trecute la repository: %+v", f)
	}

	for _, query := range []string{
		"operation_type=x", "field_id=abc", "machine_id=x", "operator_id=x",
		"planned_from=2026-01-01", "planned_to=ieri", "sort=random", "page=0",
	} {
		if rec := do(t, r, http.MethodGet, "/ops?"+query, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: vreau 400, am %d", query, rec.Code)
		}
	}

	repo.err = errors.New("db down")
	if rec := do(t, r, http.MethodGet, "/ops", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroarea repository-ului: %d", rec.Code)
	}
}

func TestFieldOperationHandler_OperatorSeesOnlyOwn(t *testing.T) {
	repo := newFieldOpRepoFake()
	r, _ := fieldOpRouter(repo, asUser("7", "operator"))

	if rec := do(t, r, http.MethodGet, "/ops?operator_id=5", ""); rec.Code != http.StatusOK {
		t.Fatalf("listă operator: %d", rec.Code)
	}
	if repo.lastFilter.OperatorID != "" || repo.lastFilter.AssignedUserID != 7 {
		t.Errorf("operatorul vede doar lucrările lui, indiferent de operator_id: %+v", repo.lastFilter)
	}
	if rec := do(t, r, http.MethodGet, "/ops/1", ""); rec.Code != http.StatusOK {
		t.Errorf("lucrarea proprie: %d", rec.Code)
	}
	if rec := do(t, r, http.MethodGet, "/ops/3", ""); rec.Code != http.StatusNotFound {
		t.Errorf("lucrarea altcuiva: %d", rec.Code)
	}

	noUser, _ := fieldOpRouter(repo, asUser("x", "operator"))
	for _, req := range [][2]string{{http.MethodGet, "/ops"}, {http.MethodGet, "/ops/1"}, {http.MethodPatch, "/ops/1/start"}, {http.MethodPatch, "/ops/2/complete"}} {
		if rec := do(t, noUser, req[0], req[1], "{}"); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s fără cont valid: %d", req[0], req[1], rec.Code)
		}
	}
}

func TestFieldOperationHandler_CRUD(t *testing.T) {
	repo := newFieldOpRepoFake()
	r, _ := fieldOpRouter(repo, asUser("1", "admin"))

	if rec := do(t, r, http.MethodGet, "/ops/1", ""); rec.Code != http.StatusOK {
		t.Errorf("detalii: %d", rec.Code)
	}
	for path, want := range map[string]int{"/ops/x": 400, "/ops/99": 404} {
		if rec := do(t, r, http.MethodGet, path, ""); rec.Code != want {
			t.Errorf("GET %s: vreau %d, am %d", path, want, rec.Code)
		}
	}

	if rec := do(t, r, http.MethodPost, "/ops", `{"field_id":"f1","operation_type":"seeding"}`); rec.Code != http.StatusCreated {
		t.Errorf("creare: %d %s", rec.Code, rec.Body)
	}
	if rec := do(t, r, http.MethodPost, "/ops", `{"field_id":"f1"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("creare fără tip: %d", rec.Code)
	}
	if rec := do(t, r, http.MethodPost, "/ops", `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("creare cu JSON invalid: %d", rec.Code)
	}

	if rec := do(t, r, http.MethodPatch, "/ops/1", `{"field_id":"f1","operation_type":"seeding","status":"planned"}`); rec.Code != http.StatusOK {
		t.Errorf("editare: %d %s", rec.Code, rec.Body)
	}
	for path, body := range map[string]string{"/ops/x": `{}`, "/ops/1": `{`} {
		if rec := do(t, r, http.MethodPatch, path, body); rec.Code != http.StatusBadRequest {
			t.Errorf("editare %s %s: %d", path, body, rec.Code)
		}
	}
	if rec := do(t, r, http.MethodPatch, "/ops/99", `{"field_id":"f1","operation_type":"seeding"}`); rec.Code != http.StatusNotFound {
		t.Errorf("editare inexistentă: %d", rec.Code)
	}

	if rec := do(t, r, http.MethodDelete, "/ops/3", ""); rec.Code != http.StatusNoContent {
		t.Errorf("ștergere: %d", rec.Code)
	}
	for path, want := range map[string]int{"/ops/x": 400, "/ops/3": 404} {
		if rec := do(t, r, http.MethodDelete, path, ""); rec.Code != want {
			t.Errorf("DELETE %s: vreau %d, am %d", path, want, rec.Code)
		}
	}
	repo.err = errors.New("db down")
	if rec := do(t, r, http.MethodDelete, "/ops/1", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("ștergere cu eroare: %d", rec.Code)
	}
	if rec := do(t, r, http.MethodGet, "/ops/1", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("detalii cu eroare: %d", rec.Code)
	}
}

func TestFieldOperationHandler_StartAndComplete(t *testing.T) {
	repo := newFieldOpRepoFake()
	r, notifs := fieldOpRouter(repo, asUser("1", "admin"))

	if rec := do(t, r, http.MethodPatch, "/ops/1/start", ""); rec.Code != http.StatusOK || repo.items[1].Status != "in_progress" {
		t.Fatalf("pornire: %d %s", rec.Code, rec.Body)
	}
	waitUntil(t, func() bool { return notifs.createdCount() == 1 })
	for path, want := range map[string]int{"/ops/x/start": 400, "/ops/99/start": 404, "/ops/3/start": 400} {
		if rec := do(t, r, http.MethodPatch, path, ""); rec.Code != want {
			t.Errorf("PATCH %s: vreau %d, am %d", path, want, rec.Code)
		}
	}

	operator, _ := fieldOpRouter(repo, asUser("7", "operator"))
	if rec := do(t, operator, http.MethodPatch, "/ops/2/start", ""); rec.Code != http.StatusOK {
		t.Errorf("operatorul își pornește lucrarea: %d", rec.Code)
	}
	if rec := do(t, operator, http.MethodPatch, "/ops/3/start", ""); rec.Code != http.StatusNotFound {
		t.Errorf("operatorul nu pornește lucrarea altcuiva: %d", rec.Code)
	}

	for path, want := range map[string]int{"/ops/x/complete": 400, "/ops/99/complete": 404, "/ops/3/complete": 400} {
		if rec := do(t, r, http.MethodPatch, path, `{}`); rec.Code != want {
			t.Errorf("PATCH %s: vreau %d, am %d", path, want, rec.Code)
		}
	}
	if rec := do(t, r, http.MethodPatch, "/ops/2/complete", `{`); rec.Code != http.StatusBadRequest {
		t.Errorf("finalizare cu JSON invalid: %d", rec.Code)
	}
	if rec := do(t, operator, http.MethodPatch, "/ops/3/complete", `{}`); rec.Code != http.StatusNotFound {
		t.Errorf("operatorul nu finalizează lucrarea altcuiva: %d", rec.Code)
	}
	repo.err = errors.New("db down")
	if rec := do(t, r, http.MethodPatch, "/ops/2/complete", `{}`); rec.Code != http.StatusInternalServerError {
		t.Errorf("finalizare cu eroare: %d", rec.Code)
	}

	unconfigured := gin.New()
	h := handlers.NewFieldOperationHandler(usecase.NewFieldOperationService(repo))
	unconfigured.PATCH("/ops/:id/complete", h.Complete)
	if rec := do(t, unconfigured, http.MethodPatch, "/ops/2/complete", `{}`); rec.Code != http.StatusNotImplemented {
		t.Errorf("fără serviciul de finalizare: %d", rec.Code)
	}
}

func TestFieldOperationHandler_Conflicts(t *testing.T) {
	r, _ := fieldOpRouter(newFieldOpRepoFake(), asUser("1", "admin"))

	rec := do(t, r, http.MethodGet, "/ops/conflicts?exclude_id=1&machine_id=4&planned_start_at=2026-05-01T08:00:00Z&planned_end_at=2026-05-01T12:00:00Z", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"resources":["machine"]`) {
		t.Fatalf("conflicte: %d %s", rec.Code, rec.Body)
	}
	for _, query := range []string{
		"exclude_id=x", "machine_id=0", "operator_id=abc", "planned_start_at=azi",
		"planned_start_at=2026-05-01T08:00:00Z&planned_end_at=mâine",
		"machine_id=4&planned_start_at=2026-05-01T08:00:00Z&planned_end_at=2026-04-01T08:00:00Z",
	} {
		if rec := do(t, r, http.MethodGet, "/ops/conflicts?"+query, ""); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: vreau 400, am %d", query, rec.Code)
		}
	}
}

func TestFieldOperationHandler_StatusCounts(t *testing.T) {
	repo := newFieldOpRepoFake()
	r, _ := fieldOpRouter(repo, asUser("7", "operator"))

	rec := do(t, r, http.MethodGet, "/ops/status-counts?q=lot&status=completed", "")
	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `"total":7`) || !strings.Contains(body, `"in_progress":0`) {
		t.Fatalf("numărători: %d %s", rec.Code, body)
	}
	if repo.lastFilter.Search != "lot" || repo.lastFilter.AssignedUserID != 7 {
		t.Errorf("aceleași filtre ca lista, restrânse la operator: %+v", repo.lastFilter)
	}
	if rec := do(t, r, http.MethodGet, "/ops/status-counts?sort=x", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("filtru invalid: %d", rec.Code)
	}
	repo.err = errors.New("db down")
	if rec := do(t, r, http.MethodGet, "/ops/status-counts", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroarea repository-ului: %d", rec.Code)
	}
	for _, sort := range []string{"recent", "planned_desc", "area_asc", "area_desc", "field_asc", "field_desc"} {
		repo.err = nil
		if rec := do(t, r, http.MethodGet, "/ops?sort="+sort, ""); rec.Code != http.StatusOK {
			t.Errorf("sort=%s: %d", sort, rec.Code)
		}
	}
}
