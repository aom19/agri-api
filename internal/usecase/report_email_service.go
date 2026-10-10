package usecase

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// ReportMailer este subsetul serviciului de e-mail folosit pentru trimiterea raportului.
type ReportMailer interface {
	SendHTML(to, subject, htmlBody string) error
}

// ReportEmailService trimite sumarul unei perioade pe e-mail, la cererea utilizatorului.
// Nu există abonamente sau trimiteri programate: le adăugăm când apar utilizatori care
// nu intră în aplicație.
type ReportEmailService struct {
	reports      *ReportService
	mailer       ReportMailer
	clientOrigin string
}

func NewReportEmailService(reports *ReportService, mailer ReportMailer, clientOrigin string) *ReportEmailService {
	return &ReportEmailService{reports: reports, mailer: mailer, clientOrigin: clientOrigin}
}

var summaryEmailTemplate = template.Must(template.New("summary").Parse(`
<!DOCTYPE html>
<html lang="ro"><body style="margin:0;padding:24px;background:#f4f6f4;font-family:Arial,Helvetica,sans-serif;color:#0d1f17;">
<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:640px;margin:0 auto;background:#ffffff;border-radius:12px;overflow:hidden;">
  <tr><td style="background:#1a5c38;color:#ffffff;padding:20px 24px;">
    <div style="font-size:12px;opacity:.85;text-transform:uppercase;letter-spacing:.08em;">AgriERP · raport</div>
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
    <p style="margin:16px 0 0;font-size:12px;color:#4a5e54;">Primești acest e-mail pentru că l-ai cerut din pagina Rapoarte.</p>
  </td></tr>
</table>
</body></html>`))

type summaryEmailRow struct {
	Label string
	Value string
}

// SendSummary trimite pe adresa dată sumarul perioadei from–to din query. Filtrele pe teren,
// tip, mașină sau operator sunt ignorate: e-mailul nu le afișează, deci cifrele ar fi înșelătoare.
func (service *ReportEmailService) SendSummary(query ReportQuery, to, firstName string) error {
	summary, err := service.reports.GetSummary(ReportQuery{From: query.From, To: query.To})
	if err != nil {
		return err
	}
	from, until := summary.Period.From, summary.Period.To

	current := summary.Current
	inventory := summary.Inventory
	rows := []summaryEmailRow{
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
	if err := summaryEmailTemplate.Execute(&body, map[string]interface{}{
		"From": formatRomanianDate(from),
		"To":   formatRomanianDate(until),
		"Name": firstName,
		"Rows": rows,
		"Link": fmt.Sprintf("%s/reports?from=%s&to=%s", service.clientOrigin, from, until),
	}); err != nil {
		return err
	}

	subject := fmt.Sprintf("AgriERP · Raport %s – %s", formatRomanianDate(from), formatRomanianDate(until))
	return service.mailer.SendHTML(to, subject, body.String())
}

func formatRomanianDate(value string) string {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return parsed.Format("02.01.2006")
}
