package email

import (
	"strings"
	"testing"

	"agri-api/internal/logger"
)

// SMTP-ul nu e disponibil în teste: verificăm că șabloanele se randează și că
// trimiterea raportează eroarea de conexiune.
func TestEmailService_SendFailsWithoutSMTP(t *testing.T) {
	svc := NewEmailService("127.0.0.1", 1, "", "", "noreply@test.local", nil)

	if err := svc.SendAccountConfirmation("a@x.ro", "http://front/confirm?token=1"); err == nil {
		t.Error("fără SMTP, confirmarea trebuie să dea eroare")
	}
	if err := svc.SendPasswordReset("a@x.ro", "http://front/reset/1"); err == nil {
		t.Error("fără SMTP, resetarea trebuie să dea eroare")
	}
	if err := svc.SendPasswordChanged("a@x.ro"); err == nil {
		t.Error("fără SMTP, notificarea trebuie să dea eroare")
	}

	// cu autentificare și logger: aceleași căi, plus logarea erorii
	withAuth := NewEmailService("127.0.0.1", 1, "user", "pass", "noreply@test.local", logger.NewLogger("development"))
	err := withAuth.SendHTML("a@x.ro", "Subiect", "<p>Salut</p>")
	if err == nil || !strings.Contains(err.Error(), "connect") {
		t.Errorf("mă așteptam la eroare de conexiune: %v", err)
	}
}
