package auth

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoToken       = errors.New("authorization token required")
	ErrInvalidToken  = errors.New("invalid or expired token")
	ErrInvalidClaims = errors.New("invalid token claims")
)

// JWTVerifier верифицирует access-токены по публичному ключу Auth Service.
type JWTVerifier struct {
	publicKey ed25519.PublicKey
}

// publicKeyResponse — ответ API Gateway на GET /api/v1/auth/public-key.
type publicKeyResponse struct {
	PublicKeyPEM string `json:"public_key_pem"`
	Algorithm    string `json:"algorithm"`
}

// NewJWTVerifierFromURL загружает публичный ключ по HTTP из API Gateway.
// Реализует retry с exponential backoff — API Gateway может быть ещё не готов.
func NewJWTVerifierFromURL(url string) (*JWTVerifier, error) {
	var lastErr error

	for attempt := 1; attempt <= 10; attempt++ {
		key, err := fetchPublicKey(url)
		if err == nil {
			log.Printf("jwt-verifier: public key loaded from %s (attempt %d)", url, attempt)
			return &JWTVerifier{publicKey: key}, nil
		}

		lastErr = err
		backoff := time.Duration(attempt) * 2 * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		log.Printf("jwt-verifier: attempt %d/10 failed: %v, retrying in %s", attempt, err, backoff)
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("jwt-verifier: failed to load public key after 10 attempts: %w", lastErr)
}

// NewJWTVerifierFromEnv создаёт верификатор из переменной окружения AUTH_PUBLIC_KEY_PEM.
func NewJWTVerifierFromEnv() (*JWTVerifier, error) {
	pemStr := os.Getenv("AUTH_PUBLIC_KEY_PEM")
	if pemStr == "" {
		return nil, errors.New("AUTH_PUBLIC_KEY_PEM not set")
	}

	key, err := parsePEMPublicKey(pemStr)
	if err != nil {
		return nil, fmt.Errorf("parse AUTH_PUBLIC_KEY_PEM: %w", err)
	}

	log.Println("jwt-verifier: public key loaded from environment")
	return &JWTVerifier{publicKey: key}, nil
}

// NewJWTVerifier создаёт верификатор, пробуя сначала env, затем URL.
func NewJWTVerifier() (*JWTVerifier, error) {
	// Сначала пробуем загрузить из переменной окружения.
	if v, err := NewJWTVerifierFromEnv(); err == nil {
		return v, nil
	}

	// Иначе — через HTTP из API Gateway.
	url := os.Getenv("AUTH_PUBLIC_KEY_URL")
	if url == "" {
		url = "http://api-gateway:8080/api/v1/auth/public-key"
	}

	return NewJWTVerifierFromURL(url)
}

// VerifyToken проверяет JWT и возвращает user_id из claims.
func (v *JWTVerifier) VerifyToken(tokenString string) (string, error) {
	if tokenString == "" {
		return "", ErrNoToken
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.publicKey, nil
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

func fetchPublicKey(url string) (ed25519.PublicKey, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var pkResp publicKeyResponse
	if err := json.Unmarshal(body, &pkResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	return parsePEMPublicKey(pkResp.PublicKeyPEM)
}

func parsePEMPublicKey(pemStr string) (ed25519.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	edKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("expected ed25519 public key, got %T", pub)
	}

	return edKey, nil
}
