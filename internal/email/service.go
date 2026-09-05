package email

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

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

	headers := []string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
	}

	message := strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.TrimSpace(htmlBody)

	var auth smtp.Auth
	if s.user != "" && s.password != "" {
		auth = smtp.PlainAuth("", s.user, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message)); err != nil {
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

// SendHTML trimite un e-mail HTML arbitrar (folosit de rapoartele programate).
func (s *EmailService) SendHTML(to, subject, htmlBody string) error {
	return s.sendHTML(to, subject, htmlBody)
}
