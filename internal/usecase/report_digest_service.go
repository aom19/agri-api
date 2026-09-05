package usecase

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"time"

	"agri-api/internal/domain"
	"agri-api/internal/repository"
)

var ErrReportSubscriptionInvalid = errors.New("abonament invalid")

// DigestMailer este subsetul serviciului de e-mail folosit pentru rapoartele programate.
type DigestMailer interface {
	SendHTML(to, subject, htmlBody string) error
}

// ReportDigestService administrează abonamentele la raportul sumar și trimite e-mailurile programate.
type ReportDigestService struct {
	repo         repository.ReportSubscriptionRepository
	reports      *ReportService
	mailer       DigestMailer
	clientOrigin string
	log          OverdueMonitorLogger
	interval     time.Duration
	now          func() time.Time
}

func NewReportDigestService(
	repo repository.ReportSubscriptionRepository,
	reports *ReportService,
	mailer DigestMailer,
	clientOrigin string,
	log OverdueMonitorLogger,
	interval time.Duration,
) *ReportDigestService {
	if interval <= 0 {
		interval = time.Minute
	}
	return &ReportDigestService{
		repo:         repo,
		reports:      reports,
		mailer:       mailer,
		clientOrigin: clientOrigin,
		log:          log,
		interval:     interval,
		now:          time.Now,
	}
}

func (service *ReportDigestService) GetSubscription(userID int64) (*domain.ReportSubscription, error) {
	return service.repo.GetByUser(userID)
}

func (service *ReportDigestService) UpsertSubscription(subscription *domain.ReportSubscription) (*domain.ReportSubscription, error) {
	switch subscription.Frequency {
	case domain.ReportFrequencyDaily, domain.ReportFrequencyWeekly, domain.ReportFrequencyMonthly:
	default:
		return nil, fmt.Errorf("%w: frecvența trebuie să fie daily, weekly sau monthly", ErrReportSubscriptionInvalid)
	}
	if subscription.SendHour < 0 || subscription.SendHour > 23 {
		return nil, fmt.Errorf("%w: ora trebuie să fie între 0 și 23", ErrReportSubscriptionInvalid)
	}
	if subscription.Weekday < 1 || subscription.Weekday > 7 {
		subscription.Weekday = 1
	}
	if err := service.repo.Upsert(subscription); err != nil {
		return nil, err
	}
	return service.repo.GetByUser(subscription.UserID)
}

func (service *ReportDigestService) Deactivate(userID int64) error {
	return service.repo.Deactivate(userID)
}

// SendNow trimite imediat raportul pentru abonamentul utilizatorului (sau unul zilnic implicit).
func (service *ReportDigestService) SendNow(userID int64, email, firstName string) error {
	subscription, err := service.repo.GetByUser(userID)
	if err != nil {
		return err
	}
	if subscription == nil {
		subscription = &domain.ReportSubscription{UserID: userID, Frequency: domain.ReportFrequencyDaily, SendHour: 7, Weekday: 1, IsActive: false}
	}
	recipient := domain.ReportSubscriptionRecipient{ReportSubscription: *subscription, Email: email, FirstName: firstName}
	return service.send(recipient, service.now())
}

// Start rulează verificarea abonamentelor la fiecare interval, până când ctx este anulat.
func (service *ReportDigestService) Start(ctx context.Context) {
	service.log.Infof("Report digest monitor started (interval %s)", service.interval)
	ticker := time.NewTicker(service.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			service.log.Infof("Report digest monitor stopped")
			return
		case now := <-ticker.C:
			sent, err := service.RunOnce(now)
			if err != nil {
				service.log.Errorf("Report digest run failed: %v", err)
			} else if sent > 0 {
				service.log.Infof("Report digest run: %d e-mail(s) sent", sent)
			}
		}
	}
}

// RunOnce trimite rapoartele scadente și returnează numărul de e-mailuri trimise.
func (service *ReportDigestService) RunOnce(now time.Time) (int, error) {
	recipients, err := service.repo.ListActiveRecipients()
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, recipient := range recipients {
		if !isDigestDue(recipient.ReportSubscription, now) {
			continue
		}
		if err := service.send(recipient, now); err != nil {
			service.log.Errorf("Report digest for %s failed: %v", recipient.Email, err)
			continue
		}
		sent++
	}
	return sent, nil
}

// isDigestDue verifică dacă a trecut ora programată de azi (în fusul serverului) și dacă
// raportul nu a fost trimis deja pentru intervalul curent.
func isDigestDue(subscription domain.ReportSubscription, now time.Time) bool {
	if !subscription.IsActive {
		return false
	}
	scheduled := time.Date(now.Year(), now.Month(), now.Day(), subscription.SendHour, 0, 0, 0, now.Location())
	if now.Before(scheduled) {
		return false
	}
	switch subscription.Frequency {
	case domain.ReportFrequencyWeekly:
		if isoWeekday(now) != subscription.Weekday {
			return false
		}
	case domain.ReportFrequencyMonthly:
		if now.Day() != 1 {
			return false
		}
	}
	return subscription.LastSentAt == nil || subscription.LastSentAt.Before(scheduled)
}

func isoWeekday(value time.Time) int {
	weekday := int(value.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

// digestPeriod întoarce intervalul raportat pentru frecvența dată: ieri, ultimele 7 zile sau luna trecută.
func digestPeriod(frequency domain.ReportFrequency, now time.Time) (string, string, string) {
	yesterday := now.AddDate(0, 0, -1)
	switch frequency {
	case domain.ReportFrequencyWeekly:
		return now.AddDate(0, 0, -7).Format("2006-01-02"), yesterday.Format("2006-01-02"), "săptămânal"
	case domain.ReportFrequencyMonthly:
		firstOfThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		lastMonthEnd := firstOfThisMonth.AddDate(0, 0, -1)
		lastMonthStart := time.Date(lastMonthEnd.Year(), lastMonthEnd.Month(), 1, 0, 0, 0, 0, now.Location())
		return lastMonthStart.Format("2006-01-02"), lastMonthEnd.Format("2006-01-02"), "lunar"
	default:
		return yesterday.Format("2006-01-02"), yesterday.Format("2006-01-02"), "zilnic"
	}
}

var digestTemplate = template.Must(template.New("digest").Parse(`
<!DOCTYPE html>
<html lang="ro"><body style="margin:0;padding:24px;background:#f4f6f4;font-family:Arial,Helvetica,sans-serif;color:#0d1f17;">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:640px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;">
  <tr><td style="background:#1a5c38;color:#ffffff;padding:20px 24px;">
    <div style="font-size:12px;opacity:.85;text-transform:uppercase;letter-spacing:.08em;">AgriERP · raport {{.Kind}}</div>
    <div style="font-size:20px;font-weight:700;margin-top:4px;">Sumar {{.From}} – {{.To}}</div>
  </td></tr>
  <tr><td style="padding:20px 24px;">
    <p style="margin:0 0 16px;">Bună{{if .Name}}, {{.Name}}{{end}}! Iată sinteza activității din perioada raportată.</p>
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:separate;border-spacing:0 8px;">
      {{range .Rows}}
      <tr>
        <td style="padding:10px 12px;background:#f4f6f4;border-radius:8px 0 0 8px;font-size:13px;color:#4a5e54;">{{.Label}}</td>
        <td style="padding:10px 12px;background:#f4f6f4;border-radius:0 8px 8px 0;font-size:15px;font-weight:700;text-align:right;white-space:nowrap;">{{.Value}}</td>
      </tr>
      {{end}}
    </table>
    <p style="margin:20px 0 0;">
      <a href="{{.Link}}" style="display:inline-block;background:#1a5c38;color:#ffffff;text-decoration:none;padding:10px 16px;border-radius:8px;font-weight:600;">Deschide raportul complet</a>
    </p>
    <p style="margin:16px 0 0;font-size:12px;color:#4a5e54;">Primești acest e-mail pentru că ai activat rapoartele programate în pagina Rapoarte. Le poți opri oricând din aceeași pagină.</p>
  </td></tr>
</table>
</body></html>`))

type digestRow struct {
	Label string
	Value string
}

func (service *ReportDigestService) send(recipient domain.ReportSubscriptionRecipient, now time.Time) error {
	from, to, kind := digestPeriod(recipient.Frequency, now)
	summary, err := service.reports.GetSummary(ReportQuery{From: from, To: to})
	if err != nil {
		return err
	}

	current := summary.Current
	inventory := summary.Inventory
	rows := []digestRow{
		{"Operațiuni în perioadă", fmt.Sprintf("%d (finalizate %d, în lucru %d, planificate %d)", current.OperationsTotal, current.OperationsCompleted, current.OperationsInProgress, current.OperationsPlanned)},
		{"Lucrări întârziate", fmt.Sprintf("%d", current.OverdueOperations)},
		{"Hectare planificate / realizate", fmt.Sprintf("%.1f ha / %.1f ha", current.PlannedAreaHa, current.RealizedAreaHa)},
		{"Cost estimat / real", fmt.Sprintf("%.2f / %.2f", current.EstimatedCost, current.RealCost)},
		{"Combustibil consumat", fmt.Sprintf("%.1f l", current.FuelUsedL)},
		{"Ore de mașină", fmt.Sprintf("%.1f h", current.MachineHours)},
		{"Stocuri sub prag minim", fmt.Sprintf("%d din %d", inventory.LowStocks, inventory.TotalStocks)},
		{"Utilaje în mentenanță", fmt.Sprintf("%d", inventory.MaintenanceAssets)},
	}

	var body bytes.Buffer
	if err := digestTemplate.Execute(&body, map[string]interface{}{
		"Kind": kind,
		"From": formatRomanianDate(from),
		"To":   formatRomanianDate(to),
		"Name": recipient.FirstName,
		"Rows": rows,
		"Link": fmt.Sprintf("%s/reports?from=%s&to=%s", service.clientOrigin, from, to),
	}); err != nil {
		return err
	}

	subject := fmt.Sprintf("AgriERP · Raport %s %s – %s", kind, formatRomanianDate(from), formatRomanianDate(to))
	if err := service.mailer.SendHTML(recipient.Email, subject, body.String()); err != nil {
		return err
	}
	if recipient.ID > 0 {
		return service.repo.MarkSent(recipient.ID, now)
	}
	return nil
}

func formatRomanianDate(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("02.01.2006")
}
