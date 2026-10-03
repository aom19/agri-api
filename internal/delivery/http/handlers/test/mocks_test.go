package handlers_test

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"agri-api/internal/domain"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// withUser pune în context utilizatorul autentificat, așa cum face AuthMiddleware.
func withUser(userID interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID != nil {
			c.Set("user_id", userID)
		}
	}
}

func do(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader *strings.Reader
	if body != "" {
		reader = strings.NewReader(body)
	} else {
		reader = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func waitUntil(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("condiția nu s-a îndeplinit la timp")
}

type machineRepoMock struct {
	getAll  func() ([]domain.Machine, error)
	getByID func(int64) (*domain.Machine, error)
	create  func(*domain.Machine) error
	update  func(int64, *domain.Machine) error
	delete  func(int64) error
}

func (m *machineRepoMock) Create(x *domain.Machine) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *machineRepoMock) GetAll() ([]domain.Machine, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *machineRepoMock) GetByID(id int64) (*domain.Machine, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *machineRepoMock) Update(id int64, x *domain.Machine) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *machineRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *machineRepoMock) UpdateStatus(*sql.Tx, int64, domain.MachineStatus) error { return nil }
func (m *machineRepoMock) DB() *sql.DB                                             { return nil }

type dashboardRepoMock struct {
	getCardStats  func() (*domain.DashboardCardStats, error)
	getQuickStats func() (*domain.DashboardQuickStats, error)
}

func (m *dashboardRepoMock) GetCardStats() (*domain.DashboardCardStats, error) {
	if m.getCardStats != nil {
		return m.getCardStats()
	}
	return &domain.DashboardCardStats{}, nil
}
func (m *dashboardRepoMock) GetQuickStats() (*domain.DashboardQuickStats, error) {
	if m.getQuickStats != nil {
		return m.getQuickStats()
	}
	return &domain.DashboardQuickStats{}, nil
}

type auditRepoMock struct {
	mu      sync.Mutex
	entries []domain.AuditEntry
	getAll  func(int, string, string) ([]domain.AuditEntry, error)
}

func (m *auditRepoMock) Log(e *domain.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, *e)
	return nil
}
func (m *auditRepoMock) GetAll(limit int, entityType, entityID string) ([]domain.AuditEntry, error) {
	if m.getAll != nil {
		return m.getAll(limit, entityType, entityID)
	}
	return nil, nil
}
func (m *auditRepoMock) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.entries)
}

type notificationRepoMock struct {
	mu            sync.Mutex
	created       int
	getByUser     func(int64, bool, int) ([]domain.UserNotification, error)
	countUnread   func(int64) (int64, error)
	markAsRead    func(int64, int64) error
	markAllAsRead func(int64) error
}

func (m *notificationRepoMock) Create(n *domain.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.created++
	n.ID = int64(m.created)
	return nil
}
func (m *notificationRepoMock) CreateUserNotification(int64, int64) error { return nil }
func (m *notificationRepoMock) GetByUser(userID int64, onlyUnread bool, limit int) ([]domain.UserNotification, error) {
	if m.getByUser != nil {
		return m.getByUser(userID, onlyUnread, limit)
	}
	return nil, nil
}
func (m *notificationRepoMock) CountUnread(userID int64) (int64, error) {
	if m.countUnread != nil {
		return m.countUnread(userID)
	}
	return 0, nil
}
func (m *notificationRepoMock) MarkAsRead(id, userID int64) error {
	if m.markAsRead != nil {
		return m.markAsRead(id, userID)
	}
	return nil
}
func (m *notificationRepoMock) MarkAllAsRead(userID int64) error {
	if m.markAllAsRead != nil {
		return m.markAllAsRead(userID)
	}
	return nil
}
func (m *notificationRepoMock) GetManagerAndAdminUserIDs() ([]int64, error) { return []int64{1}, nil }
func (m *notificationRepoMock) createdCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.created
}
