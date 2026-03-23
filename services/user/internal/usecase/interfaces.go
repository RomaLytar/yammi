package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/user/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByIDs(ctx context.Context, ids []string) ([]*domain.User, error)
	SearchByEmail(ctx context.Context, query string, limit int) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
