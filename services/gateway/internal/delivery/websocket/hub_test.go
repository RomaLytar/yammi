package websocket

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

// fakeClient создаёт клиента без реального WebSocket-соединения.
func fakeClient(hub *Hub, userID string) *Client {
	return &Client{
		hub:    hub,
		userID: userID,
		send:   make(chan []byte, sendBufferSize),
		boards: make(map[string]bool),
	}
}

// startHub запускает хаб и возвращает его. Для тестов.
func startHub() *Hub {
	hub := NewHub()
	go hub.Run()
	return hub
}

// drain читает все сообщения из канала send за timeout.
func drain(ch chan []byte, timeout time.Duration) [][]byte {
	var msgs [][]byte
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case msg := <-ch:
			msgs = append(msgs, msg)
		case <-timer.C:
			return msgs
		}
	}
}

func TestHub_RegisterUnregister(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1")
	c2 := fakeClient(hub, "user-1")
	c3 := fakeClient(hub, "user-2")

	// Регистрируем клиентов.
	hub.register <- c1
	hub.register <- c2
	hub.register <- c3
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if len(hub.clients["user-1"]) != 2 {
		t.Errorf("expected 2 connections for user-1, got %d", len(hub.clients["user-1"]))
	}
	if len(hub.clients["user-2"]) != 1 {
		t.Errorf("expected 1 connection for user-2, got %d", len(hub.clients["user-2"]))
	}
	hub.mu.RUnlock()

	// Отключаем одного клиента user-1.
	hub.unregister <- c1
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if len(hub.clients["user-1"]) != 1 {
		t.Errorf("expected 1 connection for user-1 after unregister, got %d", len(hub.clients["user-1"]))
	}
	hub.mu.RUnlock()

	// Отключаем последнего клиента user-1.
	hub.unregister <- c2
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if _, ok := hub.clients["user-1"]; ok {
		t.Error("expected user-1 to be removed from clients map")
	}
	hub.mu.RUnlock()
}

func TestHub_BoardSubscription(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1")
	c2 := fakeClient(hub, "user-2")

	hub.register <- c1
	hub.register <- c2
	time.Sleep(50 * time.Millisecond)

	// Подписываем на доску.
	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c2, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-2"}
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if len(hub.boards["board-1"]) != 2 {
		t.Errorf("expected 2 subscribers for board-1, got %d", len(hub.boards["board-1"]))
	}
	if len(hub.boards["board-2"]) != 1 {
		t.Errorf("expected 1 subscriber for board-2, got %d", len(hub.boards["board-2"]))
	}
	if len(c1.boards) != 2 {
		t.Errorf("expected c1 subscribed to 2 boards, got %d", len(c1.boards))
	}
	hub.mu.RUnlock()

	// Отписываем c2 от board-1.
	hub.unsubscribe <- &Subscription{Client: c2, BoardID: "board-1"}
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if len(hub.boards["board-1"]) != 1 {
		t.Errorf("expected 1 subscriber for board-1 after unsubscribe, got %d", len(hub.boards["board-1"]))
	}
	hub.mu.RUnlock()

	// Отписываем c1 от board-1 — доска должна быть удалена из map.
	hub.unsubscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	if _, ok := hub.boards["board-1"]; ok {
		t.Error("expected board-1 to be removed from boards map")
	}
	hub.mu.RUnlock()
}

func TestHub_UnregisterCleansUpBoardSubscriptions(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1")
	hub.register <- c1
	time.Sleep(50 * time.Millisecond)

	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-2"}
	time.Sleep(50 * time.Millisecond)

	// Отключение клиента должно очистить все его подписки.
	hub.unregister <- c1
	time.Sleep(50 * time.Millisecond)

	hub.mu.RLock()
	defer hub.mu.RUnlock()

	if _, ok := hub.boards["board-1"]; ok {
		t.Error("expected board-1 subscription to be cleaned up")
	}
	if _, ok := hub.boards["board-2"]; ok {
		t.Error("expected board-2 subscription to be cleaned up")
	}
}

func TestHub_BroadcastToBoard(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1")
	c2 := fakeClient(hub, "user-2")
	c3 := fakeClient(hub, "user-3") // не подписан

	hub.register <- c1
	hub.register <- c2
	hub.register <- c3
	time.Sleep(50 * time.Millisecond)

	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c2, BoardID: "board-1"}
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"type":"card.created","board_id":"board-1"}`)
	hub.BroadcastToBoard("board-1", msg, "")
	time.Sleep(50 * time.Millisecond)

	// c1 и c2 должны получить сообщение.
	if len(c1.send) != 1 {
		t.Errorf("expected c1 to receive 1 message, got %d", len(c1.send))
	}
	if len(c2.send) != 1 {
		t.Errorf("expected c2 to receive 1 message, got %d", len(c2.send))
	}
	// c3 не подписан — не должен получить.
	if len(c3.send) != 0 {
		t.Errorf("expected c3 to receive 0 messages, got %d", len(c3.send))
	}
}

func TestHub_BroadcastToBoard_ExcludesActor(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1") // актор
	c2 := fakeClient(hub, "user-2")

	hub.register <- c1
	hub.register <- c2
	time.Sleep(50 * time.Millisecond)

	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c2, BoardID: "board-1"}
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"type":"card.updated","board_id":"board-1"}`)
	hub.BroadcastToBoard("board-1", msg, "user-1")
	time.Sleep(50 * time.Millisecond)

	// user-1 (актор) не должен получить.
	if len(c1.send) != 0 {
		t.Errorf("expected actor c1 to receive 0 messages, got %d", len(c1.send))
	}
	// user-2 должен получить.
	if len(c2.send) != 1 {
		t.Errorf("expected c2 to receive 1 message, got %d", len(c2.send))
	}
}

func TestHub_SendToUser(t *testing.T) {
	hub := startHub()

	// У user-1 две вкладки (два соединения).
	c1a := fakeClient(hub, "user-1")
	c1b := fakeClient(hub, "user-1")
	c2 := fakeClient(hub, "user-2")

	hub.register <- c1a
	hub.register <- c1b
	hub.register <- c2
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"type":"notification","data":{"message":"hello"}}`)
	hub.SendToUser("user-1", msg)
	time.Sleep(50 * time.Millisecond)

	// Обе вкладки user-1 должны получить.
	if len(c1a.send) != 1 {
		t.Errorf("expected c1a to receive 1 message, got %d", len(c1a.send))
	}
	if len(c1b.send) != 1 {
		t.Errorf("expected c1b to receive 1 message, got %d", len(c1b.send))
	}
	// user-2 не должен получить.
	if len(c2.send) != 0 {
		t.Errorf("expected c2 to receive 0 messages, got %d", len(c2.send))
	}
}

func TestHub_SendToUser_NotConnected(t *testing.T) {
	hub := startHub()

	// Отправка пользователю, который не подключён — не должно паниковать.
	msg := []byte(`{"type":"notification"}`)
	hub.SendToUser("unknown-user", msg)
	time.Sleep(50 * time.Millisecond)
	// Тест проходит, если нет паники.
}

func TestHub_BroadcastToBoardAndUser(t *testing.T) {
	hub := startHub()

	c1 := fakeClient(hub, "user-1") // подписан на board
	c2 := fakeClient(hub, "user-2") // подписан на board (актор)
	c3 := fakeClient(hub, "user-3") // НЕ подписан на board, но целевой пользователь

	hub.register <- c1
	hub.register <- c2
	hub.register <- c3
	time.Sleep(50 * time.Millisecond)

	hub.subscribe <- &Subscription{Client: c1, BoardID: "board-1"}
	hub.subscribe <- &Subscription{Client: c2, BoardID: "board-1"}
	time.Sleep(50 * time.Millisecond)

	msg := []byte(`{"type":"member.added","board_id":"board-1"}`)
	hub.BroadcastToBoardAndUser("board-1", "user-3", msg, "user-2")
	time.Sleep(50 * time.Millisecond)

	// c1 подписан — должен получить.
	if len(c1.send) != 1 {
		t.Errorf("expected c1 to receive 1 message, got %d", len(c1.send))
	}
	// c2 — актор, исключён из рассылки по доске.
	if len(c2.send) != 0 {
		t.Errorf("expected c2 (actor) to receive 0 messages, got %d", len(c2.send))
	}
	// c3 — целевой пользователь, получает напрямую.
	if len(c3.send) != 1 {
		t.Errorf("expected c3 (target user) to receive 1 message, got %d", len(c3.send))
	}
}

func TestHub_ConcurrentOperations(t *testing.T) {
	hub := startHub()

	const numUsers = 50
	const numBoardsPerUser = 5

	var wg sync.WaitGroup
	clients := make([]*Client, numUsers)

	// Параллельная регистрация клиентов.
	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c := fakeClient(hub, "user-"+string(rune('A'+idx)))
			clients[idx] = c
			hub.register <- c
		}(i)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Параллельная подписка на доски.
	for i := 0; i < numUsers; i++ {
		for j := 0; j < numBoardsPerUser; j++ {
			wg.Add(1)
			go func(clientIdx, boardIdx int) {
				defer wg.Done()
				boardID := "board-" + string(rune('0'+boardIdx))
				hub.subscribe <- &Subscription{Client: clients[clientIdx], BoardID: boardID}
			}(i, j)
		}
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Параллельная рассылка.
	for j := 0; j < numBoardsPerUser; j++ {
		wg.Add(1)
		go func(boardIdx int) {
			defer wg.Done()
			boardID := "board-" + string(rune('0'+boardIdx))
			msg := []byte(`{"type":"test","board_id":"` + boardID + `"}`)
			hub.BroadcastToBoard(boardID, msg, "")
		}(j)
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	// Проверяем, что каждый клиент получил ровно numBoardsPerUser сообщений.
	for i := 0; i < numUsers; i++ {
		received := len(clients[i].send)
		if received != numBoardsPerUser {
			t.Errorf("client %d: expected %d messages, got %d", i, numBoardsPerUser, received)
		}
	}
}

func TestHub_ClientMessage_Subscribe(t *testing.T) {
	// Проверяем парсинг клиентского сообщения.
	raw := `{"type":"subscribe","board_id":"board-123"}`
	var msg struct {
		Type    string `json:"type"`
		BoardID string `json:"board_id,omitempty"`
	}
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if msg.Type != "subscribe" {
		t.Errorf("expected type=subscribe, got %s", msg.Type)
	}
	if msg.BoardID != "board-123" {
		t.Errorf("expected board_id=board-123, got %s", msg.BoardID)
	}
}

func TestHub_ServerMessage_Format(t *testing.T) {
	// Проверяем формат серверного сообщения.
	msg := struct {
		Type    string          `json:"type"`
		EventID string          `json:"event_id"`
		BoardID string          `json:"board_id,omitempty"`
		Data    json.RawMessage `json:"data,omitempty"`
	}{
		Type:    "card.created",
		EventID: "evt-123",
		BoardID: "board-456",
		Data:    json.RawMessage(`{"card_id":"card-789"}`),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if parsed["type"] != "card.created" {
		t.Errorf("expected type=card.created, got %v", parsed["type"])
	}
	if parsed["event_id"] != "evt-123" {
		t.Errorf("expected event_id=evt-123, got %v", parsed["event_id"])
	}
	if parsed["board_id"] != "board-456" {
		t.Errorf("expected board_id=board-456, got %v", parsed["board_id"])
	}
}
