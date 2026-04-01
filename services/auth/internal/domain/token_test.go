package domain

import (
	"errors"
	"testing"
	"time"
)

func TestRefreshToken_IsValid(t *testing.T) {
	t.Run("non-expired non-revoked token is valid", func(t *testing.T) {
		rt := &RefreshToken{
			ID:        "id-1",
			UserID:    "user-1",
			Token:     "token-1",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			Revoked:   false,
			CreatedAt: time.Now(),
		}
		if err := rt.IsValid(); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("expired token returns ErrTokenExpired", func(t *testing.T) {
		rt := &RefreshToken{
			ID:        "id-2",
			UserID:    "user-2",
			Token:     "token-2",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Revoked:   false,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		err := rt.IsValid()
		if !errors.Is(err, ErrTokenExpired) {
			t.Fatalf("expected ErrTokenExpired, got %v", err)
		}
	})

	t.Run("revoked token returns ErrTokenRevoked", func(t *testing.T) {
		rt := &RefreshToken{
			ID:        "id-3",
			UserID:    "user-3",
			Token:     "token-3",
			ExpiresAt: time.Now().Add(1 * time.Hour),
			Revoked:   true,
			CreatedAt: time.Now(),
		}
		err := rt.IsValid()
		if !errors.Is(err, ErrTokenRevoked) {
			t.Fatalf("expected ErrTokenRevoked, got %v", err)
		}
	})

	t.Run("revoked takes priority over expired", func(t *testing.T) {
		rt := &RefreshToken{
			ID:        "id-4",
			UserID:    "user-4",
			Token:     "token-4",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
			Revoked:   true,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		err := rt.IsValid()
		if !errors.Is(err, ErrTokenRevoked) {
			t.Fatalf("expected ErrTokenRevoked (checked first), got %v", err)
		}
	})
}

func TestRefreshToken_Revoke(t *testing.T) {
	rt := &RefreshToken{
		ID:        "id-5",
		UserID:    "user-5",
		Token:     "token-5",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Revoked:   false,
		CreatedAt: time.Now(),
	}

	if rt.Revoked {
		t.Fatal("expected token to not be revoked initially")
	}

	rt.Revoke()

	if !rt.Revoked {
		t.Fatal("expected token to be revoked after Revoke()")
	}

	err := rt.IsValid()
	if !errors.Is(err, ErrTokenRevoked) {
		t.Fatalf("expected ErrTokenRevoked after Revoke(), got %v", err)
	}
}

func TestNewRefreshToken(t *testing.T) {
	ttl := 24 * time.Hour
	rt, rawToken := NewRefreshToken("user-1", ttl)

	if rt.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if rt.UserID != "user-1" {
		t.Fatalf("expected UserID user-1, got %s", rt.UserID)
	}
	if rt.Token == "" {
		t.Fatal("expected non-empty Token (hash)")
	}
	if rawToken == "" {
		t.Fatal("expected non-empty rawToken")
	}
	// Token в структуре должен быть хэшем, а не сырым значением
	if rt.Token == rawToken {
		t.Fatal("expected Token to be hashed, not raw")
	}
	// Повторное хэширование raw должно давать тот же Token
	if HashToken(rawToken) != rt.Token {
		t.Fatal("expected HashToken(rawToken) to equal rt.Token")
	}
	if rt.Revoked {
		t.Fatal("expected new token to not be revoked")
	}
	if rt.ExpiresAt.Before(time.Now()) {
		t.Fatal("expected ExpiresAt to be in the future")
	}
}
