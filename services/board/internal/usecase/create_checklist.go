package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateChecklistUseCase struct {
	checklistRepo ChecklistRepository
	memberRepo    MembershipRepository
	publisher     EventPublisher
}

func NewCreateChecklistUseCase(checklistRepo ChecklistRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateChecklistUseCase {
	return &CreateChecklistUseCase{
		checklistRepo: checklistRepo,
		memberRepo:    memberRepo,
		publisher:     publisher,
	}
}

func (uc *CreateChecklistUseCase) Execute(ctx context.Context, cardID, boardID, userID, title string, position int) (*domain.Checklist, error) {
	// 1. Проверка доступа (member может создавать чеклисты)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Создаем чеклист (валидация внутри)
	checklist, err := domain.NewChecklist("", cardID, boardID, title, position)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем
	if err := uc.checklistRepo.CreateChecklist(ctx, checklist); err != nil {
		return nil, err
	}

	// 4. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishChecklistCreated(ctx, ChecklistCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ChecklistID:  checklist.ID,
			CardID:       cardID,
			BoardID:      boardID,
			ActorID:      userID,
			Title:        checklist.Title,
		}); err != nil {
			slog.Error("failed to publish ChecklistCreated", "error", err, "checklist_id", checklist.ID, "board_id", boardID)
		}
	}()

	return checklist, nil
}
