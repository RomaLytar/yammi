package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/services/board/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

const (
	cacheMaxAckPending = 500
	cacheAckWait       = 30 * time.Second
)

// CacheConsumer синхронизирует Redis кеш membership через NATS JetStream.
// При старте: flush Redis → replay вся история (DeliverAll) → real-time sync.
// Паттерн аналогичен notification/infrastructure/nats/cache_consumers.go.
type CacheConsumer struct {
	nc    *nats.Conn
	js    nats.JetStreamContext
	cache *cache.MembershipCache
}

func NewCacheConsumer(natsURL string, c *cache.MembershipCache) (*CacheConsumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("cache consumer: connect to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("cache consumer: get jetstream: %w", err)
	}

	return &CacheConsumer{nc: nc, js: js, cache: c}, nil
}

// cacheInstanceID — уникальный ID инстанса для durable consumer names.
func cacheInstanceID() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// Start очищает Redis, сбрасывает durable consumers и подписывается с DeliverAll
// для полного replay истории событий. После replay — real-time sync.
func (c *CacheConsumer) Start() error {
	// 1. Сбрасываем durable consumers чтобы DeliverAll начал с начала
	c.resetConsumers()

	// 2. Flush Redis — чистый старт перед replay
	if err := c.cache.Flush(context.Background()); err != nil {
		log.Printf("cache consumer: flush warning: %v", err)
	}

	// 3. Ensure BOARDS stream exists (идемпотентно)
	if err := c.ensureBoardsStream(); err != nil {
		return err
	}

	// 4. Subscribe с DeliverAll — replay полной истории + real-time
	if err := c.subscribeMemberAdded(); err != nil {
		return err
	}
	if err := c.subscribeMemberRemoved(); err != nil {
		return err
	}
	if err := c.subscribeBoardDeleted(); err != nil {
		return err
	}

	log.Println("cache consumer: started (membership sync via NATS)")
	return nil
}

func (c *CacheConsumer) ensureBoardsStream() error {
	_, err := c.js.AddStream(&nats.StreamConfig{
		Name:     "BOARDS",
		Subjects: []string{"board.>", "column.>", "card.>", "member.>", "attachment.>"},
		MaxAge:   30 * 24 * time.Hour,
	})
	if err != nil && err.Error() != "stream name already in use" && err != nats.ErrStreamNameAlreadyInUse {
		// Попробуем update (stream существует с другой конфигурацией)
		_, err = c.js.UpdateStream(&nats.StreamConfig{
			Name:     "BOARDS",
			Subjects: []string{"board.>", "column.>", "card.>", "member.>", "attachment.>"},
			MaxAge:   30 * 24 * time.Hour,
		})
		if err != nil {
			return fmt.Errorf("cache consumer: ensure BOARDS stream: %w", err)
		}
	}
	return nil
}

func (c *CacheConsumer) resetConsumers() {
	consumers := []string{
		"board-cache-member-added-v1-" + cacheInstanceID(),
		"board-cache-member-removed-v1-" + cacheInstanceID(),
		"board-cache-board-deleted-v1-" + cacheInstanceID(),
	}
	for _, name := range consumers {
		_ = c.js.DeleteConsumer("BOARDS", name)
	}
}

func (c *CacheConsumer) subscribeMemberAdded() error {
	_, err := c.js.Subscribe("member.added", func(msg *nats.Msg) {
		var event usecase.MemberAdded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack() // poison message — skip
			return
		}
		_ = c.cache.SetMember(context.Background(), event.BoardID, event.UserID, event.Role)
		msg.Ack()
	},
		nats.Durable("board-cache-member-added-v1-"+cacheInstanceID()),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(cacheMaxAckPending),
		nats.AckWait(cacheAckWait),
	)
	if err != nil {
		return fmt.Errorf("cache consumer: subscribe member.added: %w", err)
	}
	log.Println("cache consumer: subscribed to member.added")
	return nil
}

func (c *CacheConsumer) subscribeMemberRemoved() error {
	_, err := c.js.Subscribe("member.removed", func(msg *nats.Msg) {
		var event usecase.MemberRemoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.cache.RemoveMember(context.Background(), event.BoardID, event.UserID)
		msg.Ack()
	},
		nats.Durable("board-cache-member-removed-v1-"+cacheInstanceID()),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(cacheMaxAckPending),
		nats.AckWait(cacheAckWait),
	)
	if err != nil {
		return fmt.Errorf("cache consumer: subscribe member.removed: %w", err)
	}
	log.Println("cache consumer: subscribed to member.removed")
	return nil
}

func (c *CacheConsumer) subscribeBoardDeleted() error {
	_, err := c.js.Subscribe("board.deleted", func(msg *nats.Msg) {
		var event usecase.BoardDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.cache.RemoveBoard(context.Background(), event.BoardID)
		msg.Ack()
	},
		nats.Durable("board-cache-board-deleted-v1-"+cacheInstanceID()),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(cacheMaxAckPending),
		nats.AckWait(cacheAckWait),
	)
	if err != nil {
		return fmt.Errorf("cache consumer: subscribe board.deleted: %w", err)
	}
	log.Println("cache consumer: subscribed to board.deleted")
	return nil
}

// Close закрывает NATS соединение (drain in-flight messages).
func (c *CacheConsumer) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}
