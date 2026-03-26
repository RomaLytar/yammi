package usecase

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type MoveCardUseCase struct {
	cardRepo     CardRepository
	boardRepo    BoardRepository
	memberRepo   MembershipRepository
	activityRepo ActivityRepository
	publisher    EventPublisher
}

func NewMoveCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, activityRepo ActivityRepository, publisher EventPublisher) *MoveCardUseCase {
	return &MoveCardUseCase{
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
		memberRepo:   memberRepo,
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *MoveCardUseCase) Execute(ctx context.Context, cardID, boardID, fromColumnID, toColumnID, userID, newPosition string) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточку (фильтр по boardID — IDOR protection)
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Валидируем lexorank позицию
	if err := domain.ValidateLexorank(newPosition); err != nil {
		return nil, err
	}

	// 4. Перемещаем карточку (domain logic)
	if err := card.Move(toColumnID, newPosition); err != nil {
		return nil, err
	}

	// 5. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 6. Записываем активность (синхронно)
	changes := map[string]string{
		"from_column_id": fromColumnID,
		"to_column_id":   toColumnID,
	}
	activity, actErr := domain.NewActivity(card.ID, boardID, userID, domain.ActivityCardMoved,
		"Карточка перемещена", changes)
	if actErr == nil {
		if writeErr := uc.activityRepo.Create(ctx, activity); writeErr != nil {
			log.Printf("failed to write activity log: %v", writeErr)
		}
	}

	// 7. Обновляем updated_at доски + публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		if err := uc.publisher.PublishCardMoved(ctx, CardMoved{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.UpdatedAt,
			CardID:       cardID,
			BoardID:      boardID,
			ActorID:      userID,
			FromColumnID: fromColumnID,
			ToColumnID:   toColumnID,
			NewPosition:  newPosition,
		}); err != nil {
			slog.Error("failed to publish CardMoved", "error", err, "card_id", cardID, "board_id", boardID)
		}
	}()

	return card, nil
}
