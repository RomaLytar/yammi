package usecase

import (
	"context"

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
		_ = uc.boardRepo.TouchUpdatedAt(context.Background(), boardID)
		_ = uc.publisher.PublishColumnDeleted(context.Background(), ColumnDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ColumnID:     columnID,
			ActorID:      userID,
		})
	}()

	return nil
}
