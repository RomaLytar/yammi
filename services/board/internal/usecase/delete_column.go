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
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Удаляем (CASCADE удалит cards)
	if err := uc.columnRepo.Delete(ctx, columnID); err != nil {
		return err
	}

	// 3. Обновляем updated_at доски
	_ = uc.boardRepo.TouchUpdatedAt(ctx, boardID)

	// 4. Публикуем событие
	go func() {
		_ = uc.publisher.PublishColumnDeleted(context.Background(), ColumnDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ColumnID:     columnID,
		})
	}()

	return nil
}
