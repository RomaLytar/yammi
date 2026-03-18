package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/romanlovesweed/yammi/pkg/events"
)

type NATSPublisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func NewNATSPublisher(natsURL string) (*NATSPublisher, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("connect to nats: %w", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("get jetstream context: %w", err)
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     events.StreamUsers,
		Subjects: []string{"user.>"},
		MaxAge:   7 * 24 * time.Hour,
	})
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("ensure stream: %w", err)
	}

	log.Printf("nats publisher connected to %s", natsURL)
	return &NATSPublisher{nc: nc, js: js}, nil
}

func (p *NATSPublisher) PublishUserCreated(ctx context.Context, userID, email, name string) error {
	event := events.UserCreated{
		EventID:      uuid.New().String(),
		EventVersion: 1,
		OccurredAt:   time.Now(),
		UserID:       userID,
		Email:        email,
		Name:         name,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.js.Publish(events.SubjectUserCreated, data)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

func (p *NATSPublisher) PublishUserDeleted(ctx context.Context, userID string) error {
	event := events.UserDeleted{
		EventID:      uuid.New().String(),
		EventVersion: 1,
		OccurredAt:   time.Now(),
		UserID:       userID,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.js.Publish(events.SubjectUserDeleted, data)
	if err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

func (p *NATSPublisher) Close() {
	p.nc.Close()
}
