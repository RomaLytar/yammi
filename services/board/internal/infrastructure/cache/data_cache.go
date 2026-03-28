package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DataCache — lazy Redis кеш для labels, settings, user labels.
// Паттерн: read → Redis hit? return : PostgreSQL fallback → SET в Redis.
// Инвалидация: DEL key при create/update/delete.
// TTL 10 минут как safety net (основная инвалидация — явный DEL при записи).
type DataCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewDataCache(redisURL string) (*DataCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &DataCache{client: client, ttl: 10 * time.Minute}, nil
}

// ── Key helpers ────────────────────────────────────────────────────────

func boardLabelsKey(boardID string) string    { return "board_labels:" + boardID }
func boardSettingsKey(boardID string) string   { return "board_settings:" + boardID }
func userLabelsKey(userID string) string       { return "user_labels:" + userID }

// ── Generic get/set/del ────────────────────────────────────────────────

// Get возвращает данные из кеша. found=false при cache miss.
func (c *DataCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

// Set сохраняет данные в кеш с TTL.
func (c *DataCache) Set(ctx context.Context, key string, data []byte) error {
	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Del удаляет ключ из кеша (инвалидация).
func (c *DataCache) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// ── Board Labels ───────────────────────────────────────────────────────

func (c *DataCache) GetBoardLabels(ctx context.Context, boardID string) ([]byte, bool, error) {
	return c.Get(ctx, boardLabelsKey(boardID))
}

func (c *DataCache) SetBoardLabels(ctx context.Context, boardID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Set(ctx, boardLabelsKey(boardID), b)
}

func (c *DataCache) InvalidateBoardLabels(ctx context.Context, boardID string) error {
	return c.Del(ctx, boardLabelsKey(boardID))
}

// ── Board Settings ─────────────────────────────────────────────────────

func (c *DataCache) GetBoardSettings(ctx context.Context, boardID string) ([]byte, bool, error) {
	return c.Get(ctx, boardSettingsKey(boardID))
}

func (c *DataCache) SetBoardSettings(ctx context.Context, boardID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Set(ctx, boardSettingsKey(boardID), b)
}

func (c *DataCache) InvalidateBoardSettings(ctx context.Context, boardID string) error {
	return c.Del(ctx, boardSettingsKey(boardID))
}

// ── User Labels ────────────────────────────────────────────────────────

func (c *DataCache) GetUserLabels(ctx context.Context, userID string) ([]byte, bool, error) {
	return c.Get(ctx, userLabelsKey(userID))
}

func (c *DataCache) SetUserLabels(ctx context.Context, userID string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.Set(ctx, userLabelsKey(userID), b)
}

func (c *DataCache) InvalidateUserLabels(ctx context.Context, userID string) error {
	return c.Del(ctx, userLabelsKey(userID))
}

// Close закрывает Redis connection.
func (c *DataCache) Close() error {
	return c.client.Close()
}
