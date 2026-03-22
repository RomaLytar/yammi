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

// Set кэширует вычисленный unread count с TTL 60s.
// TTL гарантирует eventual consistency: worst case 60s stale.
// Инвалидируется раньше через Invalidate (mark read).
func (u *UnreadCounter) Set(ctx context.Context, userID string, count int) error {
	return u.client.Set(ctx, u.unreadKey(userID), count, 60*time.Second).Err()
}

// Invalidate удаляет кэш (вызывается при mark read / mark all read).
// Следующий Get вернёт -1, usecase пересчитает из SQL и закэширует.
func (u *UnreadCounter) Invalidate(ctx context.Context, userID string) error {
	return u.client.Del(ctx, u.unreadKey(userID)).Err()
}

// SetBoardSeq сохраняет последний event_seq для доски (1 SET на write, не fan-out).
func (u *UnreadCounter) SetBoardSeq(ctx context.Context, boardID string, seq int64) error {
	return u.client.Set(ctx, "board_seq:"+boardID, seq, 0).Err()
}

// GetBoardSeqs возвращает max event_seq для списка досок (1 MGET round-trip).
func (u *UnreadCounter) GetBoardSeqs(ctx context.Context, boardIDs []string) (map[string]int64, error) {
	if len(boardIDs) == 0 {
		return nil, nil
	}

	keys := make([]string, len(boardIDs))
	for i, id := range boardIDs {
		keys[i] = "board_seq:" + id
	}

	vals, err := u.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64, len(boardIDs))
	for i, v := range vals {
		if v != nil {
			if n, err := strconv.ParseInt(v.(string), 10, 64); err == nil {
				result[boardIDs[i]] = n
			}
		}
	}
	return result, nil
}

func (u *UnreadCounter) Close() error {
	return u.client.Close()
}
