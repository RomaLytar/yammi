package infrastructure

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authpb "github.com/RomaLytar/yammi/services/api-gateway/api/proto/v1"
)

var (
	ErrNoToken       = errors.New("authorization token required")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// JWTVerifier верифицирует access-токены по публичному ключу Auth Service.
// Публичный ключ кэшируется, повторный fetch не чаще чем раз в refetchCooldown.
type JWTVerifier struct {
	authClient      authpb.AuthServiceClient
	publicKey       ed25519.PublicKey
	mu              sync.RWMutex
	lastFetchAt     time.Time
	refetchCooldown time.Duration // минимальный интервал между refetch (защита от DoS)
}

func NewJWTVerifier(authClient authpb.AuthServiceClient) *JWTVerifier {
	v := &JWTVerifier{
		authClient:      authClient,
		refetchCooldown: 30 * time.Second,
	}

	// Пробуем загрузить ключ при старте (3 попытки)
	for i := 0; i < 3; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err := v.fetchPublicKey(ctx)
		cancel()
		if err == nil {
			log.Println("jwt-verifier: public key loaded")
			return v
		}
		log.Printf("jwt-verifier: fetch key attempt %d/3 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	log.Println("jwt-verifier: WARNING — started without public key, will retry on first request")
	return v
}

func (v *JWTVerifier) fetchPublicKey(ctx context.Context) error {
	resp, err := v.authClient.GetPublicKey(ctx, &authpb.GetPublicKeyRequest{})
	if err != nil {
		return fmt.Errorf("grpc GetPublicKey: %w", err)
	}

	block, _ := pem.Decode([]byte(resp.PublicKeyPem))
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}

	edKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("expected ed25519 public key, got %T", pub)
	}

	v.mu.Lock()
	v.publicKey = edKey
	v.lastFetchAt = time.Now()
	v.mu.Unlock()

	return nil
}

// VerifyToken проверяет JWT и возвращает user_id из claims.
func (v *JWTVerifier) VerifyToken(tokenString string) (string, error) {
	userID, err := v.verifyWithCurrentKey(tokenString)
	if err == nil {
		return userID, nil
	}

	// Ключ мог смениться (рестарт auth) — пробуем перезагрузить.
	// Cooldown защищает от DoS: не чаще чем раз в refetchCooldown.
	v.mu.RLock()
	canRefetch := time.Since(v.lastFetchAt) > v.refetchCooldown
	v.mu.RUnlock()

	if !canRefetch {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if fetchErr := v.fetchPublicKey(ctx); fetchErr != nil {
		return "", err
	}

	return v.verifyWithCurrentKey(tokenString)
}

func (v *JWTVerifier) verifyWithCurrentKey(tokenString string) (string, error) {
	v.mu.RLock()
	key := v.publicKey
	v.mu.RUnlock()

	if key == nil {
		return "", ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", ErrInvalidClaims
	}

	iss, _ := claims["iss"].(string)
	if iss != "yammi-auth" {
		return "", ErrInvalidClaims
	}

	return sub, nil
}
