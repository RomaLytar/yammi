package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// UnreadCounter — lazy read cache для unread count.
// НЕ обновляется на write (0 fan-out). Заполняется на read, инвалидируется на mark-read.
type UnreadCounter struct {
	client *redis.Client
}

func NewUnreadCounter(redisURL string) (*UnreadCounter, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &UnreadCounter{client: client}, nil
}

func (u *UnreadCounter) unreadKey(userID string) string {
	return "unread:" + userID
}

// Get возвращает кэшированный unread count. Возвращает -1 при cache miss.
func (u *UnreadCounter) Get(ctx context.Context, userID string) (int, error) {
	val, err := u.client.Get(ctx, u.unreadKey(userID)).Result()
	if err == redis.Nil {
		return -1, nil // cache miss
	}
	if err != nil {
		return -1, err
	}
	count, _ := strconv.Atoi(val)
	return count, nil
}

// Set кэширует вычисленный unread count (без TTL — инвалидируется через Invalidate).
func (u *UnreadCounter) Set(ctx context.Context, userID string, count int) error {
	return u.client.Set(ctx, u.unreadKey(userID), count, 0).Err()
}

// Invalidate удаляет кэш (вызывается при mark read / mark all read).
// Следующий Get вернёт -1, usecase пересчитает из SQL и закэширует.
func (u *UnreadCounter) Invalidate(ctx context.Context, userID string) error {
	return u.client.Del(ctx, u.unreadKey(userID)).Err()
}

func (u *UnreadCounter) Close() error {
	return u.client.Close()
}
