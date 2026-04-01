package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

func (uc *AuthUseCase) RefreshToken(ctx context.Context, token string) (accessToken, newRefreshToken string, err error) {
	// Хэшируем входящий токен для поиска в БД
	tokenHash := domain.HashToken(token)

	// Атомарная операция: ревокаем и получаем токен за один SQL-запрос.
	// Гарантирует, что при конкурентных запросах только один из них успешно ревокнет токен (TOCTOU protection).
	rt, err := uc.refreshTokenRepo.RevokeAndReturn(ctx, tokenHash)
	if err != nil {
		return "", "", err
	}

	user, err := uc.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}

	accessToken, err = uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	newRT, rawToken := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return "", "", err
	}

	return accessToken, rawToken, nil
}

func (uc *AuthUseCase) RevokeToken(ctx context.Context, token string) error {
	return uc.refreshTokenRepo.RevokeByToken(ctx, domain.HashToken(token))
}
