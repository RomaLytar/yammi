package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait — максимальное время на запись сообщения.
	writeWait = 10 * time.Second
	// pongWait — максимальное время ожидания pong от клиента.
	pongWait = 60 * time.Second
	// pingPeriod — период отправки ping (должен быть меньше pongWait).
	pingPeriod = 54 * time.Second
	// maxMessageSize — максимальный размер входящего сообщения.
	maxMessageSize = 4096
	// sendBufferSize — размер буфера канала отправки.
	sendBufferSize = 256
)

// Client — одно WebSocket-соединение.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	userID  string
	token   string // JWT токен для проверки доступа к доскам
	checker BoardAccessChecker
	send    chan []byte
	// boards — множество досок, на которые подписан клиент (для быстрой очистки при отключении).
	boards map[string]bool
}

// clientMessage — входящее сообщение от клиента.
type clientMessage struct {
	Type    string `json:"type"`
	BoardID string `json:"board_id,omitempty"`
}

// serverMessage — исходящее сообщение для клиента.
type serverMessage struct {
	Type    string          `json:"type"`
	EventID string          `json:"event_id,omitempty"`
	BoardID string          `json:"board_id,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// NewClient создаёт нового клиента.
func NewClient(hub *Hub, conn *websocket.Conn, userID, token string, checker BoardAccessChecker) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		userID:  userID,
		token:   token,
		checker: checker,
		send:    make(chan []byte, sendBufferSize),
		boards:  make(map[string]bool),
	}
}

// ReadPump читает сообщения из WebSocket. Обрабатывает subscribe/unsubscribe/ping.
// Запускать в отдельной горутине. При завершении — отключает клиента от хаба.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ws-client: read error user=%s: %v", c.userID, err)
			}
			return
		}

		var msg clientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("ws-client: invalid message from user=%s: %v", c.userID, err)
			continue
		}

		switch msg.Type {
		case "subscribe":
			if msg.BoardID == "" {
				continue
			}
			// Проверяем, является ли пользователь членом доски
			if c.checker != nil && !c.checker.IsMember(msg.BoardID, c.token) {
				errMsg, _ := json.Marshal(serverMessage{Type: "error", BoardID: msg.BoardID, Data: json.RawMessage(`"access denied"`)})
				select {
				case c.send <- errMsg:
				default:
				}
				continue
			}
			c.hub.subscribe <- &Subscription{Client: c, BoardID: msg.BoardID}

		case "unsubscribe":
			if msg.BoardID == "" {
				continue
			}
			c.hub.unsubscribe <- &Subscription{Client: c, BoardID: msg.BoardID}

		case "ping":
			pong, _ := json.Marshal(serverMessage{Type: "pong"})
			select {
			case c.send <- pong:
			default:
			}

		default:
			log.Printf("ws-client: unknown message type=%q from user=%s", msg.Type, c.userID)
		}
	}
}

// WritePump отправляет сообщения из канала send в WebSocket.
// Запускать в отдельной горутине.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub закрыл канал — соединение завершается.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

			// Отправляем накопленные сообщения — каждое отдельным WebSocket frame (валидный JSON).
			n := len(c.send)
			for i := 0; i < n; i++ {
				if err := c.conn.WriteMessage(websocket.TextMessage, <-c.send); err != nil {
					return
				}
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
