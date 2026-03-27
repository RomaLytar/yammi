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

// --- Board events ---

func (c *Consumer) subscribeBoardCreated() error {
	_, err := c.js.QueueSubscribe(events.SubjectBoardCreated, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.BoardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardCreated, err)
			c.sendToDLQ(msg, events.SubjectBoardCreated, consumerBoardCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardCreated, consumerBoardCreated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetBoardName(ctx, event.BoardID, event.Title)
			if err := c.memberRepo.AddMember(ctx, event.BoardID, event.OwnerID); err != nil {
				return fmt.Errorf("add owner to members cache: %w", err)
			}
			return c.createNotification(ctx, event.OwnerID, domain.TypeBoardCreated,
				fmt.Sprintf("Доска \"%s\" создана", event.Title),
				"",
				map[string]string{"board_id": event.BoardID, "board_title": event.Title})
		})
	}),
		nats.Durable(consumerBoardCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardCreated, err)
	}
	log.Printf("consumer started: %s", consumerBoardCreated)
	return nil
}

func (c *Consumer) subscribeBoardUpdated() error {
	_, err := c.js.QueueSubscribe(events.SubjectBoardUpdated, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.BoardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardUpdated, err)
			c.sendToDLQ(msg, events.SubjectBoardUpdated, consumerBoardUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardUpdated, consumerBoardUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetBoardName(ctx, event.BoardID, event.Title)
			title := c.buildTitle(ctx, fmt.Sprintf("Доска \"%s\" обновлена", event.Title), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeBoardUpdated,
				title, "",
				map[string]string{"board_id": event.BoardID, "board_title": event.Title})
			return nil
		})
	}),
		nats.Durable(consumerBoardUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardUpdated, err)
	}
	log.Printf("consumer started: %s", consumerBoardUpdated)
	return nil
}

func (c *Consumer) subscribeBoardDeleted() error {
	_, err := c.js.QueueSubscribe(events.SubjectBoardDeleted, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.BoardDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectBoardDeleted, err)
			c.sendToDLQ(msg, events.SubjectBoardDeleted, consumerBoardDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectBoardDeleted, consumerBoardDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			if boardName == "" {
				boardName = "Без названия"
			}
			title := c.buildTitle(ctx, fmt.Sprintf("Доска \"%s\" удалена", boardName), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeBoardDeleted,
				title, "",
				map[string]string{"board_id": event.BoardID, "board_title": boardName})
			if err := c.memberRepo.RemoveAllByBoard(ctx, event.BoardID); err != nil {
				log.Printf("failed to cleanup members cache for board %s: %v", event.BoardID, err)
			}
			_ = c.nameCache.DeleteBoardName(ctx, event.BoardID)
			return nil
		})
	}),
		nats.Durable(consumerBoardDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectBoardDeleted, err)
	}
	log.Printf("consumer started: %s", consumerBoardDeleted)
	return nil
}

// --- Column events ---

func (c *Consumer) subscribeColumnCreated() error {
	_, err := c.js.QueueSubscribe(events.SubjectColumnCreated, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.ColumnCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnCreated, err)
			c.sendToDLQ(msg, events.SubjectColumnCreated, consumerColumnCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnCreated, consumerColumnCreated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetColumnName(ctx, event.ColumnID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)

			title := c.buildTitle(ctx, fmt.Sprintf("Колонка \"%s\" создана", event.Title), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnCreated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": event.Title, "board_title": boardName})
			return nil
		})
	}),
		nats.Durable(consumerColumnCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnCreated, err)
	}
	log.Printf("consumer started: %s", consumerColumnCreated)
	return nil
}

func (c *Consumer) subscribeColumnUpdated() error {
	_, err := c.js.QueueSubscribe(events.SubjectColumnUpdated, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.ColumnUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnUpdated, err)
			c.sendToDLQ(msg, events.SubjectColumnUpdated, consumerColumnUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnUpdated, consumerColumnUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetColumnName(ctx, event.ColumnID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)

			title := c.buildTitle(ctx, fmt.Sprintf("Колонка \"%s\" обновлена", event.Title), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnUpdated,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": event.Title, "board_title": boardName})
			return nil
		})
	}),
		nats.Durable(consumerColumnUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnUpdated, err)
	}
	log.Printf("consumer started: %s", consumerColumnUpdated)
	return nil
}

func (c *Consumer) subscribeColumnDeleted() error {
	_, err := c.js.QueueSubscribe(events.SubjectColumnDeleted, "notification-workers", c.parallel(func(msg *nats.Msg) {
		var event events.ColumnDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectColumnDeleted, err)
			c.sendToDLQ(msg, events.SubjectColumnDeleted, consumerColumnDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectColumnDeleted, consumerColumnDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			columnName := c.nameCache.GetColumnName(ctx, event.ColumnID)
			if columnName == "" {
				columnName = "колонка"
			}

			title := c.buildTitle(ctx, fmt.Sprintf("Колонка \"%s\" удалена", columnName), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeColumnDeleted,
				title, "",
				map[string]string{"board_id": event.BoardID, "column_id": event.ColumnID, "column_title": columnName, "board_title": boardName})
			_ = c.nameCache.DeleteColumnName(ctx, event.ColumnID)
			return nil
		})
	}),
		nats.Durable(consumerColumnDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectColumnDeleted, err)
	}
	log.Printf("consumer started: %s", consumerColumnDeleted)
	return nil
}
