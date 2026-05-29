package domain

import "time"

type User struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

type UserProfile struct {
	UserID       int64      `json:"user_id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	ProfilePhoto string     `json:"profile_photo"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
