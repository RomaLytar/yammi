package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

func (uc *AuthUseCase) RefreshToken(ctx context.Context, token string) (accessToken, newRefreshToken string, err error) {
	rt, err := uc.refreshTokenRepo.GetByToken(ctx, token)
	if err != nil {
		return "", "", err
	}

	if err := rt.IsValid(); err != nil {
		return "", "", err
	}

	user, err := uc.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return "", "", err
	}

	if err := uc.refreshTokenRepo.RevokeByToken(ctx, token); err != nil {
		return "", "", err
	}

	accessToken, err = uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	newRT := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return "", "", err
	}

	return accessToken, newRT.Token, nil
}

func (uc *AuthUseCase) RevokeToken(ctx context.Context, token string) error {
	return uc.refreshTokenRepo.RevokeByToken(ctx, token)
}
