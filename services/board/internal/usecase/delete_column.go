package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteColumnUseCase struct {
	columnRepo ColumnRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteColumnUseCase(columnRepo ColumnRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteColumnUseCase {
	return &DeleteColumnUseCase{
		columnRepo: columnRepo,
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

	// 3. Публикуем событие
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
