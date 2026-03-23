package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateChecklistUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
	publisher     EventPublisher
}

func NewUpdateChecklistUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateChecklistUseCase {
	return &UpdateChecklistUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
		publisher:     publisher,
	}
}

func (uc *UpdateChecklistUseCase) Execute(ctx context.Context, checklistID, boardID, userID, title string) (*domain.Checklist, error) {
	// 1. Проверка доступа (member может обновлять чеклисты)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем чеклист
	checklist, err := uc.checklistRepo.GetChecklistByID(ctx, checklistID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем (валидация внутри)
	if err := checklist.Update(title); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.checklistRepo.UpdateChecklist(ctx, checklist); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishChecklistUpdated(context.Background(), ChecklistUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ChecklistID:  checklist.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Title:        checklist.Title,
		})
	}()

	return checklist, nil
}
