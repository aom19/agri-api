package email

import (
	"bytes"
	"crypto/rand"
	_ "embed"
	"encoding/hex"
	"fmt"
	"html/template"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"agri-api/internal/logger"
)

//go:embed templates/confirmation.html
var confirmationTpl string

//go:embed templates/password_reset.html
var passwordResetTpl string

//go:embed templates/password_changed.html
var passwordChangedTpl string

type EmailService struct {
	host     string
	port     int
	user     string
	password string
	from     string
	log      *logger.Logger

	confirmTpl         *template.Template
	resetTpl           *template.Template
	passwordChangedTpl *template.Template
}

func NewEmailService(host string, port int, user, password, from string, log *logger.Logger) *EmailService {
	return &EmailService{
		host:               host,
		port:               port,
		user:               user,
		password:           password,
		from:               from,
		log:                log,
		confirmTpl:         template.Must(template.New("confirmation").Parse(confirmationTpl)),
		resetTpl:           template.Must(template.New("password_reset").Parse(passwordResetTpl)),
		passwordChangedTpl: template.Must(template.New("password_changed").Parse(passwordChangedTpl)),
	}
}

func (s *EmailService) SendAccountConfirmation(to, confirmURL string) error {
	var buf bytes.Buffer
	if err := s.confirmTpl.Execute(&buf, struct{ ConfirmURL string }{ConfirmURL: confirmURL}); err != nil {
		return fmt.Errorf("confirmation template: %w", err)
	}
	return s.sendHTML(to, "Confirmă-ți contul AgriERP", buf.String())
}

func (s *EmailService) SendPasswordReset(to, resetURL string) error {
	var buf bytes.Buffer
	if err := s.resetTpl.Execute(&buf, struct{ ResetURL string }{ResetURL: resetURL}); err != nil {
		return fmt.Errorf("password_reset template: %w", err)
	}
	return s.sendHTML(to, "Resetare parolă AgriERP", buf.String())
}

func (s *EmailService) SendPasswordChanged(to string) error {
	var buf bytes.Buffer
	if err := s.passwordChangedTpl.Execute(&buf, nil); err != nil {
		return fmt.Errorf("password_changed template: %w", err)
	}
	return s.sendHTML(to, "Parolă modificată cu succes — AgriERP", buf.String())
}

func (s *EmailService) sendHTML(to, subject, htmlBody string) error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	from, err := mail.ParseAddress(s.from)
	if err != nil {
		return fmt.Errorf("SMTP_FROM invalid (%q): %w", s.from, err)
	}
	message, err := buildMessage(from, to, subject, htmlBody, time.Now())
	if err != nil {
		return err
	}

	// SendMail trece pe STARTTLS când serverul îl oferă (Mailtrap, Gmail etc.); fără TLS,
	// PlainAuth refuză să trimită parola, cu excepția serverelor locale (Mailpit).
	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, from.Address, []string{to}, message); err != nil {
		if s.log != nil {
			s.log.Errorf("[email] send failed to=%s subject=%s: %v", to, subject, err)
		}
		return err
	}

	if s.log != nil {
		s.log.Infof("[email] sent to=%s subject=%s", to, subject)
	}
	return nil
}

// buildMessage compune mesajul MIME: subiectul și numele expeditorului codate pentru diacritice,
// Date și Message-ID (fără ele, Gmail trimite mai ușor în spam) și corpul HTML quoted-printable,
// ca liniile să rămână sub limita SMTP de 998 de caractere.
func buildMessage(from *mail.Address, to, subject, htmlBody string, now time.Time) ([]byte, error) {
	var body bytes.Buffer
	qp := quotedprintable.NewWriter(&body)
	if _, err := qp.Write([]byte(strings.TrimSpace(htmlBody))); err != nil {
		return nil, err
	}
	if err := qp.Close(); err != nil {
		return nil, err
	}

	domain := "localhost"
	if at := strings.LastIndex(from.Address, "@"); at >= 0 {
		domain = from.Address[at+1:]
	}
	headers := []string{
		"From: " + from.String(),
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"Date: " + now.Format(time.RFC1123Z),
		fmt.Sprintf("Message-ID: <%d.%s@%s>", now.UnixNano(), randomToken(), domain),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"Content-Transfer-Encoding: quoted-printable",
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + body.String()), nil
}

func randomToken() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "0"
	}
	return hex.EncodeToString(buf)
}

// SendHTML trimite un e-mail HTML arbitrar (folosit de rapoartele programate).
func (s *EmailService) SendHTML(to, subject, htmlBody string) error {
	return s.sendHTML(to, subject, htmlBody)
}
