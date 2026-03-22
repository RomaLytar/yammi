package usecase

import "context"

// EventPublisher публикует события в NATS
type EventPublisher interface {
	PublishCommentCreated(ctx context.Context, event CommentCreated) error
	PublishCommentUpdated(ctx context.Context, event CommentUpdated) error
	PublishCommentDeleted(ctx context.Context, event CommentDeleted) error
}
