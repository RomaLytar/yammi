package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteChecklistUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
	publisher     EventPublisher
}

func NewDeleteChecklistUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteChecklistUseCase {
	return &DeleteChecklistUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
		publisher:     publisher,
	}
}

func (uc *DeleteChecklistUseCase) Execute(ctx context.Context, checklistID, boardID, userID string) error {
	// 1. Проверка доступа (любой member может удалять чеклисты)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}

	// 2. Удаляем чеклист (CASCADE удалит items)
	if err := uc.checklistRepo.DeleteChecklist(ctx, checklistID, boardID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishChecklistDeleted(context.Background(), ChecklistDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ChecklistID:  checklistID,
			BoardID:      boardID,
			ActorID:      userID,
		})
	}()

	return nil
}
