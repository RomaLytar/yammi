package usecase

import (
	"context"
	"log/slog"
	"time"

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

	// 3. Публикуем события (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishBoardCreated(ctx, BoardCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.CreatedAt,
			BoardID:      board.ID,
			OwnerID:      board.OwnerID,
			Title:        board.Title,
			Description:  board.Description,
		}); err != nil {
			slog.Error("failed to publish BoardCreated", "error", err, "board_id", board.ID)
		}
		// Публикуем MemberAdded для owner — notification cache должен узнать об участнике
		if err := uc.publisher.PublishMemberAdded(ctx, MemberAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.CreatedAt,
			BoardID:      board.ID,
			UserID:       board.OwnerID,
			ActorID:      board.OwnerID,
			Role:         string(domain.RoleOwner),
			BoardTitle:   board.Title,
		}); err != nil {
			slog.Error("failed to publish MemberAdded", "error", err, "board_id", board.ID)
		}
	}()

	return board, nil
}
