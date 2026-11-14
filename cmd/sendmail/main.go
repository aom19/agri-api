// Comanda sendmail trimite un e-mail de test cu setările SMTP din .env, ca să verifici
// furnizorul (Mailpit local, Mailtrap etc.) fără să treci prin aplicație:
//
//	go run ./cmd/sendmail -to adresa@exemplu.ro
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"agri-api/internal/config"
	"agri-api/internal/email"
	"agri-api/internal/logger"
)

func main() {
	to := flag.String("to", "", "destinatarul e-mailului de test")
	flag.Parse()
	if *to == "" {
		fmt.Fprintln(os.Stderr, "Folosește: go run ./cmd/sendmail -to adresa@exemplu.ro")
		os.Exit(2)
	}

	cfg := config.LoadConfig()
	log := logger.NewLogger(cfg.AppEnv)
	fmt.Printf("Trimit prin %s:%d, de la %s, către %s…\n", cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom, *to)

	svc := email.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPFrom, log)
	body := fmt.Sprintf(`<div style="font-family:Arial,sans-serif;color:#0d1f17">
<h2 style="color:#1a5c38">AgriERP — e-mail de test</h2>
<p>Dacă ai primit acest mesaj, setările SMTP funcționează.</p>
<p style="color:#4a5e54;font-size:13px">Server: %s:%d · trimis la %s</p>
</div>`, cfg.SMTPHost, cfg.SMTPPort, time.Now().Format("02.01.2006 15:04:05"))

	if err := svc.SendHTML(*to, "Test e-mail AgriERP ✔ — diacritice: ăîșțâ", body); err != nil {
		fmt.Fprintf(os.Stderr, "Trimiterea a eșuat: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Trimis.")
}
