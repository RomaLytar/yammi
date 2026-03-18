package usecase

import (
	"time"
)

type AuthUseCase struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	tokenGenerator   TokenGenerator
	eventPublisher   EventPublisher
	hasher           PasswordHasher
	refreshTokenTTL  time.Duration
}

func NewAuthUseCase(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	tokenGenerator TokenGenerator,
	eventPublisher EventPublisher,
	hasher PasswordHasher,
	refreshTokenTTL time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenGenerator:   tokenGenerator,
		eventPublisher:   eventPublisher,
		hasher:           hasher,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

func (uc *AuthUseCase) GetPublicKey() (pem, algorithm string) {
	return uc.tokenGenerator.GetPublicKeyPEM(), uc.tokenGenerator.GetAlgorithm()
}
