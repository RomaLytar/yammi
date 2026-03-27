package cache

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

// MembershipCache — event-driven Redis кеш для membership проверок.
// Данные поддерживаются в актуальном состоянии через NATS события (member.added/removed).
// Нет TTL — данные всегда свежие. PostgreSQL = fallback при Redis ошибках.
type MembershipCache struct {
	client *redis.Client
}

func NewMembershipCache(redisURL string) (*MembershipCache, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &MembershipCache{client: client}, nil
}

func boardRolesKey(boardID string) string { return "board_roles:" + boardID }
func userBoardsKey(userID string) string  { return "user_boards:" + userID }

// SetMember добавляет участника в кеш (HSET + SADD pipeline).
// Вызывается из NATS consumer при member.added.
func (c *MembershipCache) SetMember(ctx context.Context, boardID, userID, role string) error {
	pipe := c.client.Pipeline()
	pipe.HSet(ctx, boardRolesKey(boardID), userID, role)
	pipe.SAdd(ctx, userBoardsKey(userID), boardID)
	_, err := pipe.Exec(ctx)
	return err
}

// RemoveMember удаляет участника из кеша (HDEL + SREM pipeline).
// Вызывается из NATS consumer при member.removed.
func (c *MembershipCache) RemoveMember(ctx context.Context, boardID, userID string) error {
	pipe := c.client.Pipeline()
	pipe.HDel(ctx, boardRolesKey(boardID), userID)
	pipe.SRem(ctx, userBoardsKey(userID), boardID)
	_, err := pipe.Exec(ctx)
	return err
}

// GetRole возвращает роль участника из кеша.
// found=false если пользователь не найден в HASH (не участник или кеш не гидрирован).
func (c *MembershipCache) GetRole(ctx context.Context, boardID, userID string) (role string, found bool, err error) {
	val, err := c.client.HGet(ctx, boardRolesKey(boardID), userID).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return val, true, nil
}

// RemoveBoard удаляет все данные доски из кеша.
// 1. Получает всех member'ов (HGETALL)
// 2. Удаляет boardID из user_boards каждого member'а (SREM)
// 3. Удаляет board_roles hash (DEL)
// Вызывается из NATS consumer при board.deleted.
func (c *MembershipCache) RemoveBoard(ctx context.Context, boardID string) error {
	// Получаем всех member'ов чтобы почистить user_boards
	members, err := c.client.HGetAll(ctx, boardRolesKey(boardID)).Result()
	if err != nil {
		slog.Error("cache: failed to get board members for cleanup", "error", err, "board_id", boardID)
		// Удаляем хотя бы hash
		c.client.Del(ctx, boardRolesKey(boardID))
		return err
	}

	pipe := c.client.Pipeline()
	for userID := range members {
		pipe.SRem(ctx, userBoardsKey(userID), boardID)
	}
	pipe.Del(ctx, boardRolesKey(boardID))
	_, err = pipe.Exec(ctx)
	return err
}

// Flush очищает весь Redis DB (вызывается перед replay событий при старте).
func (c *MembershipCache) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Close закрывает Redis connection.
func (c *MembershipCache) Close() error {
	return c.client.Close()
}
