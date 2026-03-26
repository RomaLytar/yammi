package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AssignCardUseCase struct {
	cardRepo     CardRepository
	boardRepo    BoardRepository
	memberRepo   MembershipRepository
	activityRepo ActivityRepository
	publisher    EventPublisher
}

func NewAssignCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, activityRepo ActivityRepository, publisher EventPublisher) *AssignCardUseCase {
	return &AssignCardUseCase{
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
		memberRepo:   memberRepo,
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *AssignCardUseCase) Execute(ctx context.Context, cardID, boardID, userID, assigneeID string) (*domain.Card, error) {
	// 1. Проверка доступа актора
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверка что assignee — участник доски
	isAssigneeMember, _, err := uc.memberRepo.IsMember(ctx, boardID, assigneeID)
	if err != nil {
		return nil, err
	}
	if !isAssigneeMember {
		return nil, domain.ErrAssigneeNotMember
	}

	// 3. Загружаем карточку
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return nil, err
	}

	// 4. Запоминаем предыдущего assignee
	prevAssignee := card.AssigneeID

	// 5. Обновляем assignee
	card.AssigneeID = &assigneeID
	card.UpdatedAt = getCurrentTime()

	// 6. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 7. Записываем активность (async, non-blocking)
	if activity, actErr := domain.NewActivity(card.ID, boardID, userID, domain.ActivityCardAssigned,
		fmt.Sprintf("Карточка \"%s\" назначена", card.Title),
		map[string]string{"assignee_id": assigneeID}); actErr == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := uc.activityRepo.Create(ctx, activity); err != nil {
				slog.Error("failed to write activity log", "error", err, "card_id", card.ID)
			}
		}()
	}

	// 8. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		if err := uc.publisher.PublishCardAssigned(ctx, CardAssigned{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.UpdatedAt,
			CardID:       card.ID,
			BoardID:      boardID,
			ColumnID:     card.ColumnID,
			ActorID:      userID,
			AssigneeID:   assigneeID,
			PrevAssignee: prevAssignee,
			CardTitle:    card.Title,
		}); err != nil {
			slog.Error("failed to publish CardAssigned", "error", err, "card_id", card.ID, "board_id", boardID)
		}
	}()

	return card, nil
}
