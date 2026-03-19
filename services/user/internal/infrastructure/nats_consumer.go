package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/romanlovesweed/yammi/pkg/events"
	"github.com/romanlovesweed/yammi/services/user/internal/domain"
	"github.com/romanlovesweed/yammi/services/user/internal/usecase"
)

// Consumer versioning: инкремент версии при изменении логики обработки событий.
// v4 — текущая стабильная версия (создание профиля с idempotency + DLQ).
// v1 — первая версия (удаление профиля).
// NATS создаёт отдельный durable consumer для каждой версии,
// старые consumers остаются в JetStream до ручной очистки.
const (
	consumerCreated = "user-service-user-created-v4"
	consumerDeleted = "user-service-user-deleted-v1"
	maxDeliveries   = 7
	maxAckPending   = 500
	ackWait         = 30 * time.Second
)

type NATSConsumer struct {
	nc *nats.Conn
	js nats.JetStreamContext
	uc *usecase.UserUseCase
}

func NewNATSConsumer(natsURL string, uc *usecase.UserUseCase) (*NATSConsumer, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("get jetstream context: %w", err)
	}

	// Ensure USERS stream (может уже существовать — другая реплика создала)
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamUsers,
		Subjects: []string{"user.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil && !isStreamAlreadyExists(err) {
		nc.Close()
		return nil, fmt.Errorf("ensure users stream: %w", err)
	}

	// Ensure DLQ stream (30 days retention for investigation)
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
	return &NATSConsumer{nc: nc, js: js, uc: uc}, nil
}

func (c *NATSConsumer) JetStream() nats.JetStreamContext {
	return c.js
}

func (c *NATSConsumer) Start() error {
	if err := c.subscribeUserCreated(); err != nil {
		return err
	}
	if err := c.subscribeUserDeleted(); err != nil {
		return err
	}
	return nil
}

func (c *NATSConsumer) Close() {
	c.nc.Close()
}

func (c *NATSConsumer) subscribeUserCreated() error {
	_, err := c.js.Subscribe(events.SubjectUserCreated, func(msg *nats.Msg) {
		meta, metaErr := msg.Metadata()

		var event events.UserCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message, sending to DLQ: %v", err)
			c.sendToDLQ(msg, events.SubjectUserCreated, consumerCreated, err.Error())
			return
		}

		if err := c.uc.CreateProfile(context.Background(), event.UserID, event.Email, event.Name); err != nil {
			if errors.Is(err, domain.ErrEmailExists) {
				msg.Ack()
				return
			}

			numDelivered := uint64(1)
			if metaErr == nil {
				numDelivered = meta.NumDelivered
			}

			if numDelivered >= maxDeliveries {
				log.Printf("max retries (%d) exhausted for user %s, sending to DLQ: %v",
					maxDeliveries, event.UserID, err)
				c.sendToDLQ(msg, events.SubjectUserCreated, consumerCreated, err.Error())
				return
			}

			delay := backoffDelay(numDelivered)
			log.Printf("retry %d/%d for user %s in %s: %v",
				numDelivered, maxDeliveries, event.UserID, delay, err)
			msg.NakWithDelay(delay)
			return
		}

		log.Printf("created profile for user %s (%s)", event.UserID, event.Email)
		msg.Ack()
	},
		nats.Durable(consumerCreated),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)

	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectUserCreated, err)
	}

	log.Printf("consumer started: %s (maxDeliver=%d, maxAckPending=%d, ackWait=%s)",
		consumerCreated, maxDeliveries, maxAckPending, ackWait)
	return nil
}

func (c *NATSConsumer) subscribeUserDeleted() error {
	_, err := c.js.Subscribe(events.SubjectUserDeleted, func(msg *nats.Msg) {
		var event events.UserDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on user.deleted, sending to DLQ: %v", err)
			c.sendToDLQ(msg, events.SubjectUserDeleted, consumerDeleted, err.Error())
			return
		}

		if err := c.uc.DeleteProfile(context.Background(), event.UserID); err != nil {
			if errors.Is(err, domain.ErrUserNotFound) {
				msg.Ack()
				return
			}
			log.Printf("failed to delete profile for user %s: %v", event.UserID, err)
			msg.Nak()
			return
		}

		log.Printf("deleted profile for user %s", event.UserID)
		msg.Ack()
	},
		nats.Durable(consumerDeleted),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxDeliver(maxDeliveries),
		nats.AckWait(ackWait),
	)

	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectUserDeleted, err)
	}

	log.Printf("consumer started: %s", consumerDeleted)
	return nil
}

func isStreamAlreadyExists(err error) bool {
	return err != nil && err.Error() == "stream name already in use"
}
