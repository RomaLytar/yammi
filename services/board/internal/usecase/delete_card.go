package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteCardUseCase struct {
	cardRepo   CardRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteCardUseCase {
	return &DeleteCardUseCase{
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteCardUseCase) Execute(ctx context.Context, cardIDs []string, boardID, userID string) error {
	// 1. Проверка доступа
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Если не owner, проверяем что каждая карточка принадлежит пользователю
	if role != domain.RoleOwner {
		for _, cardID := range cardIDs {
			card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
			if err != nil {
				return err
			}
			if card.CreatorID != userID {
				return domain.ErrAccessDenied
			}
		}
	}

	// 3. Batch delete
	if err := uc.cardRepo.BatchDelete(ctx, boardID, cardIDs); err != nil {
		return err
	}

	// 4. Обновляем updated_at доски + публикуем события (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		for _, cardID := range cardIDs {
			if err := uc.publisher.PublishCardDeleted(ctx, CardDeleted{
				EventID:      generateEventID(),
				EventVersion: 1,
				OccurredAt:   getCurrentTime(),
				BoardID:      boardID,
				CardID:       cardID,
				ActorID:      userID,
			}); err != nil {
				slog.Error("failed to publish CardDeleted", "error", err, "card_id", cardID, "board_id", boardID)
			}
		}
	}()

	return nil
}
