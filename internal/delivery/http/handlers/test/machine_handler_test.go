package handlers_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"agri-api/internal/delivery/http/handlers"
	"agri-api/internal/domain"
	"agri-api/internal/usecase"

	"github.com/gin-gonic/gin"
)

func newMachineRouter(repo *machineRepoMock, audit *auditRepoMock, notif *notificationRepoMock) *gin.Engine {
	handler := handlers.NewMachineHandler(
		usecase.NewMachineService(repo),
		handlers.WithMachineAudit(usecase.NewAuditService(audit)),
		handlers.WithMachineNotif(usecase.NewNotificationService(notif)),
	)
	router := gin.New()
	router.Use(withUser("7"))
	router.GET("/machines", handler.GetAll)
	router.GET("/machines/:id", handler.GetByID)
	router.POST("/machines", handler.Create)
	router.PATCH("/machines/:id", handler.Update)
	router.DELETE("/machines/:id", handler.Delete)
	return router
}

func TestMachineHandler_CreateAndRead(t *testing.T) {
	repo := &machineRepoMock{
		create: func(m *domain.Machine) error { m.ID = 5; return nil },
		getAll: func() ([]domain.Machine, error) { return []domain.Machine{{ID: 1}}, nil },
		getByID: func(id int64) (*domain.Machine, error) {
			if id == 1 {
				return &domain.Machine{ID: 1, Name: "T1"}, nil
			}
			return nil, nil
		},
	}
	router := newMachineRouter(repo, &auditRepoMock{}, &notificationRepoMock{})

	if rec := do(t, router, http.MethodPost, "/machines", `{"name":"T1"`); rec.Code != http.StatusBadRequest {
		t.Errorf("JSON invalid: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPost, "/machines", `{"name":"T1","code":"TR-1","type":"barca"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("tip invalid: %d", rec.Code)
	}
	rec := do(t, router, http.MethodPost, "/machines", `{"name":"T1","code":"TR-1","type":"tractor","year":2020}`)
	if rec.Code != http.StatusCreated || !strings.Contains(rec.Body.String(), `"id":5`) {
		t.Errorf("creare: %d %s", rec.Code, rec.Body.String())
	}

	if rec := do(t, router, http.MethodGet, "/machines", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"id":1`) {
		t.Errorf("listare: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(t, router, http.MethodGet, "/machines/abc", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("id invalid: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, "/machines/9", ""); rec.Code != http.StatusNotFound {
		t.Errorf("inexistent: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, "/machines/1", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"name":"T1"`) {
		t.Errorf("citire: %d %s", rec.Code, rec.Body.String())
	}

	repo.getAll = func() ([]domain.Machine, error) { return nil, errors.New("db down") }
	repo.getByID = func(int64) (*domain.Machine, error) { return nil, errors.New("db down") }
	if rec := do(t, router, http.MethodGet, "/machines", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroare listare: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodGet, "/machines/1", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroare citire: %d", rec.Code)
	}
}

func TestMachineHandler_UpdateAndDelete(t *testing.T) {
	existing := domain.Machine{ID: 1, Name: "T1", Code: "TR-1", Type: domain.MachineTypeTractor, Status: domain.MachineStatusActive}
	// ca în DB reală, fiecare citire întoarce o copie (handler-ul și serviciul nu partajează pointer-ul)
	repo := &machineRepoMock{getByID: func(id int64) (*domain.Machine, error) {
		if id == 1 {
			copy := existing
			return &copy, nil
		}
		return nil, nil
	}}
	audit := &auditRepoMock{}
	notif := &notificationRepoMock{}
	router := newMachineRouter(repo, audit, notif)
	valid := `{"name":"T1","code":"TR-1","type":"tractor","status":"maintenance"}`

	if rec := do(t, router, http.MethodPatch, "/machines/abc", valid); rec.Code != http.StatusBadRequest {
		t.Errorf("id invalid: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPatch, "/machines/9", valid); rec.Code != http.StatusNotFound {
		t.Errorf("inexistent: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPatch, "/machines/1", `{"name":"T1"}`); rec.Code != http.StatusBadRequest {
		t.Errorf("payload incomplet: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodPatch, "/machines/1", `{"name":"T1","code":"TR-1","type":"tractor","status":"pierdut"}`); rec.Code != http.StatusInternalServerError {
		t.Errorf("status invalid: %d", rec.Code)
	}

	// trecerea în mentenanță: audit + notificare
	rec := do(t, router, http.MethodPatch, "/machines/1", valid)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"status":"maintenance"`) {
		t.Errorf("actualizare: %d %s", rec.Code, rec.Body.String())
	}
	waitUntil(t, func() bool { return audit.count() == 1 && notif.createdCount() == 1 })
	if *audit.entries[0].ActorID != 7 || audit.entries[0].Changes["status"] != "maintenance" || audit.entries[0].Changes["old_status"] != "active" {
		t.Errorf("intrare de audit greșită: %+v", audit.entries[0])
	}

	if rec := do(t, router, http.MethodDelete, "/machines/abc", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("ștergere id invalid: %d", rec.Code)
	}
	if rec := do(t, router, http.MethodDelete, "/machines/1", ""); rec.Code != http.StatusOK {
		t.Errorf("ștergere: %d", rec.Code)
	}
	repo.delete = func(int64) error { return errors.New("db down") }
	if rec := do(t, router, http.MethodDelete, "/machines/1", ""); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroare ștergere: %d", rec.Code)
	}
	repo.getByID = func(int64) (*domain.Machine, error) { return nil, errors.New("db down") }
	if rec := do(t, router, http.MethodPatch, "/machines/1", valid); rec.Code != http.StatusInternalServerError {
		t.Errorf("eroare citire la actualizare: %d", rec.Code)
	}
}
