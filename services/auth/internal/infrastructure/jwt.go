package infrastructure

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTGenerator struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	accessTTL  time.Duration
}

func NewJWTGenerator(privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, issuer string, accessTTL time.Duration) *JWTGenerator {
	return &JWTGenerator{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		accessTTL:  accessTTL,
	}
}

func (g *JWTGenerator) GenerateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iss":   g.issuer,
		"iat":   now.Unix(),
		"exp":   now.Add(g.accessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return token.SignedString(g.privateKey)
}

func (g *JWTGenerator) GetPublicKeyPEM() string {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(g.publicKey)
	if err != nil {
		return ""
	}

	block := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	return string(pem.EncodeToMemory(block))
}

func (g *JWTGenerator) GetAlgorithm() string {
	return "EdDSA"
}

func GenerateKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("generate ed25519 key pair: %w", err)
	}
	return priv, pub, nil
}

// KeyPairFromSeed восстанавливает детерминированную пару ключей из base64-encoded seed (32 bytes).
// Все реплики с одинаковым seed получат одинаковые ключи.
func KeyPairFromSeed(seedB64 string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	seed, err := base64.StdEncoding.DecodeString(seedB64)
	if err != nil {
		return nil, nil, fmt.Errorf("decode JWT_SEED: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, nil, fmt.Errorf("JWT_SEED must be %d bytes, got %d", ed25519.SeedSize, len(seed))
	}
	priv := ed25519.NewKeyFromSeed(seed)
	pub := priv.Public().(ed25519.PublicKey)
	return priv, pub, nil
}
