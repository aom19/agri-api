package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTService(secret string, accessTokenTTL, refreshTokenTTL time.Duration) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (service *JWTService) AccessTokenTTL() time.Duration {
	return service.accessTokenTTL
}

func (service *JWTService) GenerateAccess(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"jti":     uuid.NewString(),
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(service.accessTokenTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(service.secret)
}

func (service *JWTService) Parse(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return service.secret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token.Claims.(jwt.MapClaims), nil
}

// ExtractJTI extrage jti-ul dintr-un token fără a valida expiry-ul
func (service *JWTService) ExtractJTI(tokenString string) (string, error) {
	p := jwt.NewParser()
	token, _, err := p.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", errors.New("jti not found in token")
	}
	return jti, nil
}

// ExtractJTIAndTTL extrage jti și durata rămasă până la expirare
func (service *JWTService) ExtractJTIAndTTL(tokenString string) (string, time.Duration, error) {
	p := jwt.NewParser()
	token, _, err := p.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return "", 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", 0, errors.New("invalid claims")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", 0, errors.New("jti not found in token")
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return "", 0, errors.New("exp not found in token")
	}
	ttl := time.Until(time.Unix(int64(exp), 0))
	return jti, ttl, nil
}
