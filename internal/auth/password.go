package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	// Implement password hashing logic here (e.g., using bcrypt)
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	// Implement password hash comparison logic here (e.g., using bcrypt)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
