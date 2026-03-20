package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateBoardUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateBoardUseCase(boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateBoardUseCase {
	return &UpdateBoardUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *UpdateBoardUseCase) Execute(ctx context.Context, boardID, userID, title, description string, version int) (*domain.Board, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем доску
	board, err := uc.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем
	if err := board.Update(title, description); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.boardRepo.Update(ctx, board); err != nil {
		return nil, err
	}

	// 5. Публикуем событие
	go func() {
		_ = uc.publisher.PublishBoardUpdated(context.Background(), BoardUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.UpdatedAt,
			BoardID:      board.ID,
			UserID:       userID,
			Title:        board.Title,
			Description:  board.Description,
		})
	}()

	return board, nil
}
