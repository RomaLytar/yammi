package infrastructure

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/romanlovesweed/yammi/pkg/events"
	"github.com/romanlovesweed/yammi/services/user/internal/domain"
	"github.com/romanlovesweed/yammi/services/user/internal/usecase"
)

const (
	consumerName  = "user-service-user-created-v4"
	maxDeliveries = 7
	maxAckPending = 500
	ackWait       = 30 * time.Second
	maxBackoff    = 30 * time.Second
	jitterFactor  = 0.2 // ±20%
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

	// Ensure USERS stream
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamUsers,
		Subjects: []string{"user.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("ensure users stream: %w", err)
	}

	// Ensure DLQ stream (30 days retention for investigation)
	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamDLQ,
		Subjects: []string{"dlq.>"},
		MaxAge:   30 * 24 * time.Hour,
	})
	if err != nil {
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
	_, err := c.js.Subscribe(events.SubjectUserCreated, func(msg *nats.Msg) {
		meta, metaErr := msg.Metadata()

		// Poison message: can't unmarshal — send to DLQ immediately
		var event events.UserCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message, sending to DLQ: %v", err)
			c.sendToDLQ(msg, events.SubjectUserCreated, err.Error())
			return
		}

		// Process
		if err := c.uc.CreateProfile(context.Background(), event.UserID, event.Email, event.Name); err != nil {
			// Permanent: idempotency — already exists
			if errors.Is(err, domain.ErrEmailExists) {
				msg.Ack()
				return
			}

			// Check delivery count
			numDelivered := uint64(1)
			if metaErr == nil {
				numDelivered = meta.NumDelivered
			}

			// Max retries exhausted — send to DLQ
			if numDelivered >= maxDeliveries {
				log.Printf("max retries (%d) exhausted for user %s, sending to DLQ: %v",
					maxDeliveries, event.UserID, err)
				c.sendToDLQ(msg, events.SubjectUserCreated, err.Error())
				return
			}

			// Retry with exponential backoff
			delay := backoffDelay(numDelivered)
			log.Printf("retry %d/%d for user %s in %s: %v",
				numDelivered, maxDeliveries, event.UserID, delay, err)
			msg.NakWithDelay(delay)
			return
		}

		log.Printf("created profile for user %s (%s)", event.UserID, event.Email)
		msg.Ack()
	},
		nats.Durable(consumerName),
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
		consumerName, maxDeliveries, maxAckPending, ackWait)
	return nil
}

func (c *NATSConsumer) sendToDLQ(msg *nats.Msg, originalSubject, errMsg string) {
	numDelivered := uint64(0)
	if meta, err := msg.Metadata(); err == nil {
		numDelivered = meta.NumDelivered
	}

	envelope := events.DLQEnvelope{
		OriginalSubject: originalSubject,
		ConsumerName:    consumerName,
		Error:           errMsg,
		NumDelivered:    numDelivered,
		Payload:         string(msg.Data),
		FailedAt:        time.Now(),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		log.Printf("ERROR: failed to marshal DLQ envelope: %v", err)
		// Don't ack — let NATS redeliver
		msg.Nak()
		return
	}

	_, err = c.js.Publish(events.DLQSubject(originalSubject), data)
	if err != nil {
		log.Printf("ERROR: failed to publish to DLQ: %v", err)
		// Don't ack — let NATS redeliver
		msg.Nak()
		return
	}

	log.Printf("sent to DLQ: subject=%s error=%s deliveries=%d",
		events.DLQSubject(originalSubject), errMsg, numDelivered)
	msg.Ack()
}

func (c *NATSConsumer) Close() {
	c.nc.Close()
}

func backoffDelay(attempt uint64) time.Duration {
	// base=2s, exponential: 2s, 4s, 8s, 16s, 30s, 30s
	delay := time.Duration(1<<attempt) * time.Second
	if delay > maxBackoff {
		delay = maxBackoff
	}
	// jitter ±20% to avoid thundering herd
	jitter := time.Duration(float64(delay) * jitterFactor * (2*rand.Float64() - 1))
	return delay + jitter
}
