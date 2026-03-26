package usecase

import (
	"context"
	"log/slog"
	"time"

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
	// 1. Проверка доступа (только owner может обновлять доску)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return nil, domain.ErrNotOwner
	}

	// 2. Загружаем доску
	board, err := uc.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Проверка optimistic locking — клиент обязан передать актуальную версию
	if version == 0 {
		return nil, domain.ErrInvalidVersion
	}
	if board.Version != version {
		return nil, domain.ErrInvalidVersion
	}

	// 4. Обновляем
	if err := board.Update(title, description); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.boardRepo.Update(ctx, board); err != nil {
		return nil, err
	}

	// 5. Публикуем событие
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishBoardUpdated(ctx, BoardUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.UpdatedAt,
			BoardID:      board.ID,
			ActorID:      userID,
			Title:        board.Title,
			Description:  board.Description,
		}); err != nil {
			slog.Error("failed to publish BoardUpdated", "error", err, "board_id", board.ID)
		}
	}()

	return board, nil
}
