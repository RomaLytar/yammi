package usecase

import (
	"context"
	"log"
	"time"

	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type SettingsUseCase struct {
	repo      SettingsRepository
	publisher SettingsEventPublisher
}

func NewSettingsUseCase(repo SettingsRepository, publisher SettingsEventPublisher) *SettingsUseCase {
	return &SettingsUseCase{repo: repo, publisher: publisher}
}

func (uc *SettingsUseCase) Get(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	if userID == "" {
		return nil, domain.ErrEmptyUserID
	}
	return uc.repo.Get(ctx, userID)
}

func (uc *SettingsUseCase) Update(ctx context.Context, userID string, enabled, realtimeEnabled bool) (*domain.NotificationSettings, error) {
	if userID == "" {
		return nil, domain.ErrEmptyUserID
	}

	settings := &domain.NotificationSettings{
		UserID:          userID,
		Enabled:         enabled,
		RealtimeEnabled: realtimeEnabled,
		UpdatedAt:       time.Now(),
	}

	if err := uc.repo.Upsert(ctx, settings); err != nil {
		return nil, err
	}

	// Публикуем событие для инвалидации кеша на других инстансах
	if uc.publisher != nil {
		go func() {
			if err := uc.publisher.PublishSettingsUpdated(context.Background(), userID, enabled, realtimeEnabled); err != nil {
				log.Printf("failed to publish settings.updated for user %s: %v", userID, err)
			}
		}()
	}

	return uc.repo.Get(ctx, userID)
}
