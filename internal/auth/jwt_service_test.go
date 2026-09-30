package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateAndParse(t *testing.T) {
	svc := NewJWTService("secret", 15*time.Minute, time.Hour)
	if svc.AccessTokenTTL() != 15*time.Minute {
		t.Errorf("AccessTokenTTL: %s", svc.AccessTokenTTL())
	}

	token, err := svc.GenerateAccess("42", 3, "manager", "Manager")
	if err != nil || token == "" {
		t.Fatalf("GenerateAccess: %v", err)
	}
	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if claims["user_id"] != "42" || claims["role_id"] != float64(3) || claims["role_code"] != "manager" || claims["role_name"] != "Manager" {
		t.Errorf("claims greșite: %v", claims)
	}

	jti, err := svc.ExtractJTI(token)
	if err != nil || jti == "" {
		t.Errorf("ExtractJTI: %v, %q", err, jti)
	}
	jti2, ttl, err := svc.ExtractJTIAndTTL(token)
	if err != nil || jti2 != jti || ttl <= 14*time.Minute || ttl > 15*time.Minute {
		t.Errorf("ExtractJTIAndTTL: %v, %q, %s", err, jti2, ttl)
	}

	if _, err := svc.Parse("nu.e.token"); err == nil {
		t.Error("token invalid trebuie să dea eroare")
	}
	if _, err := NewJWTService("alt-secret", time.Minute, time.Hour).Parse(token); err == nil {
		t.Error("semnătura cu alt secret trebuie respinsă")
	}
	expired, _ := NewJWTService("secret", -time.Minute, time.Hour).GenerateAccess("1", 1, "a", "b")
	if _, err := svc.Parse(expired); err == nil {
		t.Error("token expirat trebuie respins")
	}
	if _, err := svc.ExtractJTI("nu.e.token"); err == nil {
		t.Error("ExtractJTI pe token invalid trebuie să dea eroare")
	}
	if _, _, err := svc.ExtractJTIAndTTL("nu.e.token"); err == nil {
		t.Error("ExtractJTIAndTTL pe token invalid trebuie să dea eroare")
	}

	// token semnat fără jti / fără exp
	noJTI, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": "1"}).SignedString([]byte("secret"))
	if _, err := svc.ExtractJTI(noJTI); err == nil {
		t.Error("token fără jti trebuie să dea eroare")
	}
	if _, _, err := svc.ExtractJTIAndTTL(noJTI); err == nil {
		t.Error("ExtractJTIAndTTL fără jti trebuie să dea eroare")
	}
	noExp, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"jti": "x"}).SignedString([]byte("secret"))
	if _, _, err := svc.ExtractJTIAndTTL(noExp); err == nil {
		t.Error("token fără exp trebuie să dea eroare")
	}
}
