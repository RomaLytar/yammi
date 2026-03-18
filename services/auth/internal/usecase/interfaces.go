package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/auth/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
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
