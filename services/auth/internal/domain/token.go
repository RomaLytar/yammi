package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string // хранится как SHA-256 хэш в БД
	ExpiresAt time.Time
	Revoked   bool
	CreatedAt time.Time
}

// NewRefreshToken создаёт новый refresh-токен.
// RawToken содержит оригинальный токен для отправки клиенту.
// Token в структуре содержит SHA-256 хэш для хранения в БД.
func NewRefreshToken(userID string, ttl time.Duration) (rt *RefreshToken, rawToken string) {
	now := time.Now()
	raw := uuid.New().String()
	return &RefreshToken{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     HashToken(raw),
		ExpiresAt: now.Add(ttl),
		Revoked:   false,
		CreatedAt: now,
	}, raw
}

// HashToken возвращает SHA-256 хэш токена для хранения/поиска в БД.
func HashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
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
