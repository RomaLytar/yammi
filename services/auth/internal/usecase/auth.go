package usecase

import (
	"context"
	"time"

	"github.com/romanlovesweed/yammi/services/auth/internal/domain"
)

type AuthUseCase struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	tokenGenerator   TokenGenerator
	refreshTokenTTL  time.Duration
}

func NewAuthUseCase(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	tokenGenerator TokenGenerator,
	refreshTokenTTL time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenGenerator:   tokenGenerator,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

func (uc *AuthUseCase) Register(ctx context.Context, email, password, name string) (userID, accessToken, refreshToken string, err error) {
	user, err := domain.NewUser(email, password, name)
	if err != nil {
		return "", "", "", err
	}

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

	return user.ID, accessToken, rt.Token, nil
}

func (uc *AuthUseCase) Login(ctx context.Context, email, password string) (userID, accessToken, refreshToken string, err error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", "", "", err
	}

	if err := user.CheckPassword(password); err != nil {
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

	// Revoke old refresh token
	if err := uc.refreshTokenRepo.RevokeByToken(ctx, token); err != nil {
		return "", "", err
	}

	accessToken, err = uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// Issue new refresh token (rotation)
	newRT := domain.NewRefreshToken(user.ID, uc.refreshTokenTTL)
	if err := uc.refreshTokenRepo.Create(ctx, newRT); err != nil {
		return "", "", err
	}

	return accessToken, newRT.Token, nil
}

func (uc *AuthUseCase) RevokeToken(ctx context.Context, token string) error {
	return uc.refreshTokenRepo.RevokeByToken(ctx, token)
}

func (uc *AuthUseCase) GetPublicKey() (pem, algorithm string) {
	return uc.tokenGenerator.GetPublicKeyPEM(), uc.tokenGenerator.GetAlgorithm()
}
