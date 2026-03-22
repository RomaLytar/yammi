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

// --- Card events ---

// buildTitle формирует title: "Действие → доска" (автор в metadata, фронт показывает отдельно)
func (c *Consumer) buildTitle(ctx context.Context, action, boardID string) string {
	boardName := c.nameCache.GetBoardName(ctx, boardID)
	if boardName != "" {
		return fmt.Sprintf("%s → %s", action, boardName)
	}
	return action
}

func (c *Consumer) subscribeCardCreated() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardCreated, "notification-workers", func(msg *nats.Msg) {
		var event events.CardCreated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardCreated, err)
			c.sendToDLQ(msg, events.SubjectCardCreated, consumerCardCreated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardCreated, consumerCardCreated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetCardName(ctx, event.CardID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)

			title := c.buildTitle(ctx, fmt.Sprintf("Карточка \"%s\" создана", event.Title), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardCreated,
				title, "",
				map[string]string{
					"board_id":    event.BoardID,
					"board_title": boardName,
					"card_id":     event.CardID,
					"card_title":  event.Title,
					"column_id":   event.ColumnID,
				})
			return nil
		})
	},
		nats.Durable(consumerCardCreated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardCreated, err)
	}
	log.Printf("consumer started: %s", consumerCardCreated)
	return nil
}

func (c *Consumer) subscribeCardUpdated() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardUpdated, "notification-workers", func(msg *nats.Msg) {
		var event events.CardUpdated
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardUpdated, err)
			c.sendToDLQ(msg, events.SubjectCardUpdated, consumerCardUpdated, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardUpdated, consumerCardUpdated, func() error {
			ctx := context.Background()
			_ = c.nameCache.SetCardName(ctx, event.CardID, event.Title)
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)

			title := c.buildTitle(ctx, fmt.Sprintf("Карточка \"%s\" обновлена", event.Title), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardUpdated,
				title, "",
				map[string]string{
					"board_id":    event.BoardID,
					"board_title": boardName,
					"card_id":     event.CardID,
					"card_title":  event.Title,
					"column_id":   event.ColumnID,
				})
			return nil
		})
	},
		nats.Durable(consumerCardUpdated),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardUpdated, err)
	}
	log.Printf("consumer started: %s", consumerCardUpdated)
	return nil
}

func (c *Consumer) subscribeCardMoved() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardMoved, "notification-workers", func(msg *nats.Msg) {
		var event events.CardMoved
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardMoved, err)
			c.sendToDLQ(msg, events.SubjectCardMoved, consumerCardMoved, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardMoved, consumerCardMoved, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			if cardName == "" {
				cardName = "карточка"
			}

			title := c.buildTitle(ctx, fmt.Sprintf("Карточка \"%s\" перемещена", cardName), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardMoved,
				title, "",
				map[string]string{
					"board_id":       event.BoardID,
					"board_title":    boardName,
					"card_id":        event.CardID,
					"card_title":     cardName,
					"from_column_id": event.FromColumnID,
					"to_column_id":   event.ToColumnID,
				})
			return nil
		})
	},
		nats.Durable(consumerCardMoved),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardMoved, err)
	}
	log.Printf("consumer started: %s", consumerCardMoved)
	return nil
}

func (c *Consumer) subscribeCardDeleted() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardDeleted, "notification-workers", func(msg *nats.Msg) {
		var event events.CardDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardDeleted, err)
			c.sendToDLQ(msg, events.SubjectCardDeleted, consumerCardDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardDeleted, consumerCardDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			if cardName == "" {
				cardName = "карточка"
			}

			title := c.buildTitle(ctx, fmt.Sprintf("Карточка \"%s\" удалена", cardName), event.BoardID)
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeCardDeleted,
				title, "",
				map[string]string{
					"board_id":    event.BoardID,
					"board_title": boardName,
					"card_id":     event.CardID,
					"card_title":  cardName,
				})
			_ = c.nameCache.DeleteCardName(ctx, event.CardID)
			return nil
		})
	},
		nats.Durable(consumerCardDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardDeleted, err)
	}
	log.Printf("consumer started: %s", consumerCardDeleted)
	return nil
}
