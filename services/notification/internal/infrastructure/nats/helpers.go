package nats

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
	"github.com/RomaLytar/yammi/services/notification/internal/domain"
	"github.com/RomaLytar/yammi/services/notification/internal/infrastructure/metrics"
)

const (
	maxBackoff   = 30 * time.Second
	jitterFactor = 0.2 // +-20%
)

func (c *Consumer) sendToDLQ(msg *nats.Msg, originalSubject, consumerName, errMsg string) {
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
		msg.Nak()
		return
	}

	_, err = c.js.Publish(events.DLQSubject(originalSubject), data)
	if err != nil {
		log.Printf("ERROR: failed to publish to DLQ: %v", err)
		msg.Nak()
		return
	}

	metrics.EventsDLQ.WithLabelValues(originalSubject).Inc()
	log.Printf("sent to DLQ: subject=%s error=%s deliveries=%d",
		events.DLQSubject(originalSubject), errMsg, numDelivered)
	msg.Ack()
}

func backoffDelay(attempt uint64) time.Duration {
	if attempt > 30 {
		return maxBackoff + time.Duration(float64(maxBackoff)*jitterFactor*(2*rand.Float64()-1))
	}
	delay := time.Duration(1<<attempt) * time.Second
	if delay > maxBackoff || delay <= 0 {
		delay = maxBackoff
	}
	jitter := time.Duration(float64(delay) * jitterFactor * (2*rand.Float64() - 1))
	return delay + jitter
}

func isStreamAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "stream name already in use" ||
		msg == "nats: stream name already in use" ||
		err == nats.ErrStreamNameAlreadyInUse
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
// Используется для direct-уведомлений (welcome, member_added, member_removed).
func (c *Consumer) createNotification(ctx context.Context, userID string, ntype domain.NotificationType, title, message string, metadata map[string]string) error {
	err := c.createUC.Execute(ctx, userID, ntype, title, message, metadata)
	if err == nil {
		metrics.NotificationsCreated.WithLabelValues(string(ntype)).Inc()
	}
	return err
}

// notifyBoardMembers создаёт один board event и инкрементирует счётчики участников.
// Event-sourcing: 1 event → 1 INSERT board_events + N Redis INCR.
func (c *Consumer) notifyBoardMembers(ctx context.Context, boardID, actorID string, ntype domain.NotificationType, title, message string, metadata map[string]string) {
	// Добавляем имя актора в metadata
	if actorID != "" {
		if actorName := c.nameCache.GetUserName(ctx, actorID); actorName != "" {
			metadata["actor_name"] = actorName
		}
	}

	if err := c.createUC.CreateBoardEvent(ctx, boardID, actorID, ntype, title, message, metadata); err != nil {
		log.Printf("failed to create board event for board %s: %v", boardID, err)
		return
	}

	metrics.BoardEventsCreated.WithLabelValues(string(ntype)).Inc()
}
