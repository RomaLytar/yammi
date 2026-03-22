package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

func TestGetUnreadCount_Success(t *testing.T) {
	unreadCounter := &mockUnreadCounter{
		getFn: func(ctx context.Context, userID string) (int, error) {
			return 5, nil
		},
	}

	uc := NewGetUnreadCountUseCase(unreadCounter)
	count, err := uc.Execute(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Errorf("expected count=5, got %d", count)
	}
}

func TestGetUnreadCount_EmptyUserID(t *testing.T) {
	uc := NewGetUnreadCountUseCase(&mockUnreadCounter{})
	_, err := uc.Execute(context.Background(), "")
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestGetUnreadCount_RedisError(t *testing.T) {
	unreadCounter := &mockUnreadCounter{
		getFn: func(ctx context.Context, userID string) (int, error) {
			return 0, errors.New("redis error")
		},
	}

	uc := NewGetUnreadCountUseCase(unreadCounter)
	_, err := uc.Execute(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}
