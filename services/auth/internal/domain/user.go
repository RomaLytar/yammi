package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func ValidateRegistration(email, password, name string) error {
	if email == "" {
		return ErrEmptyEmail
	}
	if !strings.Contains(email, "@") {
		return ErrInvalidEmail
	}
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	if name == "" {
		return ErrEmptyName
	}
	return nil
}

func NewUserWithHash(email, passwordHash, name string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
