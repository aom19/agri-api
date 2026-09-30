package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"agri-api/internal/dto"
)

func TestFormatWorkDuration(t *testing.T) {
	cases := map[time.Duration]string{
		0:                             "sub 1 min",
		20 * time.Second:              "sub 1 min",
		45 * time.Minute:              "45 min",
		2 * time.Hour:                 "2 h",
		90 * time.Minute:              "1 h 30 min",
		26*time.Hour + 30*time.Minute: "26 h 30 min",
	}
	for d, want := range cases {
		if got := FormatWorkDuration(d); got != want {
			t.Errorf("%s: %q, vreau %q", d, got, want)
		}
	}
}

func TestBuildOverdueMessage(t *testing.T) {
	end := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	start := end.Add(-3 * time.Hour)
	op := dto.OverdueFieldOperation{ID: 1, FieldName: "Lot 1", OperationTypeName: "Arat", PlannedStartAt: &start, PlannedEndAt: end, OperatorName: ptr("Ion")}
	msg := buildOverdueMessage(op, 30*time.Minute)
	if !strings.Contains(msg, "timpul estimat de lucru de 3 h") || !strings.Contains(msg, "cu 30 min") || !strings.Contains(msg, "Operator: Ion") {
		t.Errorf("mesaj cu durată estimată: %s", msg)
	}
	op.PlannedStartAt = nil
	op.OperatorName = nil
	msg = buildOverdueMessage(op, time.Hour)
	if !strings.Contains(msg, "sfârșitul planificat") || !strings.Contains(msg, "neatribuit") {
		t.Errorf("mesaj fără start planificat: %s", msg)
	}
}

func TestFieldOperationOverdueMonitor_RunOnce(t *testing.T) {
	now := time.Now()
	overdue := []dto.OverdueFieldOperation{
		{ID: 1, FieldName: "Lot 1", OperationTypeName: "Arat", PlannedEndAt: now.Add(-time.Hour), OperatorUserID: ptr(int64(7))},
		{ID: 2, FieldName: "Lot 2", OperationTypeName: "Semănat", PlannedEndAt: now.Add(-2 * time.Hour)},
	}
	var marked []int64
	repo := &fieldOperationRepoMock{
		getOverdueInProgress: func(time.Time) ([]dto.OverdueFieldOperation, error) { return overdue, nil },
		markOverdueNotified:  func(id int64, _ time.Time) error { marked = append(marked, id); return nil },
	}
	notifRepo := &notificationRepoMock{adminIDs: []int64{1}}
	auditRepo := &auditRepoMock{}
	log := &testLogger{}
	monitor := NewFieldOperationOverdueMonitor(repo, NewNotificationService(notifRepo), NewAuditService(auditRepo), log, 0)
	if monitor.interval != time.Minute {
		t.Errorf("intervalul implicit trebuie să fie 1m, am %s", monitor.interval)
	}

	notified, err := monitor.RunOnce(now)
	if err != nil || notified != 2 || len(marked) != 2 {
		t.Fatalf("RunOnce: %v, notificate=%d, marcate=%v", err, notified, marked)
	}
	if notifRepo.deliveries(1) != 2 || notifRepo.deliveries(7) != 1 {
		t.Errorf("livrări: admin=%d operator=%d", notifRepo.deliveries(1), notifRepo.deliveries(7))
	}
	waitFor(t, func() bool { return auditRepo.count() == 2 })

	// marcarea eșuează pentru a doua operațiune
	repo.markOverdueNotified = func(id int64, _ time.Time) error {
		if id == 2 {
			return errors.New("db down")
		}
		return nil
	}
	if notified, err = monitor.RunOnce(now); err != nil || notified != 1 || log.errorCount() != 1 {
		t.Errorf("marcare eșuată: %v, notificate=%d, erori=%d", err, notified, log.errorCount())
	}

	// notificarea eșuează
	notifRepo.adminErr = errors.New("db down")
	if notified, err = monitor.RunOnce(now); err != nil || notified != 0 {
		t.Errorf("notificare eșuată: %v, notificate=%d", err, notified)
	}

	// fără notificări și audit configurate, operațiunile sunt doar marcate
	repo.markOverdueNotified = nil
	bare := NewFieldOperationOverdueMonitor(repo, nil, nil, log, time.Minute)
	if notified, err = bare.RunOnce(now); err != nil || notified != 2 {
		t.Errorf("monitor fără notificări: %v, %d", err, notified)
	}

	repo.getOverdueInProgress = func(time.Time) ([]dto.OverdueFieldOperation, error) { return nil, errors.New("db down") }
	if _, err := monitor.RunOnce(now); err == nil {
		t.Error("eroarea de citire trebuie propagată")
	}
}

func TestFieldOperationOverdueMonitor_Start(t *testing.T) {
	calls := 0
	repo := &fieldOperationRepoMock{getOverdueInProgress: func(time.Time) ([]dto.OverdueFieldOperation, error) {
		calls++
		if calls == 1 {
			return []dto.OverdueFieldOperation{{ID: 1, PlannedEndAt: time.Now()}}, nil
		}
		return nil, errors.New("db down")
	}}
	log := &testLogger{}
	monitor := NewFieldOperationOverdueMonitor(repo, nil, nil, log, 5*time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	monitor.Start(ctx)

	if calls < 2 || log.errorCount() == 0 {
		t.Errorf("monitorul trebuie să ruleze imediat și apoi periodic: apeluri=%d erori=%d", calls, log.errorCount())
	}
	joined := strings.Join(log.infos, "\n")
	if !strings.Contains(joined, "1 operation(s) notified") || !strings.Contains(joined, "stopped") {
		t.Errorf("log-uri lipsă: %s", joined)
	}
}
