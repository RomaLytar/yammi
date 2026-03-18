package usecase

import (
	"context"

	"github.com/romanlovesweed/yammi/services/user/internal/domain"
)

type UserUseCase struct {
	userRepo UserRepository
}

func NewUserUseCase(userRepo UserRepository) *UserUseCase {
	return &UserUseCase{userRepo: userRepo}
}

func (uc *UserUseCase) CreateProfile(ctx context.Context, userID, email, name string) error {
	user := domain.NewUserFromEvent(userID, email, name)
	return uc.userRepo.Create(ctx, user)
}

func (uc *UserUseCase) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, userID)
}

func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID, name, avatarURL, bio string) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := user.Update(name, avatarURL, bio); err != nil {
		return nil, err
	}

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc *UserUseCase) DeleteProfile(ctx context.Context, userID string) error {
	return uc.userRepo.Delete(ctx, userID)
}
