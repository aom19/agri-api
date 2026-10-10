package usecase_test

import (
	"errors"
	"strings"
	"testing"

	"agri-api/internal/domain"
	"agri-api/internal/usecase"
)

type mailerMock struct {
	sent []string // "to|subject|body"
	fail bool
}

func (m *mailerMock) SendHTML(to, subject, body string) error {
	if m.fail {
		return errors.New("smtp down")
	}
	m.sent = append(m.sent, to+"|"+subject+"|"+body)
	return nil
}

func TestReportEmailService_SendSummary(t *testing.T) {
	var filters []domain.ReportFilter
	repo := &reportRepoMock{getOperationsMetrics: func(f domain.ReportFilter) (*domain.ReportOperationsMetrics, error) {
		filters = append(filters, f)
		return &domain.ReportOperationsMetrics{OperationsTotal: 4, OperationsCompleted: 3}, nil
	}}
	mailer := &mailerMock{}
	svc := usecase.NewReportEmailService(usecase.NewReportService(repo), mailer, "http://front")

	// filtrele din pagină nu ajung în e-mail: acesta nu le afișează
	query := usecase.ReportQuery{From: "2026-03-01", To: "2026-03-14", FieldID: "F-1", MachineID: 7}
	if err := svc.SendSummary(query, "ana@x.ro", "Ana"); err != nil || len(mailer.sent) != 1 {
		t.Fatalf("SendSummary: %v, trimise=%d", err, len(mailer.sent))
	}
	if filters[0].FieldID != "" || filters[0].MachineID != 0 {
		t.Errorf("sumarul trimis nu trebuie filtrat: %+v", filters[0])
	}
	msg := mailer.sent[0]
	for _, want := range []string{
		"ana@x.ro|AgriERP · Raport 01.03.2026 – 14.03.2026|",
		"Bună, Ana",
		"4 (finalizate 3",
		"http://front/reports?from=2026-03-01&amp;to=2026-03-14",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("e-mailul nu conține %q", want)
		}
	}

	if err := svc.SendSummary(usecase.ReportQuery{From: "2026-03-14", To: "2026-03-01"}, "ana@x.ro", ""); !errors.Is(err, usecase.ErrInvalidReportPeriod) {
		t.Errorf("perioada inversată trebuie respinsă: %v", err)
	}
	mailer.fail = true
	if err := svc.SendSummary(query, "ana@x.ro", ""); err == nil {
		t.Error("eroarea SMTP trebuie propagată")
	}
	repo.getOperationsMetrics = func(domain.ReportFilter) (*domain.ReportOperationsMetrics, error) { return nil, errors.New("db down") }
	if err := svc.SendSummary(query, "ana@x.ro", ""); err == nil {
		t.Error("eroarea raportului trebuie propagată")
	}
}

func TestFormatRomanianDate(t *testing.T) {
	if usecase.FormatRomanianDate("2026-03-14") != "14.03.2026" || usecase.FormatRomanianDate("x") != "x" {
		t.Error("formatRomanianDate")
	}
}
