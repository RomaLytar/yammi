package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"

	"github.com/RomaLytar/yammi/pkg/events"
	"github.com/RomaLytar/yammi/services/notification/internal/domain"
)

// --- Member events ---

func (c *Consumer) subscribeMemberAdded() error {
	_, err := c.js.QueueSubscribe(events.SubjectMemberAdded, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.MemberAdded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectMemberAdded, err)
			c.sendToDLQ(msg, events.SubjectMemberAdded, consumerMemberAdded, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectMemberAdded, consumerMemberAdded, func() error {
			ctx := context.Background()
			// Обновляем кеш участников
			if err := c.memberRepo.AddMember(ctx, event.BoardID, event.UserID); err != nil {
				return fmt.Errorf("update members cache: %w", err)
			}
			// Уведомляем добавленного пользователя
			return c.createNotification(ctx, event.UserID, domain.TypeMemberAdded,
				fmt.Sprintf("Вы добавлены в доску \"%s\"", event.BoardTitle),
				fmt.Sprintf("Роль: %s", event.Role),
				map[string]string{"board_id": event.BoardID, "board_title": event.BoardTitle, "role": event.Role})
		})
	}),
		nats.Durable(consumerMemberAdded),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectMemberAdded, err)
	}
	log.Printf("consumer started: %s", consumerMemberAdded)
	return nil
}

func (c *Consumer) subscribeMemberRemoved() error {
	_, err := c.js.QueueSubscribe(events.SubjectMemberRemoved, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.MemberRemoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectMemberRemoved, err)
			c.sendToDLQ(msg, events.SubjectMemberRemoved, consumerMemberRemoved, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectMemberRemoved, consumerMemberRemoved, func() error {
			ctx := context.Background()
			// Уведомляем удалённого пользователя
			if err := c.createNotification(ctx, event.UserID, domain.TypeMemberRemoved,
				fmt.Sprintf("Вы удалены из доски \"%s\"", event.BoardTitle),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": event.BoardTitle}); err != nil {
				return err
			}
			// Обновляем кеш участников
			if err := c.memberRepo.RemoveMember(ctx, event.BoardID, event.UserID); err != nil {
				log.Printf("failed to remove member from cache for board %s: %v", event.BoardID, err)
			}
			return nil
		})
	}),
		nats.Durable(consumerMemberRemoved),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectMemberRemoved, err)
	}
	log.Printf("consumer started: %s", consumerMemberRemoved)
	return nil
}
