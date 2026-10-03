package auth_test

import (
	"testing"

	"agri-api/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	hash, err := auth.HashPassword("Parola1!")
	if err != nil || hash == "" || hash == "Parola1!" {
		t.Fatalf("HashPassword: %v, %q", err, hash)
	}
	if !auth.CheckPasswordHash("Parola1!", hash) {
		t.Error("parola corectă trebuie acceptată")
	}
	if auth.CheckPasswordHash("gresit", hash) {
		t.Error("parola greșită trebuie respinsă")
	}
	if auth.CheckPasswordHash("Parola1!", "nu-e-hash") {
		t.Error("hash-ul invalid trebuie respins")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	a, b := auth.GenerateRefreshToken(), auth.GenerateRefreshToken()
	if a == "" || a == b {
		t.Error("token-urile trebuie să fie nevide și unice")
	}
}
