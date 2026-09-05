package usecase

import (
	"agri-api/internal/domain"
	"agri-api/internal/dto"
	"agri-api/internal/repository"
	"context"
	"fmt"
	"strconv"
	"time"
)

// OverdueMonitorLogger este subsetul de logging folosit de monitor (satisfăcut de *logger.Logger).
type OverdueMonitorLogger interface {
	Infof(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

// FieldOperationOverdueMonitor verifică periodic operațiunile pe teren aflate în lucru
// care au depășit sfârșitul planificat (timpul estimat de lucru) și notifică
// administratorii/managerii și operatorul asignat. O operațiune este notificată o singură
// dată; marcajul se resetează automat dacă sfârșitul planificat este modificat.
type FieldOperationOverdueMonitor struct {
	repo     repository.FieldOperationRepository
	notif    *NotificationService
	audit    *AuditService
	log      OverdueMonitorLogger
	interval time.Duration
}

func NewFieldOperationOverdueMonitor(
	repo repository.FieldOperationRepository,
	notif *NotificationService,
	audit *AuditService,
	log OverdueMonitorLogger,
	interval time.Duration,
) *FieldOperationOverdueMonitor {
	if interval <= 0 {
		interval = time.Minute
	}
	return &FieldOperationOverdueMonitor{
		repo:     repo,
		notif:    notif,
		audit:    audit,
		log:      log,
		interval: interval,
	}
}

// Start rulează o verificare imediat și apoi la fiecare interval, până când ctx este anulat.
func (m *FieldOperationOverdueMonitor) Start(ctx context.Context) {
	m.log.Infof("Field operation overdue monitor started (interval %s)", m.interval)
	m.runSafely(time.Now())

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.log.Infof("Field operation overdue monitor stopped")
			return
		case now := <-ticker.C:
			m.runSafely(now)
		}
	}
}

func (m *FieldOperationOverdueMonitor) runSafely(now time.Time) {
	notified, err := m.RunOnce(now)
	if err != nil {
		m.log.Errorf("Field operation overdue check failed: %v", err)
		return
	}
	if notified > 0 {
		m.log.Infof("Field operation overdue check: %d operation(s) notified", notified)
	}
}

// RunOnce face o singură verificare și returnează numărul de operațiuni notificate.
func (m *FieldOperationOverdueMonitor) RunOnce(now time.Time) (int, error) {
	overdue, err := m.repo.GetOverdueInProgress(now)
	if err != nil {
		return 0, err
	}

	notified := 0
	for _, op := range overdue {
		if err := m.notifyOverdue(op, now); err != nil {
			m.log.Errorf("Field operation #%d: overdue notification failed: %v", op.ID, err)
			continue
		}
		if err := m.repo.MarkOverdueNotified(op.ID, now); err != nil {
			m.log.Errorf("Field operation #%d: could not mark overdue notification as sent: %v", op.ID, err)
			continue
		}
		notified++
	}
	return notified, nil
}

func (m *FieldOperationOverdueMonitor) notifyOverdue(op dto.OverdueFieldOperation, now time.Time) error {
	entityID := strconv.FormatInt(op.ID, 10)
	exceededBy := now.Sub(op.PlannedEndAt)

	if m.notif != nil {
		var recipients []int64
		if op.OperatorUserID != nil {
			recipients = append(recipients, *op.OperatorUserID)
		}
		if err := m.notif.EmitToAudience(
			domain.NotifOperationOverdue,
			"Timp estimat de lucru depășit",
			buildOverdueMessage(op, exceededBy),
			"field_operation", entityID,
			recipients...,
		); err != nil {
			return err
		}
	}

	if m.audit != nil {
		m.audit.Log("field_operation", entityID, "overdue", nil, map[string]interface{}{
			"status":              string(domain.FieldOperationStatusInProgress),
			"planned_end_at":      op.PlannedEndAt,
			"exceeded_by_minutes": int(exceededBy / time.Minute),
		})
	}
	return nil
}

func buildOverdueMessage(op dto.OverdueFieldOperation, exceededBy time.Duration) string {
	operatorName := "neatribuit"
	if op.OperatorName != nil && *op.OperatorName != "" {
		operatorName = *op.OperatorName
	}

	if op.PlannedStartAt != nil && op.PlannedEndAt.After(*op.PlannedStartAt) {
		return fmt.Sprintf(
			"Lucrarea „%s” pe terenul „%s” a depășit cu %s timpul estimat de lucru de %s. Operator: %s.",
			op.OperationTypeName,
			op.FieldName,
			FormatWorkDuration(exceededBy),
			FormatWorkDuration(op.PlannedEndAt.Sub(*op.PlannedStartAt)),
			operatorName,
		)
	}
	return fmt.Sprintf(
		"Lucrarea „%s” pe terenul „%s” a depășit cu %s sfârșitul planificat. Operator: %s.",
		op.OperationTypeName,
		op.FieldName,
		FormatWorkDuration(exceededBy),
		operatorName,
	)
}

// FormatWorkDuration afișează o durată doar în ore și minute (fără zile), ex: "26 h 30 min".
func FormatWorkDuration(d time.Duration) string {
	totalMinutes := int(d.Round(time.Minute) / time.Minute)
	if totalMinutes <= 0 {
		return "sub 1 min"
	}
	hours := totalMinutes / 60
	minutes := totalMinutes % 60
	switch {
	case hours == 0:
		return fmt.Sprintf("%d min", minutes)
	case minutes == 0:
		return fmt.Sprintf("%d h", hours)
	default:
		return fmt.Sprintf("%d h %d min", hours, minutes)
	}
}
