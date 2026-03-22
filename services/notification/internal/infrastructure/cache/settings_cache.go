package cache

import (
	"context"
	"sync"

	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
	"github.com/romanlovesweed/yammi/services/notification/internal/usecase"
)

// SettingsCache — in-memory кеш настроек уведомлений.
// Реализует usecase.SettingsRepository (decorator pattern).
type SettingsCache struct {
	delegate usecase.SettingsRepository
	mu       sync.RWMutex
	cache    map[string]*domain.NotificationSettings
}

func NewSettingsCache(delegate usecase.SettingsRepository) *SettingsCache {
	return &SettingsCache{
		delegate: delegate,
		cache:    make(map[string]*domain.NotificationSettings),
	}
}

func (c *SettingsCache) Get(ctx context.Context, userID string) (*domain.NotificationSettings, error) {
	c.mu.RLock()
	if s, ok := c.cache[userID]; ok {
		c.mu.RUnlock()
		return s, nil
	}
	c.mu.RUnlock()

	s, err := c.delegate.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[userID] = s
	c.mu.Unlock()

	return s, nil
}

func (c *SettingsCache) BatchGet(ctx context.Context, userIDs []string) (map[string]*domain.NotificationSettings, error) {
	result := make(map[string]*domain.NotificationSettings, len(userIDs))
	var missing []string

	c.mu.RLock()
	for _, uid := range userIDs {
		if s, ok := c.cache[uid]; ok {
			result[uid] = s
		} else {
			missing = append(missing, uid)
		}
	}
	c.mu.RUnlock()

	if len(missing) == 0 {
		return result, nil
	}

	fetched, err := c.delegate.BatchGet(ctx, missing)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	for uid, s := range fetched {
		c.cache[uid] = s
		result[uid] = s
	}
	c.mu.Unlock()

	return result, nil
}

func (c *SettingsCache) Upsert(ctx context.Context, settings *domain.NotificationSettings) error {
	if err := c.delegate.Upsert(ctx, settings); err != nil {
		return err
	}

	c.mu.Lock()
	c.cache[settings.UserID] = settings
	c.mu.Unlock()

	return nil
}

// Invalidate удаляет настройки из кеша (вызывается при NATS event).
func (c *SettingsCache) Invalidate(userID string) {
	c.mu.Lock()
	delete(c.cache, userID)
	c.mu.Unlock()
}
