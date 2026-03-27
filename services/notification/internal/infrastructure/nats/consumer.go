package nats

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/cache"
	"github.com/RomaLytar/yammi/services/notification/internal/usecase"
)

// Consumer versioning: инкремент при изменении логики обработки.
const (
	// v3: QueueSubscribe для consumer groups (параллельная обработка N инстансами)
	consumerUserCreated    = "notification-service-user-created-v3"
	consumerBoardCreated   = "notification-service-board-created-v3"
	consumerBoardUpdated   = "notification-service-board-updated-v3"
	consumerBoardDeleted   = "notification-service-board-deleted-v3"
	consumerColumnCreated  = "notification-service-column-created-v3"
	consumerColumnUpdated  = "notification-service-column-updated-v3"
	consumerColumnDeleted  = "notification-service-column-deleted-v3"
	consumerCardCreated    = "notification-service-card-created-v3"
	consumerCardUpdated    = "notification-service-card-updated-v3"
	consumerCardMoved      = "notification-service-card-moved-v3"
	consumerCardDeleted      = "notification-service-card-deleted-v3"
	consumerCardAssigned     = "notification-service-card-assigned-v1"
	consumerCardUnassigned   = "notification-service-card-unassigned-v1"
	consumerMemberAdded      = "notification-service-member-added-v3"
	consumerMemberRemoved    = "notification-service-member-removed-v3"
	consumerSettingsUpdated       = "notification-service-settings-updated-v2"
	consumerAttachmentUploaded   = "notification-service-attachment-uploaded-v1"
	consumerAttachmentDeleted    = "notification-service-attachment-deleted-v1"

	maxDeliveries = 7
	maxAckPending = 500
	ackWait       = 30 * time.Second
)

// NameCache — интерфейс для кеша имён досок, пользователей и карточек.
type NameCache interface {
	SetBoardName(ctx context.Context, boardID, title string) error
	GetBoardName(ctx context.Context, boardID string) string
	DeleteBoardName(ctx context.Context, boardID string) error
	SetUserName(ctx context.Context, userID, name string) error
	GetUserName(ctx context.Context, userID string) string
	SetCardName(ctx context.Context, cardID, title string) error
	GetCardName(ctx context.Context, cardID string) string
	DeleteCardName(ctx context.Context, cardID string) error
	SetColumnName(ctx context.Context, columnID, title string) error
	GetColumnName(ctx context.Context, columnID string) string
	DeleteColumnName(ctx context.Context, columnID string) error
	TruncateCache(ctx context.Context) error
}

// workerPoolSize определяет количество параллельных goroutine для обработки событий.
// Один consumer с pool=50 обрабатывает 50 событий одновременно.
const workerPoolSize = 50

type Consumer struct {
	nc            *nats.Conn
	js            nats.JetStreamContext
	createUC      *usecase.CreateNotificationUseCase
	memberRepo    usecase.BoardMemberRepository
	nameCache     NameCache
	settingsCache *cache.SettingsCache
	pool          chan struct{} // semaphore для ограничения параллельных goroutine
}

// SetCreateUC обновляет usecase без пересоздания NATS-соединения.
func (c *Consumer) SetCreateUC(createUC *usecase.CreateNotificationUseCase) {
	c.createUC = createUC
}

func NewConsumer(natsURL string, createUC *usecase.CreateNotificationUseCase, memberRepo usecase.BoardMemberRepository, nameCache NameCache, settingsCache *cache.SettingsCache) (*Consumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("get jetstream context: %w", err)
	}

	c := &Consumer{nc: nc, js: js, createUC: createUC, memberRepo: memberRepo, nameCache: nameCache, settingsCache: settingsCache, pool: make(chan struct{}, workerPoolSize)}

	if err := c.ensureStreams(); err != nil {
		nc.Close()
		return nil, err
	}

	log.Printf("nats consumer connected to %s", natsURL)
	return c, nil
}

func (c *Consumer) JetStream() nats.JetStreamContext {
	return c.js
}

func (c *Consumer) Start() error {
	// Запускаем кеш-наполнители (DeliverAll) — проигрывают все события
	// для заполнения board_names, user_names, board_members
	cacheSubscribers := []func() error{
		c.cacheUserNames,
		c.cacheBoardNames,
		c.cacheBoardMembers,
		c.cacheColumnNames,
		c.cacheCardNames,
	}
	for _, sub := range cacheSubscribers {
		if err := sub(); err != nil {
			return err
		}
	}

	// Затем запускаем основные consumers (DeliverNew) — создают нотификации
	subscribers := []func() error{
		c.subscribeUserCreated,
		c.subscribeBoardCreated,
		c.subscribeBoardUpdated,
		c.subscribeBoardDeleted,
		c.subscribeColumnCreated,
		c.subscribeColumnUpdated,
		c.subscribeColumnDeleted,
		c.subscribeCardCreated,
		c.subscribeCardUpdated,
		c.subscribeCardMoved,
		c.subscribeCardDeleted,
		c.subscribeCardAssigned,
		c.subscribeCardUnassigned,
		c.subscribeMemberAdded,
		c.subscribeMemberRemoved,
		c.subscribeAttachmentUploaded,
		c.subscribeAttachmentDeleted,
		c.subscribeSettingsUpdated,
	}

	for _, sub := range subscribers {
		if err := sub(); err != nil {
			return err
		}
	}

	return nil
}

// resetCacheConsumers удаляет durable cache consumers из JetStream.
// При следующей подписке они будут созданы заново с DeliverAll — полный replay.
// Это гарантирует что кеш (board_members, names) всегда синхронизирован с событиями.
func (c *Consumer) resetCacheConsumers() {
	cacheConsumers := []struct {
		stream   string
		consumer string
	}{
		{"BOARDS", "notification-cache-board-created-v1"},
		{"BOARDS", "notification-cache-board-updated-v1"},
		{"BOARDS", "notification-cache-member-added-v1"},
		{"BOARDS", "notification-cache-member-removed-v1"},
		{"BOARDS", "notification-cache-column-created-v1"},
		{"BOARDS", "notification-cache-column-updated-v1"},
		{"BOARDS", "notification-cache-card-created-v1"},
		{"BOARDS", "notification-cache-card-updated-v1"},
		{"USERS", "notification-cache-users-v1"},
	}

	for _, cc := range cacheConsumers {
		if err := c.js.DeleteConsumer(cc.stream, cc.consumer); err != nil {
			// Consumer может не существовать при первом запуске — это нормально
			log.Printf("cache consumer reset: %s/%s: %v", cc.stream, cc.consumer, err)
		}
	}

	// Очищаем кеш-таблицы чтобы не было дубликатов при replay
	if err := c.memberRepo.TruncateCache(context.Background()); err != nil {
		log.Printf("WARNING: failed to truncate board_members cache: %v", err)
	}
	if err := c.nameCache.TruncateCache(context.Background()); err != nil {
		log.Printf("WARNING: failed to truncate name cache: %v", err)
	}

	log.Println("cache consumers reset — full replay on subscribe")
}

// parallel оборачивает NATS handler в goroutine pool.
// NATS доставляет до MaxAckPending(500) сообщений, pool ограничивает
// параллельную обработку до workerPoolSize goroutine.
func (c *Consumer) parallel(handler func(*nats.Msg)) func(*nats.Msg) {
	return func(msg *nats.Msg) {
		c.pool <- struct{}{}        // acquire slot (блокируется если pool полон)
		go func() {
			defer func() { <-c.pool }() // release slot
			handler(msg)
		}()
	}
}

func (c *Consumer) Close() {
	c.nc.Close()
}
