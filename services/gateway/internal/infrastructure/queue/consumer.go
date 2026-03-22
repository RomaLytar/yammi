package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	ws "github.com/romanlovesweed/yammi/services/gateway/internal/delivery/websocket"
)

// Consumer подписывается на события NATS и маршрутизирует их через Hub.
type Consumer struct {
	nc  *nats.Conn
	hub *ws.Hub
}

// boardEvent — общая структура для извлечения board_id, event_id, actor_id из события.
type boardEvent struct {
	EventID string `json:"event_id"`
	BoardID string `json:"board_id"`
	ActorID string `json:"actor_id"`
	OwnerID string `json:"owner_id"`
}

// memberEvent — событие добавления/удаления участника.
type memberEvent struct {
	EventID string `json:"event_id"`
	BoardID string `json:"board_id"`
	UserID  string `json:"user_id"`
	ActorID string `json:"actor_id"`
}

// notificationEvent — событие создания нотификации.
type notificationEvent struct {
	EventID string `json:"event_id"`
	UserID  string `json:"user_id"`
}

// wsMessage — формат сообщения для WebSocket-клиента.
type wsMessage struct {
	Type    string          `json:"type"`
	EventID string          `json:"event_id"`
	BoardID string          `json:"board_id,omitempty"`
	Data    json.RawMessage `json:"data"`
}

// NewConsumer создаёт подключение к NATS и возвращает Consumer.
func NewConsumer(natsURL string, hub *ws.Hub) (*Consumer, error) {
	var nc *nats.Conn
	var lastErr error

	// Retry подключения к NATS с backoff.
	for attempt := 1; attempt <= 10; attempt++ {
		nc, lastErr = nats.Connect(natsURL,
			nats.Name("ws-gateway"),
			nats.ReconnectWait(2*time.Second),
			nats.MaxReconnects(-1), // бесконечный reconnect
			nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
				log.Printf("nats: disconnected: %v", err)
			}),
			nats.ReconnectHandler(func(_ *nats.Conn) {
				log.Println("nats: reconnected")
			}),
		)
		if lastErr == nil {
			break
		}

		backoff := time.Duration(attempt) * 2 * time.Second
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		log.Printf("nats: connect attempt %d/10 failed: %v, retrying in %s", attempt, lastErr, backoff)
		time.Sleep(backoff)
	}

	if lastErr != nil && nc == nil {
		return nil, fmt.Errorf("nats: failed to connect after 10 attempts: %w", lastErr)
	}

	log.Printf("nats: connected to %s", natsURL)
	return &Consumer{nc: nc, hub: hub}, nil
}

// Start подписывается на все необходимые события.
// Board service публикует через plain NATS, поэтому используем обычные подписки.
func (c *Consumer) Start() error {
	// Board events — рассылка подписчикам доски.
	boardSubjects := []string{
		"board.created",
		"board.updated",
		"board.deleted",
		"column.created",
		"column.updated",
		"column.deleted",
		"columns.reordered",
		"card.created",
		"card.updated",
		"card.moved",
		"card.deleted",
	}

	for _, subj := range boardSubjects {
		subject := subj // capture for closure
		if _, err := c.nc.Subscribe(subject, func(msg *nats.Msg) {
			c.handleBoardEvent(subject, msg.Data)
		}); err != nil {
			return fmt.Errorf("subscribe to %s: %w", subject, err)
		}
		log.Printf("nats: subscribed to %s", subject)
	}

	// Member events — рассылка подписчикам доски + персональная доставка пользователю.
	memberSubjects := []string{
		"member.added",
		"member.removed",
	}

	for _, subj := range memberSubjects {
		subject := subj
		if _, err := c.nc.Subscribe(subject, func(msg *nats.Msg) {
			c.handleMemberEvent(subject, msg.Data)
		}); err != nil {
			return fmt.Errorf("subscribe to %s: %w", subject, err)
		}
		log.Printf("nats: subscribed to %s", subject)
	}

	// Notification events — персональная доставка пользователю.
	if _, err := c.nc.Subscribe("notification.created", func(msg *nats.Msg) {
		c.handleNotificationEvent(msg.Data)
	}); err != nil {
		return fmt.Errorf("subscribe to notification.created: %w", err)
	}
	log.Println("nats: subscribed to notification.created")

	return nil
}

// Close закрывает соединение с NATS.
func (c *Consumer) Close() {
	if c.nc != nil {
		c.nc.Drain()
	}
}

func (c *Consumer) handleBoardEvent(subject string, data []byte) {
	var evt boardEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		log.Printf("nats: failed to parse %s event: %v", subject, err)
		return
	}

	msg := wsMessage{
		Type:    subject,
		EventID: evt.EventID,
		BoardID: evt.BoardID,
		Data:    json.RawMessage(data),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("nats: failed to marshal ws message for %s: %v", subject, err)
		return
	}

	// Определяем актора для исключения.
	actorID := evt.ActorID

	// board.created — отправляем владельцу (он мог создать доску с другой вкладки).
	if subject == "board.created" {
		c.hub.SendToUser(evt.OwnerID, payload)
		return
	}

	c.hub.BroadcastToBoard(evt.BoardID, payload, actorID)
}

func (c *Consumer) handleMemberEvent(subject string, data []byte) {
	var evt memberEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		log.Printf("nats: failed to parse %s event: %v", subject, err)
		return
	}

	msg := wsMessage{
		Type:    subject,
		EventID: evt.EventID,
		BoardID: evt.BoardID,
		Data:    json.RawMessage(data),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("nats: failed to marshal ws message for %s: %v", subject, err)
		return
	}

	// Рассылаем подписчикам доски + отдельно целевому пользователю.
	c.hub.BroadcastToBoardAndUser(evt.BoardID, evt.UserID, payload, evt.ActorID)
}

func (c *Consumer) handleNotificationEvent(data []byte) {
	var evt notificationEvent
	if err := json.Unmarshal(data, &evt); err != nil {
		log.Printf("nats: failed to parse notification.created event: %v", err)
		return
	}

	msg := wsMessage{
		Type:    "notification",
		EventID: evt.EventID,
		Data:    json.RawMessage(data),
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		log.Printf("nats: failed to marshal ws notification message: %v", err)
		return
	}

	c.hub.SendToUser(evt.UserID, payload)
}
