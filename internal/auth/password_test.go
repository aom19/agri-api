package auth

import "testing"

func TestPasswordHashing(t *testing.T) {
	hash, err := HashPassword("Parola1!")
	if err != nil || hash == "" || hash == "Parola1!" {
		t.Fatalf("HashPassword: %v, %q", err, hash)
	}
	if !CheckPasswordHash("Parola1!", hash) {
		t.Error("parola corectă trebuie acceptată")
	}
	if CheckPasswordHash("gresit", hash) {
		t.Error("parola greșită trebuie respinsă")
	}
	if CheckPasswordHash("Parola1!", "nu-e-hash") {
		t.Error("hash-ul invalid trebuie respins")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	a, b := GenerateRefreshToken(), GenerateRefreshToken()
	if a == "" || a == b {
		t.Error("token-urile trebuie să fie nevide și unice")
	}
}
