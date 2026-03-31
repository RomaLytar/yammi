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
	if !strings.Contains(email, "@") || len(email) < 3 ||
		strings.HasPrefix(email, "@") || strings.HasSuffix(email, "@") {
		return ErrInvalidEmail
	}
	if password == "" {
		return ErrEmptyPassword
	}
	if len(password) < 8 {
		return ErrWeakPassword
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range password {
		switch {
		case 'A' <= c && c <= 'Z':
			hasUpper = true
		case 'a' <= c && c <= 'z':
			hasLower = true
		case '0' <= c && c <= '9':
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return ErrWeakPassword
	}
	if len(password) > 72 {
		return ErrPasswordTooLong
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
