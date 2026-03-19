package domain

import (
	"errors"
	"testing"
)

func TestValidateRegistration(t *testing.T) {
	t.Run("valid input passes", func(t *testing.T) {
		err := ValidateRegistration("user@example.com", "securepass", "John Doe")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("empty email fails", func(t *testing.T) {
		err := ValidateRegistration("", "securepass", "John Doe")
		if !errors.Is(err, ErrEmptyEmail) {
			t.Fatalf("expected ErrEmptyEmail, got %v", err)
		}
	})

	t.Run("invalid email format fails", func(t *testing.T) {
		err := ValidateRegistration("not-an-email", "securepass", "John Doe")
		if !errors.Is(err, ErrInvalidEmail) {
			t.Fatalf("expected ErrInvalidEmail, got %v", err)
		}
	})

	t.Run("short password fails", func(t *testing.T) {
		err := ValidateRegistration("user@example.com", "short", "John Doe")
		if !errors.Is(err, ErrWeakPassword) {
			t.Fatalf("expected ErrWeakPassword, got %v", err)
		}
	})

	t.Run("empty password fails", func(t *testing.T) {
		err := ValidateRegistration("user@example.com", "", "John Doe")
		if !errors.Is(err, ErrEmptyPassword) {
			t.Fatalf("expected ErrEmptyPassword, got %v", err)
		}
	})

	t.Run("empty name fails", func(t *testing.T) {
		err := ValidateRegistration("user@example.com", "securepass", "")
		if !errors.Is(err, ErrEmptyName) {
			t.Fatalf("expected ErrEmptyName, got %v", err)
		}
	})
}

func TestNewUserWithHash(t *testing.T) {
	user := NewUserWithHash("user@example.com", "hash123", "John Doe")

	if user.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected email user@example.com, got %s", user.Email)
	}
	if user.Name != "John Doe" {
		t.Fatalf("expected name John Doe, got %s", user.Name)
	}
	if user.PasswordHash != "hash123" {
		t.Fatalf("expected password hash hash123, got %s", user.PasswordHash)
	}
	if user.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if user.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
}
