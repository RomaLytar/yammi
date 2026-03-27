package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

// --- User events ---

func (c *Consumer) subscribeUserCreated() error {
	_, err := c.js.QueueSubscribe(events.SubjectUserCreated, "notification-workers", c.parallel(func(msg *nats.Msg) {
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
	}),
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

// subscribeSettingsUpdated подписывается на изменения настроек для инвалидации кеша.
func (c *Consumer) subscribeSettingsUpdated() error {
	if c.settingsCache == nil {
		return nil
	}

	_, err := c.js.QueueSubscribe(events.SubjectNotificationSettingsUpdated, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.NotificationSettingsUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on settings.updated: %v", err)
			msg.Ack()
			return
		}
		c.settingsCache.Invalidate(event.UserID)
		msg.Ack()
	}),
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
