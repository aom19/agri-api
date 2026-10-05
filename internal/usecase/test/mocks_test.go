package usecase_test

import (
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
)

// Mock-uri simple pentru repository-uri, bazate pe câmpuri-funcție: metodele fără
// implementare setată returnează valori zero, deci fiecare test configurează doar ce-i trebuie.

func ptr[T any](v T) *T { return &v }

// newSQLMock creează un *sql.DB fals pentru serviciile care deschid tranzacții.
func newSQLMock(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, mock
}

// testLogger implementează OverdueMonitorLogger și reține mesajele.
type testLogger struct {
	mu     sync.Mutex
	infos  []string
	errors []string
}

func (l *testLogger) Infof(template string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infos = append(l.infos, fmt.Sprintf(template, args...))
}

func (l *testLogger) Errorf(template string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errors = append(l.errors, fmt.Sprintf(template, args...))
}

func (l *testLogger) errorCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.errors)
}

// ─── Machine ─────────────────────────────────────────────────────────────────

type machineRepoMock struct {
	create       func(*domain.Machine) error
	getAll       func() ([]domain.Machine, error)
	getByID      func(int64) (*domain.Machine, error)
	update       func(int64, *domain.Machine) error
	delete       func(int64) error
	updateStatus func(*sql.Tx, int64, domain.MachineStatus) error
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
func (m *machineRepoMock) UpdateStatus(tx *sql.Tx, id int64, s domain.MachineStatus) error {
	if m.updateStatus != nil {
		return m.updateStatus(tx, id, s)
	}
	return nil
}
func (m *machineRepoMock) DB() *sql.DB { return nil }

// ─── Operator ────────────────────────────────────────────────────────────────

type operatorRepoMock struct {
	create    func(*domain.Operator) error
	getAll    func() ([]domain.Operator, error)
	getByID   func(int64) (*domain.Operator, error)
	update    func(int64, *domain.Operator) error
	setActive func(int64, bool) error
}

func (m *operatorRepoMock) Create(x *domain.Operator) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *operatorRepoMock) GetAll() ([]domain.Operator, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *operatorRepoMock) GetByID(id int64) (*domain.Operator, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *operatorRepoMock) Update(id int64, x *domain.Operator) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *operatorRepoMock) SetActive(id int64, active bool) error {
	if m.setActive != nil {
		return m.setActive(id, active)
	}
	return nil
}

// ─── Implement ───────────────────────────────────────────────────────────────

type implementRepoMock struct {
	create     func(*domain.Implement) error
	getAll     func() ([]domain.Implement, error)
	getByID    func(int64) (*domain.Implement, error)
	update     func(int64, *domain.Implement) error
	delete     func(int64) error
	activate   func(int64) error
	deactivate func(int64) error
}

func (m *implementRepoMock) Create(x *domain.Implement) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *implementRepoMock) GetAll() ([]domain.Implement, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *implementRepoMock) GetByID(id int64) (*domain.Implement, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *implementRepoMock) Update(id int64, x *domain.Implement) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *implementRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *implementRepoMock) Activate(id int64) error {
	if m.activate != nil {
		return m.activate(id)
	}
	return nil
}
func (m *implementRepoMock) Deactivate(id int64) error {
	if m.deactivate != nil {
		return m.deactivate(id)
	}
	return nil
}
func (m *implementRepoMock) DB() *sql.DB { return nil }

// ─── Resource types / resources / stocks ─────────────────────────────────────

type resourceTypeRepoMock struct {
	getAll  func() ([]domain.ResourceType, error)
	getByID func(int64) (*domain.ResourceType, error)
	create  func(*domain.ResourceType) error
	update  func(int64, *domain.ResourceType) error
	delete  func(int64) error
}

func (m *resourceTypeRepoMock) GetAll() ([]domain.ResourceType, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *resourceTypeRepoMock) GetByID(id int64) (*domain.ResourceType, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *resourceTypeRepoMock) Create(x *domain.ResourceType) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *resourceTypeRepoMock) Update(id int64, x *domain.ResourceType) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *resourceTypeRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}

type resourceRepoMock struct {
	getAll  func() ([]domain.Resource, error)
	getByID func(int64) (*domain.Resource, error)
	create  func(*domain.Resource) error
	update  func(int64, *domain.Resource) error
	delete  func(int64) error
}

func (m *resourceRepoMock) GetAll() ([]domain.Resource, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *resourceRepoMock) GetByID(id int64) (*domain.Resource, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *resourceRepoMock) Create(x *domain.Resource) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *resourceRepoMock) Update(id int64, x *domain.Resource) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *resourceRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}

type stockRepoMock struct {
	getAll          func() ([]domain.Stock, error)
	getByID         func(int64) (*domain.Stock, error)
	getByResourceID func(int64) (*domain.Stock, error)
	create          func(*domain.Stock) error
	updateMinimum   func(int64, float64) error
	delete          func(int64) error
}

func (m *stockRepoMock) GetAll() ([]domain.Stock, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *stockRepoMock) GetByID(id int64) (*domain.Stock, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *stockRepoMock) GetByResourceID(id int64) (*domain.Stock, error) {
	if m.getByResourceID != nil {
		return m.getByResourceID(id)
	}
	return nil, nil
}
func (m *stockRepoMock) Create(_ *sql.Tx, x *domain.Stock) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *stockRepoMock) UpdateMinimum(id int64, minimum float64) error {
	if m.updateMinimum != nil {
		return m.updateMinimum(id, minimum)
	}
	return nil
}
func (m *stockRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}

// ─── Field ───────────────────────────────────────────────────────────────────

type fieldRepoMock struct {
	create  func(*domain.Field) error
	getAll  func() ([]domain.Field, error)
	getByID func(string) (*domain.Field, error)
	update  func(string, *domain.Field) error
	delete  func(string) error
}

func (m *fieldRepoMock) Create(x *domain.Field) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *fieldRepoMock) GetAll() ([]domain.Field, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *fieldRepoMock) GetByID(id string) (*domain.Field, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *fieldRepoMock) Update(id string, x *domain.Field) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *fieldRepoMock) Delete(id string) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}

// ─── Operation types / templates ─────────────────────────────────────────────

type operationTypeRepoMock struct {
	getAll  func() ([]domain.OperationType, error)
	getByID func(int64) (*domain.OperationType, error)
	create  func(*domain.OperationType) error
	update  func(int64, *domain.OperationType) error
	delete  func(int64) error
}

func (m *operationTypeRepoMock) GetAll() ([]domain.OperationType, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *operationTypeRepoMock) GetByID(id int64) (*domain.OperationType, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *operationTypeRepoMock) Create(x *domain.OperationType) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *operationTypeRepoMock) Update(id int64, x *domain.OperationType) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *operationTypeRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}

type operationTemplateRepoMock struct {
	getAll             func() ([]domain.OperationTemplate, error)
	getByID            func(int64) (*domain.OperationTemplate, error)
	getByOperationType func(int64) ([]domain.OperationTemplate, error)
	create             func(*domain.OperationTemplate) error
	update             func(int64, *domain.OperationTemplate) error
	delete             func(int64) error
	setResources       func(int64, []domain.TemplateResource) error
	setMachineTypes    func(int64, []string) error
	setImplementTypes  func(int64, []string) error
}

func (m *operationTemplateRepoMock) GetAll() ([]domain.OperationTemplate, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *operationTemplateRepoMock) GetByID(id int64) (*domain.OperationTemplate, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *operationTemplateRepoMock) GetByOperationType(id int64) ([]domain.OperationTemplate, error) {
	if m.getByOperationType != nil {
		return m.getByOperationType(id)
	}
	return nil, nil
}
func (m *operationTemplateRepoMock) Create(x *domain.OperationTemplate) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *operationTemplateRepoMock) Update(id int64, x *domain.OperationTemplate) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *operationTemplateRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *operationTemplateRepoMock) SetResources(id int64, r []domain.TemplateResource) error {
	if m.setResources != nil {
		return m.setResources(id, r)
	}
	return nil
}
func (m *operationTemplateRepoMock) SetMachineTypes(id int64, t []string) error {
	if m.setMachineTypes != nil {
		return m.setMachineTypes(id, t)
	}
	return nil
}
func (m *operationTemplateRepoMock) SetImplementTypes(id int64, t []string) error {
	if m.setImplementTypes != nil {
		return m.setImplementTypes(id, t)
	}
	return nil
}

// ─── Field operations ────────────────────────────────────────────────────────

type fieldOperationRepoMock struct {
	getAll                 func(repository.FieldOperationFilter) ([]dto.FieldOperationResponse, error)
	getByID                func(int64) (*dto.FieldOperationResponse, error)
	getByIDForAssignedUser func(int64, int64) (*dto.FieldOperationResponse, error)
	create                 func(*domain.FieldOperation) error
	update                 func(int64, *domain.FieldOperation) error
	updateChecklist        func(int64, domain.FieldOperationChecklist) error
	updateStatus           func(int64, domain.FieldOperationStatus) error
	markStarted            func(int64, time.Time) error
	complete               func(*sql.Tx, int64, domain.FieldOperationCompletion, time.Time) error
	getTemplateResources   func(int64) ([]domain.TemplateResourceUsage, error)
	addMachineHours        func(*sql.Tx, int64, float64) error
	delete                 func(int64) error
	getOverdueInProgress   func(time.Time) ([]dto.OverdueFieldOperation, error)
	markOverdueNotified    func(int64, time.Time) error
}

func (m *fieldOperationRepoMock) GetAll(f repository.FieldOperationFilter) ([]dto.FieldOperationResponse, error) {
	if m.getAll != nil {
		return m.getAll(f)
	}
	return nil, nil
}
func (m *fieldOperationRepoMock) GetByID(id int64) (*dto.FieldOperationResponse, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *fieldOperationRepoMock) GetByIDForAssignedUser(id, userID int64) (*dto.FieldOperationResponse, error) {
	if m.getByIDForAssignedUser != nil {
		return m.getByIDForAssignedUser(id, userID)
	}
	return nil, nil
}
func (m *fieldOperationRepoMock) Create(x *domain.FieldOperation) error {
	if m.create != nil {
		return m.create(x)
	}
	return nil
}
func (m *fieldOperationRepoMock) Update(id int64, x *domain.FieldOperation) error {
	if m.update != nil {
		return m.update(id, x)
	}
	return nil
}
func (m *fieldOperationRepoMock) UpdateChecklist(id int64, c domain.FieldOperationChecklist) error {
	if m.updateChecklist != nil {
		return m.updateChecklist(id, c)
	}
	return nil
}
func (m *fieldOperationRepoMock) UpdateStatus(id int64, s domain.FieldOperationStatus) error {
	if m.updateStatus != nil {
		return m.updateStatus(id, s)
	}
	return nil
}
func (m *fieldOperationRepoMock) MarkStarted(id int64, at time.Time) error {
	if m.markStarted != nil {
		return m.markStarted(id, at)
	}
	return nil
}
func (m *fieldOperationRepoMock) Complete(tx *sql.Tx, id int64, c domain.FieldOperationCompletion, at time.Time) error {
	if m.complete != nil {
		return m.complete(tx, id, c, at)
	}
	return nil
}
func (m *fieldOperationRepoMock) GetTemplateResources(id int64) ([]domain.TemplateResourceUsage, error) {
	if m.getTemplateResources != nil {
		return m.getTemplateResources(id)
	}
	return nil, nil
}
func (m *fieldOperationRepoMock) AddMachineHours(tx *sql.Tx, id int64, h float64) error {
	if m.addMachineHours != nil {
		return m.addMachineHours(tx, id, h)
	}
	return nil
}
func (m *fieldOperationRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *fieldOperationRepoMock) GetOverdueInProgress(now time.Time) ([]dto.OverdueFieldOperation, error) {
	if m.getOverdueInProgress != nil {
		return m.getOverdueInProgress(now)
	}
	return nil, nil
}
func (m *fieldOperationRepoMock) MarkOverdueNotified(id int64, at time.Time) error {
	if m.markOverdueNotified != nil {
		return m.markOverdueNotified(id, at)
	}
	return nil
}

// ─── Dashboard / audit / notifications ───────────────────────────────────────

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
	logErr  error
	getAll  func(int, string, string) ([]domain.AuditEntry, error)
}

func (m *auditRepoMock) Log(e *domain.AuditEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, *e)
	return m.logErr
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
	created       []domain.Notification
	userNotifs    map[int64][]int64 // userID -> notification IDs
	createErr     error
	userNotifErr  func(userID int64) error
	adminIDs      []int64
	adminErr      error
	getByUser     func(int64, bool, int) ([]domain.UserNotification, error)
	countUnread   func(int64) (int64, error)
	markAsRead    func(int64, int64) error
	markAllAsRead func(int64) error
}

func (m *notificationRepoMock) Create(n *domain.Notification) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.createErr != nil {
		return m.createErr
	}
	n.ID = int64(len(m.created) + 1)
	n.CreatedAt = time.Now()
	m.created = append(m.created, *n)
	return nil
}
func (m *notificationRepoMock) CreateUserNotification(notificationID, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.userNotifErr != nil {
		if err := m.userNotifErr(userID); err != nil {
			return err
		}
	}
	if m.userNotifs == nil {
		m.userNotifs = map[int64][]int64{}
	}
	m.userNotifs[userID] = append(m.userNotifs[userID], notificationID)
	return nil
}
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
func (m *notificationRepoMock) GetManagerAndAdminUserIDs() ([]int64, error) {
	return m.adminIDs, m.adminErr
}
func (m *notificationRepoMock) deliveries(userID int64) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.userNotifs[userID])
}
func (m *notificationRepoMock) createdCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.created)
}

// ─── Users / roles / permissions ─────────────────────────────────────────────

type userRepoMock struct {
	getAll             func() ([]domain.User, error)
	getByEmail         func(string) (*domain.User, error)
	getByID            func(int64) (*domain.User, error)
	create             func(*domain.User) error
	update             func(*domain.User) error
	delete             func(int64) error
	disable            func(int64) error
	enable             func(int64) error
	markEmailConfirmed func(int64) error
	updatePassword     func(int64, string) error
	updateRole         func(int64, int64) error
	getProfile         func(int64) (*domain.UserProfile, error)
	upsertProfile      func(*domain.UserProfile) error
}

func (m *userRepoMock) GetAll() ([]domain.User, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *userRepoMock) GetByEmail(email string) (*domain.User, error) {
	if m.getByEmail != nil {
		return m.getByEmail(email)
	}
	return nil, nil
}
func (m *userRepoMock) GetByID(id int64) (*domain.User, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *userRepoMock) Create(u *domain.User) error {
	if m.create != nil {
		return m.create(u)
	}
	return nil
}
func (m *userRepoMock) Update(u *domain.User) error {
	if m.update != nil {
		return m.update(u)
	}
	return nil
}
func (m *userRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *userRepoMock) Disable(id int64) error {
	if m.disable != nil {
		return m.disable(id)
	}
	return nil
}
func (m *userRepoMock) Enable(id int64) error {
	if m.enable != nil {
		return m.enable(id)
	}
	return nil
}
func (m *userRepoMock) MarkEmailConfirmed(id int64) error {
	if m.markEmailConfirmed != nil {
		return m.markEmailConfirmed(id)
	}
	return nil
}
func (m *userRepoMock) UpdatePassword(id int64, hash string) error {
	if m.updatePassword != nil {
		return m.updatePassword(id, hash)
	}
	return nil
}
func (m *userRepoMock) UpdateRole(userID, roleID int64) error {
	if m.updateRole != nil {
		return m.updateRole(userID, roleID)
	}
	return nil
}
func (m *userRepoMock) GetProfile(id int64) (*domain.UserProfile, error) {
	if m.getProfile != nil {
		return m.getProfile(id)
	}
	return nil, nil
}
func (m *userRepoMock) UpsertProfile(p *domain.UserProfile) error {
	if m.upsertProfile != nil {
		return m.upsertProfile(p)
	}
	return nil
}

type roleRepoMock struct {
	getAll         func() ([]domain.Role, error)
	getByID        func(int64) (*domain.Role, error)
	getByCode      func(string) (*domain.Role, error)
	create         func(*domain.Role) error
	update         func(*domain.Role) error
	delete         func(int64) error
	getPermissions func(int64) ([]domain.Permission, error)
	setPermissions func(int64, []int64) error
}

func (m *roleRepoMock) GetAll() ([]domain.Role, error) {
	if m.getAll != nil {
		return m.getAll()
	}
	return nil, nil
}
func (m *roleRepoMock) GetByID(id int64) (*domain.Role, error) {
	if m.getByID != nil {
		return m.getByID(id)
	}
	return nil, nil
}
func (m *roleRepoMock) GetByCode(code string) (*domain.Role, error) {
	if m.getByCode != nil {
		return m.getByCode(code)
	}
	return nil, nil
}
func (m *roleRepoMock) Create(r *domain.Role) error {
	if m.create != nil {
		return m.create(r)
	}
	return nil
}
func (m *roleRepoMock) Update(r *domain.Role) error {
	if m.update != nil {
		return m.update(r)
	}
	return nil
}
func (m *roleRepoMock) Delete(id int64) error {
	if m.delete != nil {
		return m.delete(id)
	}
	return nil
}
func (m *roleRepoMock) GetPermissions(id int64) ([]domain.Permission, error) {
	if m.getPermissions != nil {
		return m.getPermissions(id)
	}
	return nil, nil
}
func (m *roleRepoMock) SetPermissions(id int64, ids []int64) error {
	if m.setPermissions != nil {
		return m.setPermissions(id, ids)
	}
	return nil
}

type permissionRepoMock struct {
	getAllPermissions func() ([]domain.Permission, error)
	getPermissionByID func(int64) (*domain.Permission, error)
	hasPermission     func(int64, string) (bool, error)
}

func (m *permissionRepoMock) GetAllPermissions() ([]domain.Permission, error) {
	if m.getAllPermissions != nil {
		return m.getAllPermissions()
	}
	return nil, nil
}
func (m *permissionRepoMock) GetPermissionByID(id int64) (*domain.Permission, error) {
	if m.getPermissionByID != nil {
		return m.getPermissionByID(id)
	}
	return nil, nil
}
func (m *permissionRepoMock) HasPermission(roleID int64, name string) (bool, error) {
	if m.hasPermission != nil {
		return m.hasPermission(roleID, name)
	}
	return false, nil
}

// ─── Stock movements ─────────────────────────────────────────────────────────

type stockMovementRepoMock struct {
	lockStockByID       func(*sql.Tx, int64) (*repository.StockLock, error)
	lockStockByResource func(*sql.Tx, int64) (*repository.StockLock, error)
	applyMovement       func(*sql.Tx, *domain.StockMovement) error
	list                func(domain.StockMovementFilter) ([]domain.StockMovement, error)
	listFuelStocks      func() ([]domain.FuelStock, error)
}

func (m *stockMovementRepoMock) LockStockByID(tx *sql.Tx, id int64) (*repository.StockLock, error) {
	if m.lockStockByID != nil {
		return m.lockStockByID(tx, id)
	}
	return nil, nil
}
func (m *stockMovementRepoMock) LockStockByResource(tx *sql.Tx, id int64) (*repository.StockLock, error) {
	if m.lockStockByResource != nil {
		return m.lockStockByResource(tx, id)
	}
	return nil, nil
}
func (m *stockMovementRepoMock) ApplyMovement(tx *sql.Tx, mv *domain.StockMovement) error {
	if m.applyMovement != nil {
		return m.applyMovement(tx, mv)
	}
	return nil
}
func (m *stockMovementRepoMock) List(f domain.StockMovementFilter) ([]domain.StockMovement, error) {
	if m.list != nil {
		return m.list(f)
	}
	return nil, nil
}
func (m *stockMovementRepoMock) ListFuelStocks() ([]domain.FuelStock, error) {
	if m.listFuelStocks != nil {
		return m.listFuelStocks()
	}
	return []domain.FuelStock{}, nil
}

// ─── Crops ───────────────────────────────────────────────────────────────────

type cropRepoMock struct {
	getSeasons              func() ([]domain.Season, error)
	getSeasonByID           func(int64) (*domain.Season, error)
	getActiveOrLatestSeason func() (*domain.Season, error)
	createSeason            func(*domain.Season) error
	updateSeason            func(int64, *domain.Season) error
	deleteSeason            func(int64) error
	getCrops                func() ([]domain.Crop, error)
	getCropByID             func(int64) (*domain.Crop, error)
	createCrop              func(*domain.Crop) error
	updateCrop              func(int64, *domain.Crop) error
	deleteCrop              func(int64) error
	listFieldCrops          func(domain.FieldCropFilter) ([]domain.FieldCrop, error)
	getFieldCropByID        func(int64) (*domain.FieldCrop, error)
	createFieldCrop         func(*domain.FieldCrop) error
	updateFieldCrop         func(int64, *domain.FieldCrop) error
	deleteFieldCrop         func(int64) error
	relinkOperations        func(int64) error
	lockFieldCropHarvest    func(*sql.Tx, int64) (*float64, float64, error)
	ensureHarvestResource   func(*sql.Tx, *domain.Crop) (int64, error)
	ensureStock             func(*sql.Tx, int64) error
	markHarvestRecorded     func(*sql.Tx, int64, float64) error
}

func (m *cropRepoMock) GetSeasons() ([]domain.Season, error) {
	if m.getSeasons != nil {
		return m.getSeasons()
	}
	return nil, nil
}
func (m *cropRepoMock) GetSeasonByID(id int64) (*domain.Season, error) {
	if m.getSeasonByID != nil {
		return m.getSeasonByID(id)
	}
	return nil, nil
}
func (m *cropRepoMock) GetActiveOrLatestSeason() (*domain.Season, error) {
	if m.getActiveOrLatestSeason != nil {
		return m.getActiveOrLatestSeason()
	}
	return nil, nil
}
func (m *cropRepoMock) CreateSeason(s *domain.Season) error {
	if m.createSeason != nil {
		return m.createSeason(s)
	}
	return nil
}
func (m *cropRepoMock) UpdateSeason(id int64, s *domain.Season) error {
	if m.updateSeason != nil {
		return m.updateSeason(id, s)
	}
	return nil
}
func (m *cropRepoMock) DeleteSeason(id int64) error {
	if m.deleteSeason != nil {
		return m.deleteSeason(id)
	}
	return nil
}
func (m *cropRepoMock) GetCrops() ([]domain.Crop, error) {
	if m.getCrops != nil {
		return m.getCrops()
	}
	return nil, nil
}
func (m *cropRepoMock) GetCropByID(id int64) (*domain.Crop, error) {
	if m.getCropByID != nil {
		return m.getCropByID(id)
	}
	return nil, nil
}
func (m *cropRepoMock) CreateCrop(c *domain.Crop) error {
	if m.createCrop != nil {
		return m.createCrop(c)
	}
	return nil
}
func (m *cropRepoMock) UpdateCrop(id int64, c *domain.Crop) error {
	if m.updateCrop != nil {
		return m.updateCrop(id, c)
	}
	return nil
}
func (m *cropRepoMock) DeleteCrop(id int64) error {
	if m.deleteCrop != nil {
		return m.deleteCrop(id)
	}
	return nil
}
func (m *cropRepoMock) ListFieldCrops(f domain.FieldCropFilter) ([]domain.FieldCrop, error) {
	if m.listFieldCrops != nil {
		return m.listFieldCrops(f)
	}
	return nil, nil
}
func (m *cropRepoMock) GetFieldCropByID(id int64) (*domain.FieldCrop, error) {
	if m.getFieldCropByID != nil {
		return m.getFieldCropByID(id)
	}
	return nil, nil
}
func (m *cropRepoMock) CreateFieldCrop(fc *domain.FieldCrop) error {
	if m.createFieldCrop != nil {
		return m.createFieldCrop(fc)
	}
	return nil
}
func (m *cropRepoMock) UpdateFieldCrop(id int64, fc *domain.FieldCrop) error {
	if m.updateFieldCrop != nil {
		return m.updateFieldCrop(id, fc)
	}
	return nil
}
func (m *cropRepoMock) DeleteFieldCrop(id int64) error {
	if m.deleteFieldCrop != nil {
		return m.deleteFieldCrop(id)
	}
	return nil
}
func (m *cropRepoMock) RelinkOperations(id int64) error {
	if m.relinkOperations != nil {
		return m.relinkOperations(id)
	}
	return nil
}
func (m *cropRepoMock) LockFieldCropHarvest(tx *sql.Tx, id int64) (*float64, float64, error) {
	if m.lockFieldCropHarvest != nil {
		return m.lockFieldCropHarvest(tx, id)
	}
	return nil, 0, nil
}
func (m *cropRepoMock) EnsureHarvestResource(tx *sql.Tx, c *domain.Crop) (int64, error) {
	if m.ensureHarvestResource != nil {
		return m.ensureHarvestResource(tx, c)
	}
	return 0, nil
}
func (m *cropRepoMock) EnsureStock(tx *sql.Tx, id int64) error {
	if m.ensureStock != nil {
		return m.ensureStock(tx, id)
	}
	return nil
}
func (m *cropRepoMock) MarkHarvestRecorded(tx *sql.Tx, id int64, qty float64) error {
	if m.markHarvestRecorded != nil {
		return m.markHarvestRecorded(tx, id, qty)
	}
	return nil
}

// ─── Weather snapshots / report subscriptions ────────────────────────────────

type weatherSnapshotRepoMock struct {
	insert             func(*domain.WeatherSnapshot) (bool, error)
	getDailyAggregates func(domain.ReportFilter) ([]domain.WeatherDailyAggregate, error)
	getLatestPerField  func() ([]domain.WeatherSnapshot, error)
	countInPeriod      func(domain.ReportFilter) (int, int, error)
}

func (m *weatherSnapshotRepoMock) Insert(s *domain.WeatherSnapshot) (bool, error) {
	if m.insert != nil {
		return m.insert(s)
	}
	return true, nil
}
func (m *weatherSnapshotRepoMock) GetDailyAggregates(f domain.ReportFilter) ([]domain.WeatherDailyAggregate, error) {
	if m.getDailyAggregates != nil {
		return m.getDailyAggregates(f)
	}
	return nil, nil
}
func (m *weatherSnapshotRepoMock) GetLatestPerField() ([]domain.WeatherSnapshot, error) {
	if m.getLatestPerField != nil {
		return m.getLatestPerField()
	}
	return nil, nil
}
func (m *weatherSnapshotRepoMock) CountInPeriod(f domain.ReportFilter) (int, int, error) {
	if m.countInPeriod != nil {
		return m.countInPeriod(f)
	}
	return 0, 0, nil
}

type reportSubscriptionRepoMock struct {
	getByUser            func(int64) (*domain.ReportSubscription, error)
	upsert               func(*domain.ReportSubscription) error
	deactivate           func(int64) error
	listActiveRecipients func() ([]domain.ReportSubscriptionRecipient, error)
	markSent             func(int64, time.Time) error
}

func (m *reportSubscriptionRepoMock) GetByUser(id int64) (*domain.ReportSubscription, error) {
	if m.getByUser != nil {
		return m.getByUser(id)
	}
	return nil, nil
}
func (m *reportSubscriptionRepoMock) Upsert(s *domain.ReportSubscription) error {
	if m.upsert != nil {
		return m.upsert(s)
	}
	return nil
}
func (m *reportSubscriptionRepoMock) Deactivate(id int64) error {
	if m.deactivate != nil {
		return m.deactivate(id)
	}
	return nil
}
func (m *reportSubscriptionRepoMock) ListActiveRecipients() ([]domain.ReportSubscriptionRecipient, error) {
	if m.listActiveRecipients != nil {
		return m.listActiveRecipients()
	}
	return nil, nil
}
func (m *reportSubscriptionRepoMock) MarkSent(id int64, at time.Time) error {
	if m.markSent != nil {
		return m.markSent(id, at)
	}
	return nil
}

// ─── Reports ─────────────────────────────────────────────────────────────────

type reportRepoMock struct {
	getOperationsMetrics     func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error)
	getInventorySnapshot     func() (*domain.ReportInventorySnapshot, error)
	getOperationsTimeline    func(domain.ReportFilter, string) ([]domain.ReportTimeBucket, error)
	getOperationsByType      func(domain.ReportFilter) ([]domain.ReportOperationTypeStat, error)
	getOperationRows         func(domain.ReportFilter, int) ([]domain.ReportOperationRow, error)
	getFieldRows             func(domain.ReportFilter) ([]domain.ReportFieldRow, error)
	getMachineStatusCounts   func() ([]domain.ReportNamedCount, error)
	getImplementStatusCounts func() ([]domain.ReportNamedCount, error)
	getMachinesByType        func() ([]domain.ReportNamedCount, error)
	getMachinesByFuel        func() ([]domain.ReportNamedCount, error)
	getMachinesByYear        func() ([]domain.ReportNamedCount, error)
	getMachineRows           func(domain.ReportFilter) ([]domain.ReportMachineRow, error)
	getImplementRows         func(domain.ReportFilter) ([]domain.ReportImplementRow, error)
	getOperatorRows          func(domain.ReportFilter) ([]domain.ReportOperatorRow, error)
	getStockRows             func() ([]domain.ReportStockRow, error)
	getEstimatedConsumption  func(domain.ReportFilter) ([]domain.ReportResourceConsumption, error)
	getRealConsumption       func(domain.ReportFilter) ([]domain.ReportResourceConsumption, error)
	getMovementTotals        func(domain.ReportFilter) (*domain.ReportMovementTotals, error)
	getFieldCropRows         func(int64, string) ([]domain.ReportFieldCropRow, error)
}

func (m *reportRepoMock) GetOperationsMetrics(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
	if m.getOperationsMetrics != nil {
		return m.getOperationsMetrics(f)
	}
	return &domain.ReportOperationsMetrics{}, nil
}
func (m *reportRepoMock) GetInventorySnapshot() (*domain.ReportInventorySnapshot, error) {
	if m.getInventorySnapshot != nil {
		return m.getInventorySnapshot()
	}
	return &domain.ReportInventorySnapshot{}, nil
}
func (m *reportRepoMock) GetOperationsTimeline(f domain.ReportFilter, g string) ([]domain.ReportTimeBucket, error) {
	if m.getOperationsTimeline != nil {
		return m.getOperationsTimeline(f, g)
	}
	return nil, nil
}
func (m *reportRepoMock) GetOperationsByType(f domain.ReportFilter) ([]domain.ReportOperationTypeStat, error) {
	if m.getOperationsByType != nil {
		return m.getOperationsByType(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetOperationRows(f domain.ReportFilter, limit int) ([]domain.ReportOperationRow, error) {
	if m.getOperationRows != nil {
		return m.getOperationRows(f, limit)
	}
	return nil, nil
}
func (m *reportRepoMock) GetFieldRows(f domain.ReportFilter) ([]domain.ReportFieldRow, error) {
	if m.getFieldRows != nil {
		return m.getFieldRows(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetMachineStatusCounts() ([]domain.ReportNamedCount, error) {
	if m.getMachineStatusCounts != nil {
		return m.getMachineStatusCounts()
	}
	return nil, nil
}
func (m *reportRepoMock) GetImplementStatusCounts() ([]domain.ReportNamedCount, error) {
	if m.getImplementStatusCounts != nil {
		return m.getImplementStatusCounts()
	}
	return nil, nil
}
func (m *reportRepoMock) GetMachinesByType() ([]domain.ReportNamedCount, error) {
	if m.getMachinesByType != nil {
		return m.getMachinesByType()
	}
	return nil, nil
}
func (m *reportRepoMock) GetMachinesByFuel() ([]domain.ReportNamedCount, error) {
	if m.getMachinesByFuel != nil {
		return m.getMachinesByFuel()
	}
	return nil, nil
}
func (m *reportRepoMock) GetMachinesByYear() ([]domain.ReportNamedCount, error) {
	if m.getMachinesByYear != nil {
		return m.getMachinesByYear()
	}
	return nil, nil
}
func (m *reportRepoMock) GetMachineRows(f domain.ReportFilter) ([]domain.ReportMachineRow, error) {
	if m.getMachineRows != nil {
		return m.getMachineRows(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetImplementRows(f domain.ReportFilter) ([]domain.ReportImplementRow, error) {
	if m.getImplementRows != nil {
		return m.getImplementRows(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetOperatorRows(f domain.ReportFilter) ([]domain.ReportOperatorRow, error) {
	if m.getOperatorRows != nil {
		return m.getOperatorRows(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetStockRows() ([]domain.ReportStockRow, error) {
	if m.getStockRows != nil {
		return m.getStockRows()
	}
	return nil, nil
}
func (m *reportRepoMock) GetEstimatedConsumption(f domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	if m.getEstimatedConsumption != nil {
		return m.getEstimatedConsumption(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetRealConsumption(f domain.ReportFilter) ([]domain.ReportResourceConsumption, error) {
	if m.getRealConsumption != nil {
		return m.getRealConsumption(f)
	}
	return nil, nil
}
func (m *reportRepoMock) GetMovementTotals(f domain.ReportFilter) (*domain.ReportMovementTotals, error) {
	if m.getMovementTotals != nil {
		return m.getMovementTotals(f)
	}
	return &domain.ReportMovementTotals{}, nil
}
func (m *reportRepoMock) GetFieldCropRows(seasonID int64, fieldID string) ([]domain.ReportFieldCropRow, error) {
	if m.getFieldCropRows != nil {
		return m.getFieldCropRows(seasonID, fieldID)
	}
	return nil, nil
}
