package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// UnreadCounter управляет счётчиками непрочитанных уведомлений в Redis.
// O(1) операции вместо SQL COUNT(*).
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

// Increment увеличивает счётчик непрочитанных для пользователя.
func (u *UnreadCounter) Increment(ctx context.Context, userID string) error {
	return u.client.Incr(ctx, u.unreadKey(userID)).Err()
}

// IncrementMany увеличивает счётчики для списка пользователей (pipeline).
func (u *UnreadCounter) IncrementMany(ctx context.Context, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}

	pipe := u.client.Pipeline()
	for _, uid := range userIDs {
		pipe.Incr(ctx, u.unreadKey(uid))
	}
	_, err := pipe.Exec(ctx)
	return err
}

// Get возвращает количество непрочитанных.
func (u *UnreadCounter) Get(ctx context.Context, userID string) (int, error) {
	val, err := u.client.Get(ctx, u.unreadKey(userID)).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	count, _ := strconv.Atoi(val)
	return count, nil
}

// Reset обнуляет счётчик (mark all as read).
func (u *UnreadCounter) Reset(ctx context.Context, userID string) error {
	return u.client.Set(ctx, u.unreadKey(userID), 0, 0).Err()
}

// Decrement уменьшает счётчик на 1 (mark single as read).
func (u *UnreadCounter) Decrement(ctx context.Context, userID string) error {
	result := u.client.Decr(ctx, u.unreadKey(userID))
	if result.Err() != nil {
		return result.Err()
	}
	// Не допускаем отрицательных значений
	if result.Val() < 0 {
		return u.client.Set(ctx, u.unreadKey(userID), 0, 0).Err()
	}
	return nil
}

func (u *UnreadCounter) Close() error {
	return u.client.Close()
}
