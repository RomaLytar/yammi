package usecase

import (
	"context"
	"log/slog"
	"time"

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

// Execute удаляет одну или несколько досок (batch). Только owner может удалить.
func (uc *DeleteBoardUseCase) Execute(ctx context.Context, boardIDs []string, userID string) error {
	// 1. Проверяем ownership для каждой доски
	for _, boardID := range boardIDs {
		isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
		if err != nil {
			return err
		}
		if !isMember || role != domain.RoleOwner {
			return domain.ErrAccessDenied
		}
	}

	// 2. Batch delete в одной транзакции
	if err := uc.boardRepo.BatchDelete(ctx, boardIDs); err != nil {
		return err
	}

	// 3. Публикуем события
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for _, bid := range boardIDs {
			if err := uc.publisher.PublishBoardDeleted(ctx, BoardDeleted{
				EventID:      generateEventID(),
				EventVersion: 1,
				OccurredAt:   getCurrentTime(),
				BoardID:      bid,
				ActorID:      userID,
			}); err != nil {
				slog.Error("failed to publish BoardDeleted", "error", err, "board_id", bid)
			}
		}
	}()

	return nil
}
