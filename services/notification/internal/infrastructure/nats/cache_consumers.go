package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
)

// instanceID возвращает hostname для уникальных имён cache-консьюмеров.
// Каждый инстанс должен проигрывать всю историю (DeliverAll) независимо.
func instanceID() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// --- Cache-only consumers (DeliverAll — проигрывают историю для наполнения кешей) ---

func (c *Consumer) cacheUserNames() error {
	_, err := c.js.Subscribe(events.SubjectUserCreated, func(msg *nats.Msg) {
		var event events.UserCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetUserName(context.Background(), event.UserID, event.Name)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-users-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache user names: %w", err)
	}
	log.Println("cache consumer started: user names")
	return nil
}

func (c *Consumer) cacheBoardNames() error {
	// board.created → сохраняем имя доски
	_, err := c.js.Subscribe(events.SubjectBoardCreated, func(msg *nats.Msg) {
		var event events.BoardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetBoardName(context.Background(), event.BoardID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-board-created-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache board created: %w", err)
	}

	// board.updated → обновляем имя доски
	_, err = c.js.Subscribe(events.SubjectBoardUpdated, func(msg *nats.Msg) {
		var event events.BoardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetBoardName(context.Background(), event.BoardID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-board-updated-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache board updated: %w", err)
	}

	log.Println("cache consumer started: board names")
	return nil
}

func (c *Consumer) cacheBoardMembers() error {
	// member.added → добавляем в кеш
	_, err := c.js.Subscribe(events.SubjectMemberAdded, func(msg *nats.Msg) {
		var event events.MemberAdded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.memberRepo.AddMember(context.Background(), event.BoardID, event.UserID)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-member-added-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache member added: %w", err)
	}

	// member.removed → удаляем из кеша
	_, err = c.js.Subscribe(events.SubjectMemberRemoved, func(msg *nats.Msg) {
		var event events.MemberRemoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.memberRepo.RemoveMember(context.Background(), event.BoardID, event.UserID)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-member-removed-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache member removed: %w", err)
	}

	log.Println("cache consumer started: board members")
	return nil
}

func (c *Consumer) cacheColumnNames() error {
	_, err := c.js.Subscribe(events.SubjectColumnCreated, func(msg *nats.Msg) {
		var event events.ColumnCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetColumnName(context.Background(), event.ColumnID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-column-created-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache column created: %w", err)
	}

	_, err = c.js.Subscribe(events.SubjectColumnUpdated, func(msg *nats.Msg) {
		var event events.ColumnUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetColumnName(context.Background(), event.ColumnID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-column-updated-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache column updated: %w", err)
	}

	log.Println("cache consumer started: column names")
	return nil
}

func (c *Consumer) cacheCardNames() error {
	_, err := c.js.Subscribe(events.SubjectCardCreated, func(msg *nats.Msg) {
		var event events.CardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetCardName(context.Background(), event.CardID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-card-created-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache card created: %w", err)
	}

	_, err = c.js.Subscribe(events.SubjectCardUpdated, func(msg *nats.Msg) {
		var event events.CardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			msg.Ack()
			return
		}
		_ = c.nameCache.SetCardName(context.Background(), event.CardID, event.Title)
		msg.Ack()
	},
		nats.Durable(fmt.Sprintf("notification-cache-card-updated-v1-%s", instanceID())),
		nats.ManualAck(),
		nats.DeliverAll(),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe cache card updated: %w", err)
	}

	log.Println("cache consumer started: card names")
	return nil
}
