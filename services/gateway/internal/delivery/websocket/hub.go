package websocket

import (
	"log"
	"sync"
)

// Hub управляет всеми WebSocket-соединениями и маршрутизирует сообщения.
type Hub struct {
	// clients — все подключённые клиенты по userID (один пользователь может иметь несколько вкладок).
	clients map[string]map[*Client]bool

	// boards — подписки на доски: board_id → множество клиентов.
	boards map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	subscribe  chan *Subscription
	unsubscribe chan *Subscription
	broadcast  chan *BroadcastMessage

	mu sync.RWMutex
}

// Subscription — запрос на подписку/отписку клиента от доски.
type Subscription struct {
	Client  *Client
	BoardID string
}

// BroadcastMessage — сообщение для рассылки.
type BroadcastMessage struct {
	// BoardID — для board-специфичных событий (рассылка подписчикам доски).
	BoardID string
	// UserID — для user-специфичных событий (нотификации).
	UserID string
	// Data — JSON payload.
	Data []byte
	// ExcludeUserID — не отправлять этому пользователю (актору).
	ExcludeUserID string
}

func NewHub() *Hub {
	return &Hub{
		clients:     make(map[string]map[*Client]bool),
		boards:      make(map[string]map[*Client]bool),
		register:    make(chan *Client, 256),
		unregister:  make(chan *Client, 256),
		subscribe:   make(chan *Subscription, 256),
		unsubscribe: make(chan *Subscription, 256),
		broadcast:   make(chan *BroadcastMessage, 1024),
	}
}

// Run — основной цикл обработки каналов. Запускать в отдельной горутине.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.addClient(client)

		case client := <-h.unregister:
			h.removeClient(client)

		case sub := <-h.subscribe:
			h.addBoardSubscription(sub)

		case sub := <-h.unsubscribe:
			h.removeBoardSubscription(sub)

		case msg := <-h.broadcast:
			h.handleBroadcast(msg)
		}
	}
}

func (h *Hub) addClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]bool)
	}
	h.clients[client.userID][client] = true

	log.Printf("hub: client registered, user=%s, total_connections=%d", client.userID, h.totalClients())
}

func (h *Hub) removeClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Удаляем из всех подписок на доски.
	for boardID := range client.boards {
		if subscribers, ok := h.boards[boardID]; ok {
			delete(subscribers, client)
			if len(subscribers) == 0 {
				delete(h.boards, boardID)
			}
		}
	}

	// Удаляем из списка клиентов пользователя.
	if userClients, ok := h.clients[client.userID]; ok {
		delete(userClients, client)
		if len(userClients) == 0 {
			delete(h.clients, client.userID)
		}
	}

	close(client.send)
	log.Printf("hub: client unregistered, user=%s, total_connections=%d", client.userID, h.totalClients())
}

func (h *Hub) addBoardSubscription(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.boards[sub.BoardID] == nil {
		h.boards[sub.BoardID] = make(map[*Client]bool)
	}
	h.boards[sub.BoardID][sub.Client] = true
	sub.Client.boards[sub.BoardID] = true

	log.Printf("hub: user=%s subscribed to board=%s", sub.Client.userID, sub.BoardID)
}

func (h *Hub) removeBoardSubscription(sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, ok := h.boards[sub.BoardID]; ok {
		delete(subscribers, sub.Client)
		if len(subscribers) == 0 {
			delete(h.boards, sub.BoardID)
		}
	}
	delete(sub.Client.boards, sub.BoardID)

	log.Printf("hub: user=%s unsubscribed from board=%s", sub.Client.userID, sub.BoardID)
}

func (h *Hub) handleBroadcast(msg *BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.BoardID != "" {
		h.broadcastToBoard(msg.BoardID, msg.Data, msg.ExcludeUserID)
	}

	if msg.UserID != "" {
		h.sendToUser(msg.UserID, msg.Data)
	}
}

func (h *Hub) broadcastToBoard(boardID string, data []byte, excludeUserID string) {
	subscribers, ok := h.boards[boardID]
	if !ok {
		return
	}

	for client := range subscribers {
		if excludeUserID != "" && client.userID == excludeUserID {
			continue
		}
		select {
		case client.send <- data:
		default:
			// Буфер переполнен — клиент не успевает читать.
			log.Printf("hub: dropping message for slow client, user=%s", client.userID)
		}
	}
}

func (h *Hub) sendToUser(userID string, data []byte) {
	userClients, ok := h.clients[userID]
	if !ok {
		return
	}

	for client := range userClients {
		select {
		case client.send <- data:
		default:
			log.Printf("hub: dropping message for slow client, user=%s", client.userID)
		}
	}
}

// BroadcastToBoard отправляет сообщение всем подписчикам доски, кроме excludeUserID.
func (h *Hub) BroadcastToBoard(boardID string, data []byte, excludeUserID string) {
	h.broadcast <- &BroadcastMessage{
		BoardID:       boardID,
		ExcludeUserID: excludeUserID,
		Data:          data,
	}
}

// SendToUser отправляет сообщение всем соединениям конкретного пользователя.
func (h *Hub) SendToUser(userID string, data []byte) {
	h.broadcast <- &BroadcastMessage{
		UserID: userID,
		Data:   data,
	}
}

// BroadcastToBoardAndUser отправляет сообщение подписчикам доски и дополнительно конкретному пользователю.
func (h *Hub) BroadcastToBoardAndUser(boardID string, userID string, data []byte, excludeUserID string) {
	h.broadcast <- &BroadcastMessage{
		BoardID:       boardID,
		UserID:        userID,
		ExcludeUserID: excludeUserID,
		Data:          data,
	}
}

func (h *Hub) totalClients() int {
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}
