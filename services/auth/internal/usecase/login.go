package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
)

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (userID, accessToken, refreshToken string, err error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", "", err
	}

	if err := uc.hasher.Verify(password, user.PasswordHash); err != nil {
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

	return user.ID, accessToken, rt.Token, nil
}
