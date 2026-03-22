package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	RevokeByToken(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID string) error
}

type TokenGenerator interface {
	GenerateAccessToken(userID, email string) (string, error)
	GetPublicKeyPEM() string
	GetAlgorithm() string
}

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, userID, email, name string) error
	PublishUserDeleted(ctx context.Context, userID string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}
