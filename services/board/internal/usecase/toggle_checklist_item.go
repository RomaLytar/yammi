package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ToggleChecklistItemUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
	publisher     EventPublisher
}

func NewToggleChecklistItemUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository, publisher EventPublisher) *ToggleChecklistItemUseCase {
	return &ToggleChecklistItemUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
		publisher:     publisher,
	}
}

func (uc *ToggleChecklistItemUseCase) Execute(ctx context.Context, itemID, boardID, userID string) (bool, error) {
	// 1. Проверка доступа (member может переключать элементы)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return false, err
	}
	if !isMember {
		return false, domain.ErrAccessDenied
	}

	// 2. Атомарно переключаем (UPDATE ... SET is_checked = NOT is_checked RETURNING — один запрос вместо двух)
	newChecked, err := uc.checklistRepo.ToggleItemAtomic(ctx, itemID, boardID)
	if err != nil {
		return false, err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishChecklistItemToggled(ctx, ChecklistItemToggled{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ItemID:       itemID,
			BoardID:      boardID,
			ActorID:      userID,
			IsChecked:    newChecked,
		}); err != nil {
			slog.Error("failed to publish ChecklistItemToggled", "error", err, "item_id", itemID, "board_id", boardID)
		}
	}()

	return newChecked, nil
}
