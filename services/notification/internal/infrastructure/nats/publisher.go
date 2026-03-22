package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

type Publisher struct {
	js nats.JetStreamContext
}

func NewPublisher(js nats.JetStreamContext) *Publisher {
	return &Publisher{js: js}
}

func (p *Publisher) PublishNotificationCreated(ctx context.Context, n *domain.Notification) error {
	event := events.NotificationCreated{
		EventID:      uuid.New().String(),
		EventVersion: 1,
		OccurredAt:   time.Now(),
		ID:           n.ID,
		UserID:       n.UserID,
		Type:         string(n.Type),
		Title:        n.Title,
		Message:      n.Message,
		Metadata:     n.Metadata,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal notification event: %w", err)
	}

	_, err = p.js.Publish(events.SubjectNotificationCreated, data)
	if err != nil {
		return fmt.Errorf("publish notification.created: %w", err)
	}

	return nil
}

func (p *Publisher) PublishNotificationsBatch(ctx context.Context, notifications []*domain.Notification) error {
	for _, n := range notifications {
		if err := p.PublishNotificationCreated(ctx, n); err != nil {
			return err
		}
	}
	return nil
}

// PublishBoardEventNotification публикует ОДНО событие для WebSocket push.
// Gateway рассылает подписчикам доски (вместо N отдельных notification.created).
func (p *Publisher) PublishBoardEventNotification(ctx context.Context, event *domain.BoardEvent) error {
	// Формат совместимый с фронтендом (те же поля что notification.created)
	payload := map[string]interface{}{
		"id":         event.ID,
		"event_id":   event.ID,
		"board_id":   event.BoardID,
		"actor_id":   event.ActorID,
		"type":       string(event.EventType),
		"title":      event.Title,
		"message":    event.Message,
		"metadata":   event.Metadata,
		"created_at": event.CreatedAt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal board event notification: %w", err)
	}

	// Публикуем на отдельный subject — gateway подписан и broadcast'ит подписчикам board
	_, err = p.js.Publish("notification.board_event", data)
	if err != nil {
		return fmt.Errorf("publish notification.board_event: %w", err)
	}

	return nil
}

func (p *Publisher) PublishSettingsUpdated(ctx context.Context, userID string, enabled, realtimeEnabled bool) error {
	event := events.NotificationSettingsUpdated{
		EventID:         uuid.New().String(),
		EventVersion:    1,
		OccurredAt:      time.Now(),
		UserID:          userID,
		Enabled:         enabled,
		RealtimeEnabled: realtimeEnabled,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal settings event: %w", err)
	}

	_, err = p.js.Publish(events.SubjectNotificationSettingsUpdated, data)
	if err != nil {
		return fmt.Errorf("publish notification.settings.updated: %w", err)
	}

	return nil
}
