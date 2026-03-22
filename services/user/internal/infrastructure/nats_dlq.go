package infrastructure

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
)

type DLQMonitor struct {
	js  nats.JetStreamContext
	sub *nats.Subscription
}

func NewDLQMonitor(js nats.JetStreamContext) *DLQMonitor {
	return &DLQMonitor{js: js}
}

func (m *DLQMonitor) Start() error {
	sub, err := m.js.Subscribe("dlq.user.>", func(msg *nats.Msg) {
		var envelope events.DLQEnvelope
		if err := json.Unmarshal(msg.Data, &envelope); err != nil {
			log.Printf("DLQ MONITOR: failed to unmarshal envelope: %v", err)
			msg.Ack()
			return
		}

		log.Printf("DLQ ALERT: subject=%s consumer=%s error=%q deliveries=%d failed_at=%s payload=%s",
			envelope.OriginalSubject,
			envelope.ConsumerName,
			envelope.Error,
			envelope.NumDelivered,
			envelope.FailedAt.Format("2006-01-02T15:04:05Z"),
			string(envelope.Payload),
		)
		msg.Ack()
	}, nats.Durable("user-service-dlq-monitor"), nats.ManualAck(), nats.DeliverAll())

	if err != nil {
		return fmt.Errorf("subscribe to DLQ: %w", err)
	}

	m.sub = sub
	log.Printf("DLQ monitor started on dlq.user.>")
	return nil
}

func (m *DLQMonitor) Close() {
	if m.sub != nil {
		m.sub.Unsubscribe()
	}
}
