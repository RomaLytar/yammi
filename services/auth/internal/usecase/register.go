package usecase

import (
	"context"
	"log"

	"github.com/romanlovesweed/yammi/services/auth/internal/domain"
)

func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string) (userID, accessToken, refreshToken string, err error) {
	if err := domain.ValidateRegistration(email, password, name); err != nil {
		return "", "", "", err
	}

	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return "", "", "", err
	}

	user := domain.NewUserWithHash(email, hash, name)

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return "", "", "", err
	}

	accessToken, err = uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", "", err
	}

	rt := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", "", "", err
	}

	if err := uc.eventPublisher.PublishUserCreated(ctx, user.ID, user.Email, user.Name); err != nil {
		log.Printf("WARNING: failed to publish UserCreated event for user %s: %v", user.ID, err)
	}

	return user.ID, accessToken, rt.Token, nil
}

func (uc *AuthUseCase) DeleteUser(ctx context.Context, userID string) error {
	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	if err := uc.eventPublisher.PublishUserDeleted(ctx, userID); err != nil {
		log.Printf("WARNING: failed to publish UserDeleted event for user %s: %v", userID, err)
	}

	return nil
}
