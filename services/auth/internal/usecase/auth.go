package usecase

import (
	"time"
)

// LoginLimiter tracks failed login attempts and enforces temporary lockouts.
type LoginLimiter interface {
	Check(email string) error
	RecordFailure(email string)
	Reset(email string)
}

type AuthUseCase struct {
	userRepo         UserRepository
	refreshTokenRepo RefreshTokenRepository
	tokenGenerator   TokenGenerator
	eventPublisher   EventPublisher
	hasher           PasswordHasher
	loginLimiter     LoginLimiter
	refreshTokenTTL  time.Duration
}

func NewAuthUseCase(
	userRepo UserRepository,
	refreshTokenRepo RefreshTokenRepository,
	tokenGenerator TokenGenerator,
	eventPublisher EventPublisher,
	hasher PasswordHasher,
	loginLimiter LoginLimiter,
	refreshTokenTTL time.Duration,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		tokenGenerator:   tokenGenerator,
		eventPublisher:   eventPublisher,
		hasher:           hasher,
		loginLimiter:     loginLimiter,
		refreshTokenTTL:  refreshTokenTTL,
	}
}

func (uc *AuthUseCase) GetPublicKey() (pem, algorithm string) {
	return uc.tokenGenerator.GetPublicKeyPEM(), uc.tokenGenerator.GetAlgorithm()
}
