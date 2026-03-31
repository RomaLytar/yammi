package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (userID, accessToken, refreshToken string, err error) {
	// Проверяем блокировку по brute-force
	if err := uc.loginLimiter.Check(email); err != nil {
		return "", "", "", err
	}

	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			// Константное время: выполняем bcrypt даже если пользователь не найден,
			// чтобы атакующий не мог определить существование email по времени ответа.
			_ = uc.hasher.Verify(password, "$2a$10$invalidhashpaddingtoresisttimingattacksx")
			uc.loginLimiter.RecordFailure(email)
			slog.Warn("login attempt for non-existent user", "email", email)
			return "", "", "", domain.ErrInvalidPassword
		}
		return "", "", "", err
	}

	if err := uc.hasher.Verify(password, user.PasswordHash); err != nil {
		uc.loginLimiter.RecordFailure(email)
		slog.Warn("failed login attempt", "user_id", user.ID, "email", email)
		return "", "", "", err
	}

	uc.loginLimiter.Reset(email)
	slog.Info("successful login", "user_id", user.ID)

	accessToken, err = uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", "", err
	}

	rt := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", "", "", err
	}

	return user.ID, accessToken, rt.Token, nil
}
