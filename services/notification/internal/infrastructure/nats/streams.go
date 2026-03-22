package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
)

// ensureStream создаёт или обновляет JetStream stream.
func (c *Consumer) ensureStream(cfg *nats.StreamConfig) error {
	_, err := c.js.AddStream(cfg)
	if err == nil {
		return nil
	}
	if isStreamAlreadyExists(err) {
		// Stream существует — обновляем конфигурацию (например, MaxAge)
		_, err = c.js.UpdateStream(cfg)
		if err != nil {
			return fmt.Errorf("update stream %s: %w", cfg.Name, err)
		}
		return nil
	}
	return fmt.Errorf("add stream %s: %w", cfg.Name, err)
}

// ensureStreams создаёт необходимые JetStream стримы.
func (c *Consumer) ensureStreams() error {
	if err := c.ensureStream(&nats.StreamConfig{
		Name:     events.StreamUsers,
		Subjects: []string{"user.>"},
		MaxAge:   30 * 24 * time.Hour,
	}); err != nil {
		return err
	}

	if err := c.ensureStream(&nats.StreamConfig{
		Name:     events.StreamBoards,
		Subjects: []string{"board.>", "column.>", "card.>", "member.>"},
		MaxAge:   30 * 24 * time.Hour,
	}); err != nil {
		return err
	}

	if err := c.ensureStream(&nats.StreamConfig{
		Name:     events.StreamNotifications,
		Subjects: []string{"notification.>"},
		MaxAge:   30 * 24 * time.Hour,
	}); err != nil {
		return err
	}

	if err := c.ensureStream(&nats.StreamConfig{
		Name:     events.StreamDLQ,
		Subjects: []string{"dlq.>"},
		MaxAge:   30 * 24 * time.Hour,
	}); err != nil {
		return err
	}

	return nil
}
