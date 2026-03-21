package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/user/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	SearchByEmail(ctx context.Context, query string, limit int) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
