package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"agri-api/internal/auth"
)

func TestBlacklist_WithoutRedis(t *testing.T) {
	// Redis inaccesibil: verificăm comportamentul best-effort al blacklist-ului.
	bl := auth.NewBlacklist(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond}))

	if err := bl.Add(context.Background(), "jti", 0); err != nil {
		t.Errorf("un token deja expirat nu se mai adaugă: %v", err)
	}
	if err := bl.Add(context.Background(), "jti", time.Minute); err == nil {
		t.Error("fără Redis, adăugarea trebuie să dea eroare")
	}
	if _, err := bl.IsBlacklisted(context.Background(), "jti"); err == nil {
		t.Error("fără Redis, verificarea trebuie să dea eroare")
	}
}
