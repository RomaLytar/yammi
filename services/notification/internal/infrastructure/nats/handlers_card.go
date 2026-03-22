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
			// Обновляем кеш имени карточки (для других нотификаций)
			_ = c.nameCache.SetCardName(ctx, event.CardID, event.Title)
			// Не создаём нотификацию "обновлена" — это шум.
			// Значимые изменения (assign, move, delete) имеют свои события.
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

func (c *Consumer) subscribeCardAssigned() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardAssigned, "notification-workers", func(msg *nats.Msg) {
		var event events.CardAssigned
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardAssigned, err)
			c.sendToDLQ(msg, events.SubjectCardAssigned, consumerCardAssigned, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardAssigned, consumerCardAssigned, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardTitle := event.CardTitle
			if cardTitle == "" {
				cardTitle = c.nameCache.GetCardName(ctx, event.CardID)
			}
			if cardTitle == "" {
				cardTitle = "карточка"
			}

			// Только персональное уведомление назначенному (не спамим всю доску)
			if event.AssigneeID != event.ActorID {
				actorName := c.nameCache.GetUserName(ctx, event.ActorID)
				if actorName == "" {
					actorName = "Участник"
				}
				assigneeTitle := fmt.Sprintf("Вам назначили задачу \"%s\"", cardTitle)
				if boardName != "" {
					assigneeTitle = fmt.Sprintf("%s в доске \"%s\"", assigneeTitle, boardName)
				}
				_ = c.createNotification(ctx, event.AssigneeID, domain.TypeCardAssigned,
					assigneeTitle, "",
					map[string]string{
						"board_id":    event.BoardID,
						"board_title": boardName,
						"card_id":     event.CardID,
						"card_title":  cardTitle,
						"actor_id":    event.ActorID,
						"actor_name":  actorName,
					})
			}
			return nil
		})
	},
		nats.Durable(consumerCardAssigned),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardAssigned, err)
	}
	log.Printf("consumer started: %s", consumerCardAssigned)
	return nil
}

func (c *Consumer) subscribeCardUnassigned() error {
	_, err := c.js.QueueSubscribe(events.SubjectCardUnassigned, "notification-workers", func(msg *nats.Msg) {
		var event events.CardUnassigned
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectCardUnassigned, err)
			c.sendToDLQ(msg, events.SubjectCardUnassigned, consumerCardUnassigned, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectCardUnassigned, consumerCardUnassigned, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardTitle := event.CardTitle
			if cardTitle == "" {
				cardTitle = c.nameCache.GetCardName(ctx, event.CardID)
			}
			if cardTitle == "" {
				cardTitle = "карточка"
			}

			// Только персональное уведомление тому, с кого сняли (не спамим доску)
			if event.PrevAssignee != event.ActorID {
				prevTitle := fmt.Sprintf("С вас сняли задачу \"%s\"", cardTitle)
				if boardName != "" {
					prevTitle = fmt.Sprintf("%s в доске \"%s\"", prevTitle, boardName)
				}
				_ = c.createNotification(ctx, event.PrevAssignee, domain.TypeCardUnassigned,
					prevTitle, "",
					map[string]string{
						"board_id":    event.BoardID,
						"board_title": boardName,
						"card_id":     event.CardID,
						"card_title":  cardTitle,
					})
			}
			return nil
		})
	},
		nats.Durable(consumerCardUnassigned),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectCardUnassigned, err)
	}
	log.Printf("consumer started: %s", consumerCardUnassigned)
	return nil
}

func (c *Consumer) subscribeAttachmentUploaded() error {
	_, err := c.js.QueueSubscribe(events.SubjectAttachmentUploaded, "notification-workers", func(msg *nats.Msg) {
		var event events.AttachmentUploaded
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectAttachmentUploaded, err)
			c.sendToDLQ(msg, events.SubjectAttachmentUploaded, consumerAttachmentUploaded, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectAttachmentUploaded, consumerAttachmentUploaded, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			if cardName == "" {
				cardName = "карточка"
			}

			title := fmt.Sprintf("Файл \"%s\" прикреплён к \"%s\"", event.FileName, cardName)
			if boardName != "" {
				title = fmt.Sprintf("%s в доске \"%s\"", title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeAttachmentUploaded,
				title, "",
				map[string]string{
					"board_id":      event.BoardID,
					"board_title":   boardName,
					"card_id":       event.CardID,
					"card_title":    cardName,
					"attachment_id": event.AttachmentID,
					"file_name":     event.FileName,
				})
			return nil
		})
	},
		nats.Durable(consumerAttachmentUploaded),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectAttachmentUploaded, err)
	}
	log.Printf("consumer started: %s", consumerAttachmentUploaded)
	return nil
}

func (c *Consumer) subscribeAttachmentDeleted() error {
	_, err := c.js.QueueSubscribe(events.SubjectAttachmentDeleted, "notification-workers", func(msg *nats.Msg) {
		var event events.AttachmentDeleted
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("poison message on %s, sending to DLQ: %v", events.SubjectAttachmentDeleted, err)
			c.sendToDLQ(msg, events.SubjectAttachmentDeleted, consumerAttachmentDeleted, err.Error())
			return
		}

		c.handleWithRetry(msg, events.SubjectAttachmentDeleted, consumerAttachmentDeleted, func() error {
			ctx := context.Background()
			boardName := c.nameCache.GetBoardName(ctx, event.BoardID)
			cardName := c.nameCache.GetCardName(ctx, event.CardID)
			if cardName == "" {
				cardName = "карточка"
			}

			title := fmt.Sprintf("Файл \"%s\" удалён из \"%s\"", event.FileName, cardName)
			if boardName != "" {
				title = fmt.Sprintf("%s в доске \"%s\"", title, boardName)
			}
			c.notifyBoardMembers(ctx, event.BoardID, event.ActorID, domain.TypeAttachmentDeleted,
				title, "",
				map[string]string{
					"board_id":      event.BoardID,
					"board_title":   boardName,
					"card_id":       event.CardID,
					"card_title":    cardName,
					"attachment_id": event.AttachmentID,
					"file_name":     event.FileName,
				})
			return nil
		})
	},
		nats.Durable(consumerAttachmentDeleted),
		nats.ManualAck(),
		nats.DeliverNew(),
		nats.MaxDeliver(maxDeliveries),
		nats.MaxAckPending(maxAckPending),
		nats.AckWait(ackWait),
	)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", events.SubjectAttachmentDeleted, err)
	}
	log.Printf("consumer started: %s", consumerAttachmentDeleted)
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
