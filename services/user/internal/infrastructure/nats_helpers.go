package infrastructure

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"github.com/nats-io/nats.go"

	"github.com/romanlovesweed/yammi/pkg/events"
)

const (
	maxBackoff   = 30 * time.Second
	jitterFactor = 0.2 // ±20%
)

func (c *NATSConsumer) sendToDLQ(msg *nats.Msg, originalSubject, errMsg string) {
	numDelivered := uint64(0)
	if meta, err := msg.Metadata(); err == nil {
		numDelivered = meta.NumDelivered
	}

	envelope := events.DLQEnvelope{
		OriginalSubject: originalSubject,
		ConsumerName:    consumerCreated,
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

	log.Printf("sent to DLQ: subject=%s error=%s deliveries=%d",
		events.DLQSubject(originalSubject), errMsg, numDelivered)
	msg.Ack()
}

func backoffDelay(attempt uint64) time.Duration {
	delay := time.Duration(1<<attempt) * time.Second
	if delay > maxBackoff {
		delay = maxBackoff
	}
	jitter := time.Duration(float64(delay) * jitterFactor * (2*rand.Float64() - 1))
	return delay + jitter
}
