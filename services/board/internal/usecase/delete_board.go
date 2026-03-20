package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteBoardUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteBoardUseCase(boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteBoardUseCase {
	return &DeleteBoardUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteBoardUseCase) Execute(ctx context.Context, boardID, userID string) error {
	// 1. Проверка: только owner может удалить доску
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember || role != domain.RoleOwner {
		return domain.ErrAccessDenied
	}

	// 2. Удаляем (CASCADE удалит columns, cards, members)
	if err := uc.boardRepo.Delete(ctx, boardID); err != nil {
		return err
	}

	// 3. Публикуем событие
	go func() {
		_ = uc.publisher.PublishBoardDeleted(context.Background(), BoardDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			UserID:       userID,
		})
	}()

	return nil
}
