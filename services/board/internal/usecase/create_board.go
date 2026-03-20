package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateBoardUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewCreateBoardUseCase(boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateBoardUseCase {
	return &CreateBoardUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *CreateBoardUseCase) Execute(ctx context.Context, title, description, ownerID string) (*domain.Board, error) {
	// 1. Создаем доменную сущность (валидация внутри)
	board, err := domain.NewBoard(title, description, ownerID)
	if err != nil {
		return nil, err
	}

	// 2. Сохраняем (автоматически создает owner в board_members)
	if err := uc.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishBoardCreated(context.Background(), BoardCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.CreatedAt,
			BoardID:      board.ID,
			OwnerID:      board.OwnerID,
			Title:        board.Title,
			Description:  board.Description,
		})
	}()

	return board, nil
}
