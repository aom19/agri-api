package email_test

import (
	"bufio"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strconv"
	"strings"
	"testing"

	"agri-api/internal/email"
	"agri-api/internal/logger"
)

// SMTP-ul nu e disponibil în teste: verificăm că șabloanele se randează și că
// trimiterea raportează eroarea de conexiune.
func TestEmailService_SendFailsWithoutSMTP(t *testing.T) {
	svc := email.NewEmailService("127.0.0.1", 1, "", "", "noreply@test.local", nil)

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
	withAuth := email.NewEmailService("127.0.0.1", 1, "user", "pass", "noreply@test.local", logger.NewLogger("development"))
	err := withAuth.SendHTML("a@x.ro", "Subiect", "<p>Salut</p>")
	if err == nil || !strings.Contains(err.Error(), "connect") {
		t.Errorf("mă așteptam la eroare de conexiune: %v", err)
	}
}

// fakeSMTP acceptă o singură conexiune fără TLS și reține plicul și mesajul primit.
type fakeSMTP struct {
	addr     string
	mailFrom string
	rcptTo   string
	data     string
	done     chan struct{}
}

func startFakeSMTP(t *testing.T) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	srv := &fakeSMTP{addr: ln.Addr().String(), done: make(chan struct{})}
	go func() {
		defer close(srv.done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		r := bufio.NewReader(conn)
		reply := func(line string) { fmt.Fprintf(conn, "%s\r\n", line) }
		reply("220 fake ESMTP")
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			cmd := strings.TrimSpace(line)
			switch upper := strings.ToUpper(cmd); {
			case strings.HasPrefix(upper, "EHLO"), strings.HasPrefix(upper, "HELO"):
				reply("250 fake")
			case strings.HasPrefix(upper, "MAIL FROM:"):
				srv.mailFrom = cmd[len("MAIL FROM:"):]
				reply("250 ok")
			case strings.HasPrefix(upper, "RCPT TO:"):
				srv.rcptTo = cmd[len("RCPT TO:"):]
				reply("250 ok")
			case upper == "DATA":
				reply("354 go ahead")
				var data strings.Builder
				for {
					l, err := r.ReadString('\n')
					if err != nil || l == ".\r\n" {
						break
					}
					data.WriteString(l)
				}
				srv.data = data.String()
				reply("250 queued")
			case upper == "QUIT":
				reply("221 bye")
				return
			default:
				reply("250 ok")
			}
		}
	}()
	return srv
}

// Mesajul trimis: expeditor cu nume, subiect cu diacritice codat, Date, Message-ID și corp HTML
// quoted-printable care se decodează înapoi exact.
func TestEmailService_MessageFormat(t *testing.T) {
	srv := startFakeSMTP(t)
	host, portRaw, _ := net.SplitHostPort(srv.addr)
	port, _ := strconv.Atoi(portRaw)
	svc := email.NewEmailService(host, port, "", "", "AgriERP <noreply@agri.ro>", nil)

	html := "<p>Recolta de grâu a fost înregistrată.</p>" + strings.Repeat("x", 1200)
	if err := svc.SendHTML("ana@exemplu.ro", "Confirmă-ți contul", html); err != nil {
		t.Fatal(err)
	}
	<-srv.done

	if srv.mailFrom != "<noreply@agri.ro>" || srv.rcptTo != "<ana@exemplu.ro>" {
		t.Errorf("plicul folosește adresele simple: from=%s to=%s", srv.mailFrom, srv.rcptTo)
	}
	msg, err := mail.ReadMessage(strings.NewReader(srv.data))
	if err != nil {
		t.Fatal(err)
	}
	subject, _ := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject"))
	from, _ := mail.ParseAddress(msg.Header.Get("From"))
	if subject != "Confirmă-ți contul" || from == nil || from.Name != "AgriERP" {
		t.Errorf("antet: subject=%q from=%v", subject, from)
	}
	if msg.Header.Get("Date") == "" || !strings.HasSuffix(msg.Header.Get("Message-ID"), "@agri.ro>") {
		t.Errorf("Date și Message-ID: %v", msg.Header)
	}
	body, _ := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if got := strings.TrimRight(string(body), "\r\n"); got != html {
		t.Errorf("corpul decodat diferă de cel trimis: %q…", got[:min(len(got), 80)])
	}
	for _, line := range strings.Split(srv.data, "\r\n") {
		if len(line) > 998 {
			t.Fatalf("linie de %d caractere: SMTP permite cel mult 998", len(line))
		}
	}
}

func TestEmailService_InvalidFrom(t *testing.T) {
	svc := email.NewEmailService("127.0.0.1", 1, "", "", "nu e adresă", nil)
	if err := svc.SendHTML("a@x.ro", "S", "<p>x</p>"); err == nil || !strings.Contains(err.Error(), "SMTP_FROM") {
		t.Errorf("expeditorul invalid trebuie semnalat: %v", err)
	}
}
