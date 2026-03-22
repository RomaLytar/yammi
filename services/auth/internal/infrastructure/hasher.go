package infrastructure

import (
	"log"
	"runtime"

	"github.com/RomaLytar/yammi/services/auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type BcryptPool struct {
	sem  chan struct{}
	cost int
}

func NewBcryptPool(workers, cost int) *BcryptPool {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	log.Printf("bcrypt pool: %d workers, cost=%d", workers, cost)
	return &BcryptPool{sem: make(chan struct{}, workers), cost: cost}
}

func (p *BcryptPool) Hash(password string) (string, error) {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), p.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (p *BcryptPool) Verify(password, hash string) error {
	p.sem <- struct{}{}
	defer func() { <-p.sem }()

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return domain.ErrInvalidPassword
	}
	return nil
}
