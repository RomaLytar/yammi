package nats

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/usecase"
)

// EventPublisher реализует интерфейс usecase.EventPublisher
type EventPublisher struct {
	publisher *Publisher
}

func NewEventPublisher(publisher *Publisher) *EventPublisher {
	return &EventPublisher{publisher: publisher}
}

func (p *EventPublisher) PublishBoardCreated(ctx context.Context, event usecase.BoardCreated) error {
	return p.publisher.Publish(ctx, "board.created", event)
}

func (p *EventPublisher) PublishBoardUpdated(ctx context.Context, event usecase.BoardUpdated) error {
	return p.publisher.Publish(ctx, "board.updated", event)
}

func (p *EventPublisher) PublishBoardDeleted(ctx context.Context, event usecase.BoardDeleted) error {
	return p.publisher.Publish(ctx, "board.deleted", event)
}

func (p *EventPublisher) PublishColumnCreated(ctx context.Context, event usecase.ColumnAdded) error {
	return p.publisher.Publish(ctx, "column.created", event)
}

func (p *EventPublisher) PublishColumnUpdated(ctx context.Context, event usecase.ColumnUpdated) error {
	return p.publisher.Publish(ctx, "column.updated", event)
}

func (p *EventPublisher) PublishColumnDeleted(ctx context.Context, event usecase.ColumnDeleted) error {
	return p.publisher.Publish(ctx, "column.deleted", event)
}

func (p *EventPublisher) PublishColumnsReordered(ctx context.Context, event usecase.ColumnsReordered) error {
	return p.publisher.Publish(ctx, "columns.reordered", event)
}

func (p *EventPublisher) PublishCardCreated(ctx context.Context, event usecase.CardCreated) error {
	return p.publisher.Publish(ctx, "card.created", event)
}

func (p *EventPublisher) PublishCardUpdated(ctx context.Context, event usecase.CardUpdated) error {
	return p.publisher.Publish(ctx, "card.updated", event)
}

func (p *EventPublisher) PublishCardMoved(ctx context.Context, event usecase.CardMoved) error {
	return p.publisher.Publish(ctx, "card.moved", event)
}

func (p *EventPublisher) PublishCardDeleted(ctx context.Context, event usecase.CardDeleted) error {
	return p.publisher.Publish(ctx, "card.deleted", event)
}

func (p *EventPublisher) PublishMemberAdded(ctx context.Context, event usecase.MemberAdded) error {
	return p.publisher.Publish(ctx, "member.added", event)
}

func (p *EventPublisher) PublishMemberRemoved(ctx context.Context, event usecase.MemberRemoved) error {
	return p.publisher.Publish(ctx, "member.removed", event)
}
