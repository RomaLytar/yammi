package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/romanlovesweed/yammi/pkg/events"
	"github.com/romanlovesweed/yammi/services/notification/internal/domain"
	"github.com/romanlovesweed/yammi/services/notification/internal/infrastructure/cache"
	"github.com/romanlovesweed/yammi/services/notification/internal/infrastructure/metrics"
	"github.com/romanlovesweed/yammi/services/notification/internal/usecase"
)

// Consumer versioning: инкремент при изменении логики обработки.
const (
	consumerUserCreated    = "notification-service-user-created-v2"
	consumerBoardCreated   = "notification-service-board-created-v2"
	consumerBoardUpdated   = "notification-service-board-updated-v2"
	consumerBoardDeleted   = "notification-service-board-deleted-v2"
	consumerColumnCreated  = "notification-service-column-created-v2"
	consumerColumnUpdated  = "notification-service-column-updated-v2"
	consumerColumnDeleted  = "notification-service-column-deleted-v2"
	consumerCardCreated    = "notification-service-card-created-v2"
	consumerCardUpdated    = "notification-service-card-updated-v2"
	consumerCardMoved      = "notification-service-card-moved-v2"
	consumerCardDeleted    = "notification-service-card-deleted-v2"
	consumerMemberAdded    = "notification-service-member-added-v2"
	consumerMemberRemoved    = "notification-service-member-removed-v2"
	consumerSettingsUpdated  = "notification-service-settings-updated-v1"

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
}

type Consumer struct {
	nc            *nats.Conn
	js            nats.JetStreamContext
	createUC      *usecase.CreateNotificationUseCase
	memberRepo    usecase.BoardMemberRepository
	nameCache     NameCache
	settingsCache *cache.SettingsCache
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

	// Ensure USERS stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamUsers,
		Subjects: []string{"user.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil && !isStreamAlreadyExists(err) {
		nc.Close()
		return nil, fmt.Errorf("ensure users stream: %w", err)
	}

	// Ensure BOARDS stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamBoards,
		Subjects: []string{"board.>", "column.>", "card.>", "member.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil && !isStreamAlreadyExists(err) {
		nc.Close()
		return nil, fmt.Errorf("ensure boards stream: %w", err)
	}

	// Ensure NOTIFICATIONS stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamNotifications,
		Subjects: []string{"notification.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil && !isStreamAlreadyExists(err) {
		nc.Close()
		return nil, fmt.Errorf("ensure notifications stream: %w", err)
	}

	// Ensure DLQ stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamDLQ,
		Subjects: []string{"dlq.>"},
		MaxAge:   30 * 24 * time.Hour,
	})
	if err != nil && !isStreamAlreadyExists(err) {
		nc.Close()
		return nil, fmt.Errorf("ensure dlq stream: %w", err)
	}

	log.Printf("nats consumer connected to %s", natsURL)
	return &Consumer{nc: nc, js: js, createUC: createUC, memberRepo: memberRepo, nameCache: nameCache, settingsCache: settingsCache}, nil
}

func (c *Consumer) JetStream() nats.JetStreamContext {
	return c.js
}

func (c *Consumer) Start() error {
	// Сначала запускаем кеш-наполнители (DeliverAll) — проигрывают все старые события
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
		c.subscribeMemberAdded,
		c.subscribeMemberRemoved,
		c.subscribeSettingsUpdated,
	}

	for _, sub := range subscribers {
		if err := sub(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Consumer) Close() {
	c.nc.Close()
}

// handleWithRetry обрабатывает сообщение с ретраями и DLQ.
func (c *Consumer) handleWithRetry(msg *nats.Msg, subject, consumer string, handler func() error) {
	start := time.Now()

	if err := handler(); err != nil {
		metrics.EventErrors.WithLabelValues(subject).Inc()
		metrics.EventProcessingDuration.WithLabelValues(subject).Observe(time.Since(start).Seconds())

		meta, metaErr := msg.Metadata()
		numDelivered := uint64(1)
		if metaErr == nil {
			numDelivered = meta.NumDelivered
		}

		if numDelivered >= maxDeliveries {
			log.Printf("max retries (%d) exhausted for %s, sending to DLQ: %v",
				maxDeliveries, subject, err)
			c.sendToDLQ(msg, subject, consumer, err.Error())
			return
		}

		metrics.EventRetries.WithLabelValues(subject).Inc()
		delay := backoffDelay(numDelivered)
		log.Printf("retry %d/%d for %s in %s: %v",
			numDelivered, maxDeliveries, subject, delay, err)
		msg.NakWithDelay(delay)
		return
	}

	metrics.EventsConsumed.WithLabelValues(subject).Inc()
	metrics.EventProcessingDuration.WithLabelValues(subject).Observe(time.Since(start).Seconds())
	msg.Ack()
}

// createNotification обёртка над createUC.Execute с метриками.
func (c *Consumer) createNotification(ctx context.Context, userID string, ntype domain.NotificationType, title, message string, metadata map[string]string) error {
	err := c.createUC.Execute(ctx, userID, ntype, title, message, metadata)
	if err == nil {
		metrics.NotificationsCreated.WithLabelValues(string(ntype)).Inc()
	}
	return err
}

// notifyBoardMembers создаёт уведомления для всех участников доски, кроме актора.
// Использует batch INSERT + batch settings check для минимизации DB round-trips.
func (c *Consumer) notifyBoardMembers(ctx context.Context, boardID, actorID string, ntype domain.NotificationType, title, message string, metadata map[string]string) {
	memberIDs, err := c.memberRepo.ListMemberIDs(ctx, boardID)
	if err != nil {
		log.Printf("failed to list members for board %s: %v", boardID, err)
		return
	}

	// Добавляем имя актора в metadata
	if actorID != "" {
		if actorName := c.nameCache.GetUserName(ctx, actorID); actorName != "" {
			metadata["actor_name"] = actorName
		}
	}

	// Собираем batch-запросы, клонируя metadata для каждого участника
	var requests []usecase.NotificationRequest
	for _, memberID := range memberIDs {
		if memberID == actorID {
			continue
		}
		meta := make(map[string]string, len(metadata))
		for k, v := range metadata {
			meta[k] = v
		}
		requests = append(requests, usecase.NotificationRequest{
			UserID:   memberID,
			Type:     ntype,
			Title:    title,
			Message:  message,
			Metadata: meta,
		})
	}

	if len(requests) == 0 {
		return
	}

	created, err := c.createUC.BatchExecute(ctx, requests)
	if err != nil {
		log.Printf("failed to batch create notifications for board %s: %v", boardID, err)
		return
	}

	if created > 0 {
		metrics.NotificationsCreated.WithLabelValues(string(ntype)).Add(float64(created))
	}
}

// subscribeSettingsUpdated подписывается на изменения настроек для инвалидации кеша.
func (c *Consumer) subscribeSettingsUpdated() error {
	if c.settingsCache == nil {
		return nil
	}

	_, err := c.js.Subscribe(events.SubjectNotificationSettingsUpdated, func(msg *nats.Msg) {
		var event events.NotificationSettingsUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on settings.updated: %v", err)
			msg.Ack()
			return
		}
		c.settingsCache.Invalidate(event.UserID)
		msg.Ack()
	},
		nats.Durable(consumerSettingsUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(3),
		nats.MaxAckPending(100),
		nats.AckWait(10*time.Second),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectNotificationSettingsUpdated, err)
	}
	log.Printf("consumer started: %s", consumerSettingsUpdated)
	return nil
}

// --- User events ---

func (c *Consumer) subscribeUserCreated() error {
	_, err := c.js.Subscribe(events.SubjectUserCreated, func(msg *nats.Msg) {
		var event events.UserCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectUserCreated, err)
			c.sendToDLQ(msg, events.SubjectUserCreated, consumerUserCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectUserCreated, consumerUserCreated, func() error {
			ctx := context.Background()
			// Кешируем имя пользователя
			_ = c.nameCache.SetUserName(ctx, event.UserID, event.Name)
			return c.createNotification(ctx, event.UserID, domain.TypeWelcome,
				"Добро пожаловать в Yammi!",
				"Мы рады, что вы с нами. Начните с создания первой доски.",
				map[string]string{"user_name": event.Name})
		})
	},
		nats.Durable(consumerUserCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectUserCreated, err)
	}
	log.Printf("consumer started: %s", consumerUserCreated)
	return nil
}

// --- Board events ---

func (c *Consumer) subscribeBoardCreated() error {
	_, err := c.js.Subscribe(events.SubjectBoardCreated, func(msg *nats.Msg) {
		var event events.BoardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardCreated, err)
			c.sendToDLQ(msg, events.SubjectBoardCreated, consumerBoardCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardCreated, consumerBoardCreated, func() error {
			ctx := context.Background()
			// Кешируем имя доски
			_ = c.nameCache.SetBoardName(ctx, event.BoardID, event.Title)
			// Добавляем владельца в кеш участников
			if err := c.memberRepo.AddMember(ctx, event.BoardID, event.OwnerID); err != nil {
				return fmt.Errorf("add owner to members cache: %w", err)
			}
			return c.createNotification(ctx, event.OwnerID, domain.TypeBoardCreated,
				fmt.Sprintf("Доска \"%s\" создана", event.Title),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": event.Title})
		})
	},
		nats.Durable(consumerBoardCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardCreated, err)
	}
	log.Printf("consumer started: %s", consumerBoardCreated)
	return nil
}

func (c *Consumer) subscribeBoardUpdated() error {
	_, err := c.js.Subscribe(events.SubjectBoardUpdated, func(msg *nats.Msg) {
		var event events.BoardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardUpdated, err)
			c.sendToDLQ(msg, events.SubjectBoardUpdated, consumerBoardUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardUpdated, consumerBoardUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetBoardName(ctx, event.BoardID, event.Title)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeBoardUpdated,
				fmt.Sprintf("Доска \"%s\" обновлена", event.Title),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": event.Title})
			return nil
		})
	},
		nats.Durable(consumerBoardUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardUpdated, err)
	}
	log.Printf("consumer started: %s", consumerBoardUpdated)
	return nil
}

func (c *Consumer) subscribeBoardDeleted() error {
	_, err := c.js.Subscribe(events.SubjectBoardDeleted, func(msg *nats.Msg) {
		var event events.BoardDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardDeleted, err)
			c.sendToDLQ(msg, events.SubjectBoardDeleted, consumerBoardDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardDeleted, consumerBoardDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			if boardName == "" {
				boardName = "Без названия"
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeBoardDeleted,
				fmt.Sprintf("Доска \"%s\" удалена", boardName),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": boardName})
			// Очищаем кеши
			if err := c.memberRepo.RemoveAllByBoard(ctx, event.BoardID); err != nil {
				log.Printf("failed to cleanup members cache for board %s: %v", event.BoardID, err)
			}
			_ = c.nameCache.DeleteBoardName(ctx, event.BoardID)
			return nil
		})
	},
		nats.Durable(consumerBoardDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardDeleted, err)
	}
	log.Printf("consumer started: %s", consumerBoardDeleted)
	return nil
}

// --- Column events ---

func (c *Consumer) subscribeColumnCreated() error {
	_, err := c.js.Subscribe(events.SubjectColumnCreated, func(msg *nats.Msg) {
		var event events.ColumnCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnCreated, err)
			c.sendToDLQ(msg, events.SubjectColumnCreated, consumerColumnCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnCreated, consumerColumnCreated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetColumnName(ctx, event.ColumnID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			title := fmt.Sprintf("Колонка \"%s\" создана", event.Title)
			if boardName != "" {
				title = fmt.Sprintf("Колонка \"%s\" создана в доске \"%s\"", event.Title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnCreated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": event.Title, "board_title": boardName})
			return nil
		})
	},
		nats.Durable(consumerColumnCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnCreated, err)
	}
	log.Printf("consumer started: %s", consumerColumnCreated)
	return nil
}

func (c *Consumer) subscribeColumnUpdated() error {
	_, err := c.js.Subscribe(events.SubjectColumnUpdated, func(msg *nats.Msg) {
		var event events.ColumnUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnUpdated, err)
			c.sendToDLQ(msg, events.SubjectColumnUpdated, consumerColumnUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnUpdated, consumerColumnUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetColumnName(ctx, event.ColumnID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			title := fmt.Sprintf("Колонка \"%s\" обновлена", event.Title)
			if boardName != "" {
				title = fmt.Sprintf("Колонка \"%s\" обновлена в доске \"%s\"", event.Title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnUpdated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": event.Title, "board_title": boardName})
			return nil
		})
	},
		nats.Durable(consumerColumnUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnUpdated, err)
	}
	log.Printf("consumer started: %s", consumerColumnUpdated)
	return nil
}

func (c *Consumer) subscribeColumnDeleted() error {
	_, err := c.js.Subscribe(events.SubjectColumnDeleted, func(msg *nats.Msg) {
		var event events.ColumnDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnDeleted, err)
			c.sendToDLQ(msg, events.SubjectColumnDeleted, consumerColumnDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnDeleted, consumerColumnDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			columnName := c.nameCache.GetColumnName(ctx, event.ColumnID)
			title := "Колонка удалена"
			if columnName != "" && boardName != "" {
				title = fmt.Sprintf("Колонка \"%s\" удалена в доске \"%s\"", columnName, boardName)
			} else if columnName != "" {
				title = fmt.Sprintf("Колонка \"%s\" удалена", columnName)
			} else if boardName != "" {
				title = fmt.Sprintf("Колонка удалена в доске \"%s\"", boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnDeleted,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": columnName, "board_title": boardName})
			_ = c.nameCache.DeleteColumnName(ctx, event.ColumnID)
			return nil
		})
	},
		nats.Durable(consumerColumnDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnDeleted, err)
	}
	log.Printf("consumer started: %s", consumerColumnDeleted)
	return nil
}

// --- Card events ---

func (c *Consumer) subscribeCardCreated() error {
	_, err := c.js.Subscribe(events.SubjectCardCreated, func(msg *nats.Msg) {
		var event events.CardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardCreated, err)
			c.sendToDLQ(msg, events.SubjectCardCreated, consumerCardCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardCreated, consumerCardCreated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetCardName(ctx, event.CardID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			title := fmt.Sprintf("Карточка \"%s\" создана", event.Title)
			if boardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" создана в доске \"%s\"", event.Title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardCreated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "card_id": event.CardID, "card_title": event.Title, "board_title": boardName})
			return nil
		})
	},
		nats.Durable(consumerCardCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardCreated, err)
	}
	log.Printf("consumer started: %s", consumerCardCreated)
	return nil
}

func (c *Consumer) subscribeCardUpdated() error {
	_, err := c.js.Subscribe(events.SubjectCardUpdated, func(msg *nats.Msg) {
		var event events.CardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardUpdated, err)
			c.sendToDLQ(msg, events.SubjectCardUpdated, consumerCardUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardUpdated, consumerCardUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetCardName(ctx, event.CardID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			title := fmt.Sprintf("Карточка \"%s\" обновлена", event.Title)
			if boardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" обновлена в доске \"%s\"", event.Title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardUpdated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "card_id": event.CardID, "card_title": event.Title, "board_title": boardName})
			return nil
		})
	},
		nats.Durable(consumerCardUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardUpdated, err)
	}
	log.Printf("consumer started: %s", consumerCardUpdated)
	return nil
}

func (c *Consumer) subscribeCardMoved() error {
	_, err := c.js.Subscribe(events.SubjectCardMoved, func(msg *nats.Msg) {
		var event events.CardMoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardMoved, err)
			c.sendToDLQ(msg, events.SubjectCardMoved, consumerCardMoved, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardMoved, consumerCardMoved, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			title := "Карточка перемещена"
			if cardName != "" && boardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" перемещена в доске \"%s\"", cardName, boardName)
			} else if cardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" перемещена", cardName)
			} else if boardName != "" {
				title = fmt.Sprintf("Карточка перемещена в доске \"%s\"", boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardMoved,
				title, "",
				map[string]string{
					"board_id":         event.BoardID,
					"board_title":      boardName,
					"card_id":          event.CardID,
					"card_title":       cardName,
					"source_column_id": event.SourceColumnID,
					"target_column_id": event.TargetColumnID,
				})
			return nil
		})
	},
		nats.Durable(consumerCardMoved),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardMoved, err)
	}
	log.Printf("consumer started: %s", consumerCardMoved)
	return nil
}

func (c *Consumer) subscribeCardDeleted() error {
	_, err := c.js.Subscribe(events.SubjectCardDeleted, func(msg *nats.Msg) {
		var event events.CardDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardDeleted, err)
			c.sendToDLQ(msg, events.SubjectCardDeleted, consumerCardDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardDeleted, consumerCardDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			title := "Карточка удалена"
			if cardName != "" && boardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" удалена в доске \"%s\"", cardName, boardName)
			} else if cardName != "" {
				title = fmt.Sprintf("Карточка \"%s\" удалена", cardName)
			} else if boardName != "" {
				title = fmt.Sprintf("Карточка удалена в доске \"%s\"", boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardDeleted,
				title, "",
				map[string]string{"board_id": event.BoardID, "board_title": boardName, "card_id": event.CardID, "card_title": cardName})
			_ = c.nameCache.DeleteCardName(ctx, event.CardID)
			return nil
		})
	},
		nats.Durable(consumerCardDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardDeleted, err)
	}
	log.Printf("consumer started: %s", consumerCardDeleted)
	return nil
}

// --- Member events ---

func (c *Consumer) subscribeMemberAdded() error {
	_, err := c.js.Subscribe(events.SubjectMemberAdded, func(msg *nats.Msg) {
		var event events.MemberAdded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectMemberAdded, err)
			c.sendToDLQ(msg, events.SubjectMemberAdded, consumerMemberAdded, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectMemberAdded, consumerMemberAdded, func() error {
			ctx := context.Background()
			// Обновляем кеш участников
			if err := c.memberRepo.AddMember(ctx, event.BoardID, event.UserID); err != nil {
				return fmt.Errorf("update members cache: %w", err)
			}
			// Уведомляем добавленного пользователя
			return c.createNotification(ctx, event.UserID, domain.TypeMemberAdded,
				fmt.Sprintf("Вы добавлены в доску \"%s\"", event.BoardTitle),
				fmt.Sprintf("Роль: %s", event.Role),
				map[string]string{"board_id": event.BoardID, "board_title": event.BoardTitle, "role": event.Role})
		})
	},
		nats.Durable(consumerMemberAdded),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectMemberAdded, err)
	}
	log.Printf("consumer started: %s", consumerMemberAdded)
	return nil
}

func (c *Consumer) subscribeMemberRemoved() error {
	_, err := c.js.Subscribe(events.SubjectMemberRemoved, func(msg *nats.Msg) {
		var event events.MemberRemoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectMemberRemoved, err)
			c.sendToDLQ(msg, events.SubjectMemberRemoved, consumerMemberRemoved, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectMemberRemoved, consumerMemberRemoved, func() error {
			ctx := context.Background()
			// Уведомляем удалённого пользователя
			if err := c.createNotification(ctx, event.UserID, domain.TypeMemberRemoved,
				fmt.Sprintf("Вы удалены из доски \"%s\"", event.BoardTitle),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": event.BoardTitle}); err != nil {
				return err
			}
			// Обновляем кеш участников
			if err := c.memberRepo.RemoveMember(ctx, event.BoardID, event.UserID); err != nil {
				log.Printf("failed to remove member from cache for board %s: %v", event.BoardID, err)
			}
			return nil
		})
	},
		nats.Durable(consumerMemberRemoved),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectMemberRemoved, err)
	}
	log.Printf("consumer started: %s", consumerMemberRemoved)
	return nil
}

// --- Cache-only consumers (DeliverAll — проигрывают историю для наполнения кешей) ---

func (c *Consumer) cacheUserNames() error {
	_, err := c.js.Subscribe(events.SubjectUserCreated, func(msg *nats.Msg) {
		var event events.UserCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetUserName(context.Background(), event.UserID, event.Name)
		msg.Ack()
	},
		nats.Durable("notification-cache-users-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache user names: %w", err)
	}
	log.Println("cache consumer started: user names")
	return nil
}

func (c *Consumer) cacheBoardNames() error {
	// board.created → сохраняем имя доски
	_, err := c.js.Subscribe(events.SubjectBoardCreated, func(msg *nats.Msg) {
		var event events.BoardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetBoardName(context.Background(), event.BoardID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-board-created-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache board created: %w", err)
	}

	// board.updated → обновляем имя доски
	_, err = c.js.Subscribe(events.SubjectBoardUpdated, func(msg *nats.Msg) {
		var event events.BoardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetBoardName(context.Background(), event.BoardID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-board-updated-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache board updated: %w", err)
	}

	log.Println("cache consumer started: board names")
	return nil
}

func (c *Consumer) cacheBoardMembers() error {
	// member.added → добавляем в кеш
	_, err := c.js.Subscribe(events.SubjectMemberAdded, func(msg *nats.Msg) {
		var event events.MemberAdded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.memberRepo.AddMember(context.Background(), event.BoardID, event.UserID)
		msg.Ack()
	},
		nats.Durable("notification-cache-member-added-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache member added: %w", err)
	}

	// member.removed → удаляем из кеша
	_, err = c.js.Subscribe(events.SubjectMemberRemoved, func(msg *nats.Msg) {
		var event events.MemberRemoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.memberRepo.RemoveMember(context.Background(), event.BoardID, event.UserID)
		msg.Ack()
	},
		nats.Durable("notification-cache-member-removed-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache member removed: %w", err)
	}

	log.Println("cache consumer started: board members")
	return nil
}

func (c *Consumer) cacheColumnNames() error {
	_, err := c.js.Subscribe(events.SubjectColumnCreated, func(msg *nats.Msg) {
		var event events.ColumnCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetColumnName(context.Background(), event.ColumnID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-column-created-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache column created: %w", err)
	}

	_, err = c.js.Subscribe(events.SubjectColumnUpdated, func(msg *nats.Msg) {
		var event events.ColumnUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetColumnName(context.Background(), event.ColumnID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-column-updated-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache column updated: %w", err)
	}

	log.Println("cache consumer started: column names")
	return nil
}

func (c *Consumer) cacheCardNames() error {
	_, err := c.js.Subscribe(events.SubjectCardCreated, func(msg *nats.Msg) {
		var event events.CardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetCardName(context.Background(), event.CardID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-card-created-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache card created: %w", err)
	}

	_, err = c.js.Subscribe(events.SubjectCardUpdated, func(msg *nats.Msg) {
		var event events.CardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetCardName(context.Background(), event.CardID, event.Title)
		msg.Ack()
	},
		nats.Durable("notification-cache-card-updated-v1"),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache card updated: %w", err)
	}

	log.Println("cache consumer started: card names")
	return nil
}
