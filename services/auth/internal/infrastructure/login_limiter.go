package infrastructure

import (
	"errors"
	"sync"
	"time"
)

var ErrAccountLocked = errors.New("too many failed attempts, account temporarily locked")

type loginAttempt struct {
	count    int
	lockedAt time.Time
}

// LoginLimiter tracks failed login attempts and enforces temporary lockouts.
type LoginLimiter struct {
	mu              sync.Mutex
	attempts        map[string]*loginAttempt
	maxAttempts     int
	lockoutDuration time.Duration
	stop            chan struct{}
}

func NewLoginLimiter(maxAttempts int, lockoutDuration time.Duration) *LoginLimiter {
	l := &LoginLimiter{
		attempts:        make(map[string]*loginAttempt),
		maxAttempts:     maxAttempts,
		lockoutDuration: lockoutDuration,
		stop:            make(chan struct{}),
	}
	go l.cleanup()
	return l
}

// Check returns ErrAccountLocked if the email is currently locked out.
func (l *LoginLimiter) Check(email string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		return nil
	}

	if a.count >= l.maxAttempts && !a.lockedAt.IsZero() {
		if time.Since(a.lockedAt) < l.lockoutDuration {
			return ErrAccountLocked
		}
		// Lockout expired — reset
		delete(l.attempts, email)
	}

	return nil
}

// RecordFailure increments the failure counter. Locks the account when threshold is reached.
func (l *LoginLimiter) RecordFailure(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	a, ok := l.attempts[email]
	if !ok {
		a = &loginAttempt{}
		l.attempts[email] = a
	}

	a.count++
	if a.count >= l.maxAttempts {
		a.lockedAt = time.Now()
	}
}

// Reset clears the failure counter on successful login.
func (l *LoginLimiter) Reset(email string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, email)
}

// Stop stops the background cleanup goroutine.
func (l *LoginLimiter) Stop() {
	close(l.stop)
}

func (l *LoginLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			for email, a := range l.attempts {
				if !a.lockedAt.IsZero() && now.Sub(a.lockedAt) > l.lockoutDuration {
					delete(l.attempts, email)
				}
			}
			l.mu.Unlock()
		case <-l.stop:
			return
		}
	}
}
