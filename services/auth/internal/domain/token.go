package domain

import (
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

func NewRefreshToken(userID string, ttl time.Duration) *RefreshToken {
	now := time.Now()
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
		CreatedAt: now,
	}
}

func (rt *RefreshToken) IsValid() error {
	if rt.Revoked {
		return ErrTokenRevoked
	}
	if time.Now().After(rt.ExpiresAt) {
		return ErrTokenExpired
	}
	return nil
}

func (rt *RefreshToken) Revoke() {
	rt.Revoked = true
}
