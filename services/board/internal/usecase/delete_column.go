package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteColumnUseCase struct {
	columnRepo ColumnRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteColumnUseCase(columnRepo ColumnRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteColumnUseCase {
	return &DeleteColumnUseCase{
		columnRepo: columnRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteColumnUseCase) Execute(ctx context.Context, columnID, boardID, userID string) error {
	// 1. Проверка доступа (только owner может удалять колонки)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return domain.ErrNotOwner
	}

	// 2. Удаляем карточки и колонку
	if err := uc.columnRepo.Delete(ctx, columnID, boardID); err != nil {
		return err
	}

	// 3. Обновляем updated_at доски + публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		if err := uc.publisher.PublishColumnDeleted(ctx, ColumnDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ColumnID:     columnID,
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish ColumnDeleted", "error", err, "board_id", boardID)
		}
	}()

	return nil
}
