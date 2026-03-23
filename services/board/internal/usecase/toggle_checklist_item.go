package usecase

import (
	"context"
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

func (uc *ToggleChecklistItemUseCase) Execute(ctx context.Context, itemID, boardID, userID string, isChecked bool) error {
	// 1. Проверка доступа (member может переключать элементы)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Переключаем элемент
	if err := uc.checklistRepo.ToggleItem(ctx, itemID, boardID, isChecked); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishChecklistItemToggled(context.Background(), ChecklistItemToggled{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ItemID:       itemID,
			BoardID:      boardID,
			ActorID:      userID,
			IsChecked:    isChecked,
		})
	}()

	return nil
}
