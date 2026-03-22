package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
)

func TestSettings_Get_Success(t *testing.T) {
	settings := &mockSettingsRepo{
		getFn: func(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
			return domain.DefaultSettings(userID), nil
		},
	}

	uc := NewSettingsUseCase(settings)
	s, err := uc.Get(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.UserID != "user-1" {
		t.Errorf("expected UserID=user-1, got %s", s.UserID)
	}
	if !s.Enabled {
		t.Error("expected Enabled=true")
	}
}

func TestSettings_Get_EmptyUserID(t *testing.T) {
	uc := NewSettingsUseCase(&mockSettingsRepo{})
	_, err := uc.Get(context.Background(), "")
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestSettings_Update_Success(t *testing.T) {
	var upserted bool
	settings := &mockSettingsRepo{
		upsertFn: func(ctx context.Context, s *domain.NotificationSettings) error {
			upserted = true
			if s.Enabled {
				t.Error("expected Enabled=false")
			}
			return nil
		},
		getFn: func(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
			s := domain.DefaultSettings(userID)
			s.Enabled = false
			return s, nil
		},
	}

	uc := NewSettingsUseCase(settings)
	s, err := uc.Update(context.Background(), "user-1", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !upserted {
		t.Error("expected Upsert to be called")
	}
	if s.Enabled {
		t.Error("expected Enabled=false after update")
	}
}

func TestSettings_Update_EmptyUserID(t *testing.T) {
	uc := NewSettingsUseCase(&mockSettingsRepo{})
	_, err := uc.Update(context.Background(), "", true, true)
	if !errors.Is(err, domain.ErrEmptyUserID) {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestSettings_Update_RepoError(t *testing.T) {
	settings := &mockSettingsRepo{
		upsertFn: func(ctx context.Context, s *domain.NotificationSettings) error {
			return errors.New("db error")
		},
	}

	uc := NewSettingsUseCase(settings)
	_, err := uc.Update(context.Background(), "user-1", true, true)
	if err == nil {
		t.Fatal("expected error")
	}
}
