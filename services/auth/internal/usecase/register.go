package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
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

	rt, rawToken := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", "", "", err
	}

	// Публикуем событие (async, non-blocking — не задерживаем ответ клиенту)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.eventPublisher.PublishUserCreated(ctx, user.ID, user.Email, user.Name); err != nil {
			slog.Error("failed to publish UserCreated", "error", err, "user_id", user.ID)
		}
	}()

	return user.ID, accessToken, rawToken, nil
}

func (uc *AuthUseCase) DeleteUser(ctx context.Context, userID string) error {
	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	// Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.eventPublisher.PublishUserDeleted(ctx, userID); err != nil {
			slog.Error("failed to publish UserDeleted", "error", err, "user_id", userID)
		}
	}()

	return nil
}
