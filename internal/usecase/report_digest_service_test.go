package usecase

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"agri-api/internal/domain"
)

type mailerMock struct {
	mu      sync.Mutex
	sent    []string // "to|subject|body"
	failFor string
}

func (m *mailerMock) SendHTML(to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if to == m.failFor {
		return errors.New("smtp down")
	}
	m.sent = append(m.sent, to+"|"+subject+"|"+body)
	return nil
}

func (m *mailerMock) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sent)
}

func newDigestService(repo *reportSubscriptionRepoMock, mailer *mailerMock, log *testLogger) *ReportDigestService {
	reports := NewReportService(&reportRepoMock{})
	svc := NewReportDigestService(repo, reports, mailer, "http://front", log, 0)
	svc.now = func() time.Time { return time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC) }
	return svc
}

func TestReportDigestService_Subscription(t *testing.T) {
	var saved *domain.ReportSubscription
	repo := &reportSubscriptionRepoMock{
		upsert:    func(s *domain.ReportSubscription) error { saved = s; return nil },
		getByUser: func(id int64) (*domain.ReportSubscription, error) { return saved, nil },
	}
	svc := newDigestService(repo, &mailerMock{}, &testLogger{})
	if svc.interval != time.Minute {
		t.Errorf("intervalul implicit trebuie să fie 1m, am %s", svc.interval)
	}

	if _, err := svc.UpsertSubscription(&domain.ReportSubscription{Frequency: "yearly"}); !errors.Is(err, ErrReportSubscriptionInvalid) {
		t.Errorf("frecvență invalidă: %v", err)
	}
	if _, err := svc.UpsertSubscription(&domain.ReportSubscription{Frequency: domain.ReportFrequencyDaily, SendHour: 24}); !errors.Is(err, ErrReportSubscriptionInvalid) {
		t.Errorf("oră invalidă: %v", err)
	}
	result, err := svc.UpsertSubscription(&domain.ReportSubscription{UserID: 1, Frequency: domain.ReportFrequencyWeekly, SendHour: 7, Weekday: 9})
	if err != nil || result.Weekday != 1 || saved.UserID != 1 {
		t.Errorf("UpsertSubscription: %v, %+v", err, result)
	}
	if sub, err := svc.GetSubscription(1); err != nil || sub != saved {
		t.Errorf("GetSubscription: %v", err)
	}
	if err := svc.Deactivate(1); err != nil {
		t.Errorf("Deactivate: %v", err)
	}
	repo.upsert = func(*domain.ReportSubscription) error { return errors.New("db down") }
	if _, err := svc.UpsertSubscription(&domain.ReportSubscription{Frequency: domain.ReportFrequencyDaily, SendHour: 7}); err == nil {
		t.Error("eroarea de salvare trebuie propagată")
	}
}

func TestIsDigestDue(t *testing.T) {
	// duminică 15 martie 2026, ora 09:30
	now := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	sentToday := now.Add(-10 * time.Minute) // după ora programată (09:00)
	yesterday := now.AddDate(0, 0, -1)
	active := func(freq domain.ReportFrequency, hour, weekday int, last *time.Time) domain.ReportSubscription {
		return domain.ReportSubscription{IsActive: true, Frequency: freq, SendHour: hour, Weekday: weekday, LastSentAt: last}
	}
	cases := map[string]struct {
		sub  domain.ReportSubscription
		want bool
	}{
		"inactiv":              {domain.ReportSubscription{Frequency: domain.ReportFrequencyDaily}, false},
		"înainte de oră":       {active(domain.ReportFrequencyDaily, 10, 1, nil), false},
		"zilnic, netrimis":     {active(domain.ReportFrequencyDaily, 9, 1, nil), true},
		"zilnic, trimis ieri":  {active(domain.ReportFrequencyDaily, 9, 1, &yesterday), true},
		"zilnic, trimis azi":   {active(domain.ReportFrequencyDaily, 9, 1, &sentToday), false},
		"săptămânal, altă zi":  {active(domain.ReportFrequencyWeekly, 9, 1, nil), false},
		"săptămânal, duminică": {active(domain.ReportFrequencyWeekly, 9, 7, nil), true},
		"lunar, nu e ziua 1":   {active(domain.ReportFrequencyMonthly, 9, 1, nil), false},
	}
	for name, c := range cases {
		if got := isDigestDue(c.sub, now); got != c.want {
			t.Errorf("%s: %v", name, got)
		}
	}
	firstOfMonth := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	if !isDigestDue(active(domain.ReportFrequencyMonthly, 9, 1, nil), firstOfMonth) {
		t.Error("lunar pe 1 trebuie să fie scadent")
	}
	if isoWeekday(now) != 7 || isoWeekday(firstOfMonth) != 3 {
		t.Error("isoWeekday: duminică=7, miercuri=3")
	}
}

func TestDigestPeriodAndDates(t *testing.T) {
	now := time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC)
	if from, to, kind := digestPeriod(domain.ReportFrequencyDaily, now); from != "2026-03-14" || to != "2026-03-14" || kind != "zilnic" {
		t.Errorf("zilnic: %s %s %s", from, to, kind)
	}
	if from, to, kind := digestPeriod(domain.ReportFrequencyWeekly, now); from != "2026-03-08" || to != "2026-03-14" || kind != "săptămânal" {
		t.Errorf("săptămânal: %s %s %s", from, to, kind)
	}
	if from, to, kind := digestPeriod(domain.ReportFrequencyMonthly, now); from != "2026-02-01" || to != "2026-02-28" || kind != "lunar" {
		t.Errorf("lunar: %s %s %s", from, to, kind)
	}
	if formatRomanianDate("2026-03-14") != "14.03.2026" || formatRomanianDate("x") != "x" {
		t.Error("formatRomanianDate")
	}
}

func TestReportDigestService_SendNow(t *testing.T) {
	marked := int64(0)
	repo := &reportSubscriptionRepoMock{markSent: func(id int64, _ time.Time) error { marked = id; return nil }}
	mailer := &mailerMock{}
	svc := newDigestService(repo, mailer, &testLogger{})

	// fără abonament → raport zilnic implicit, fără MarkSent
	if err := svc.SendNow(1, "ana@x.ro", "Ana"); err != nil || mailer.count() != 1 || marked != 0 {
		t.Fatalf("SendNow implicit: %v, trimise=%d, marcat=%d", err, mailer.count(), marked)
	}
	msg := mailer.sent[0]
	if !strings.Contains(msg, "zilnic") || !strings.Contains(msg, "Bună, Ana") || !strings.Contains(msg, "http://front/reports?from=2026-03-14") {
		t.Errorf("conținut e-mail greșit: %s", msg[:200])
	}

	repo.getByUser = func(int64) (*domain.ReportSubscription, error) {
		return &domain.ReportSubscription{ID: 5, UserID: 1, Frequency: domain.ReportFrequencyWeekly, SendHour: 7, IsActive: true}, nil
	}
	if err := svc.SendNow(1, "ana@x.ro", ""); err != nil || marked != 5 || !strings.Contains(mailer.sent[1], "săptămânal") {
		t.Errorf("SendNow cu abonament: %v, marcat=%d", err, marked)
	}

	repo.markSent = func(int64, time.Time) error { return errors.New("db down") }
	if err := svc.SendNow(1, "ana@x.ro", ""); err == nil {
		t.Error("eroarea la marcare trebuie propagată")
	}
	mailer.failFor = "ana@x.ro"
	if err := svc.SendNow(1, "ana@x.ro", ""); err == nil {
		t.Error("eroarea SMTP trebuie propagată")
	}
	repo.getByUser = func(int64) (*domain.ReportSubscription, error) { return nil, errors.New("db down") }
	if err := svc.SendNow(1, "ana@x.ro", ""); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}

	broken := NewReportDigestService(repo, NewReportService(&reportRepoMock{
		getOperationsMetrics: func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) { return nil, errors.New("db down") },
	}), mailer, "", &testLogger{}, time.Minute)
	repo.getByUser = nil
	if err := broken.SendNow(1, "b@x.ro", ""); err == nil {
		t.Error("eroarea raportului trebuie propagată")
	}
}

func TestReportDigestService_RunOnceAndStart(t *testing.T) {
	log := &testLogger{}
	mailer := &mailerMock{failFor: "fail@x.ro"}
	var listed int32
	repo := &reportSubscriptionRepoMock{listActiveRecipients: func() ([]domain.ReportSubscriptionRecipient, error) {
		atomic.AddInt32(&listed, 1)
		daily := domain.ReportSubscription{ID: 1, IsActive: true, Frequency: domain.ReportFrequencyDaily}
		return []domain.ReportSubscriptionRecipient{
			{ReportSubscription: daily, Email: "ok@x.ro"},
			{ReportSubscription: daily, Email: "fail@x.ro"},
			{ReportSubscription: domain.ReportSubscription{ID: 2, IsActive: false}, Email: "inactiv@x.ro"},
		}, nil
	}}
	svc := newDigestService(repo, mailer, log)

	sent, err := svc.RunOnce(time.Date(2026, 3, 15, 9, 30, 0, 0, time.UTC))
	if err != nil || sent != 1 || log.errorCount() != 1 {
		t.Fatalf("RunOnce: %v, trimise=%d, erori=%d", err, sent, log.errorCount())
	}

	svc.interval = 5 * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	svc.Start(ctx)
	if atomic.LoadInt32(&listed) < 2 {
		t.Error("Start trebuie să ruleze verificarea periodic")
	}
	log.mu.Lock()
	stopped := strings.Contains(strings.Join(log.infos, "\n"), "stopped")
	log.mu.Unlock()
	if !stopped {
		t.Error("oprirea monitorului trebuie logată")
	}

	repo.listActiveRecipients = func() ([]domain.ReportSubscriptionRecipient, error) { return nil, errors.New("db down") }
	if _, err := svc.RunOnce(time.Now()); err == nil {
		t.Error("eroarea de listare trebuie propagată")
	}
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	svc.Start(ctx)
	if log.errorCount() < 2 {
		t.Error("eroarea din rulare trebuie logată")
	}
}
