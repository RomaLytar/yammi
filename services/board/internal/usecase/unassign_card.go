package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UnassignCardUseCase struct {
	cardRepo     CardRepository
	boardRepo    BoardRepository
	memberRepo   MembershipRepository
	activityRepo ActivityRepository
	publisher    EventPublisher
}

func NewUnassignCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, activityRepo ActivityRepository, publisher EventPublisher) *UnassignCardUseCase {
	return &UnassignCardUseCase{
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
		memberRepo:   memberRepo,
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *UnassignCardUseCase) Execute(ctx context.Context, cardID, boardID, userID string) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточку
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Если карточка не назначена — ничего делать не нужно
	if card.AssigneeID == nil {
		return card, nil
	}

	// 4. Запоминаем предыдущего assignee и снимаем
	prevAssignee := *card.AssigneeID
	card.AssigneeID = nil
	card.UpdatedAt = getCurrentTime()

	// 5. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 6. Записываем активность
	activity, actErr := domain.NewActivity(card.ID, boardID, userID, domain.ActivityCardUnassigned,
		fmt.Sprintf("Назначение снято с карточки \"%s\"", card.Title),
		map[string]string{"prev_assignee": prevAssignee})
	if actErr == nil {
		if writeErr := uc.activityRepo.Create(ctx, activity); writeErr != nil {
			log.Printf("failed to write activity log: %v", writeErr)
		}
	}

	// 7. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.boardRepo.TouchUpdatedAt(context.Background(), boardID)
		_ = uc.publisher.PublishCardUnassigned(context.Background(), CardUnassigned{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.UpdatedAt,
			CardID:       card.ID,
			BoardID:      boardID,
			ColumnID:     card.ColumnID,
			ActorID:      userID,
			PrevAssignee: prevAssignee,
			CardTitle:    card.Title,
		})
	}()

	return card, nil
}
