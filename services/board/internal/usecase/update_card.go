package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateCardUseCase struct {
	cardRepo     CardRepository
	boardRepo    BoardRepository
	memberRepo   MembershipRepository
	activityRepo ActivityRepository
	publisher    EventPublisher
}

func NewUpdateCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, activityRepo ActivityRepository, publisher EventPublisher) *UpdateCardUseCase {
	return &UpdateCardUseCase{
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
		memberRepo:   memberRepo,
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *UpdateCardUseCase) Execute(ctx context.Context, cardID, boardID, userID, title, description string, assigneeID *string, version int) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Валидация assignee — должен быть участником доски
	if assigneeID != nil && *assigneeID != "" {
		isAssigneeMember, _, err := uc.memberRepo.IsMember(ctx, boardID, *assigneeID)
		if err != nil {
			return nil, err
		}
		if !isAssigneeMember {
			return nil, domain.ErrAssigneeNotMember
		}
	}

	// 3. Загружаем карточку (фильтр по boardID — IDOR protection)
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return nil, err
	}

	// 4. Запоминаем предыдущие значения для детекта изменений
	prevAssignee := card.AssigneeID
	prevTitle := card.Title
	prevDescription := card.Description

	// 5. Обновляем
	if err := card.Update(title, description, assigneeID); err != nil {
		return nil, err
	}

	// 6. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 7. Записываем активность — только если изменился title или description
	// Assignee changes записываются отдельно через assign/unassign events
	if prevTitle != card.Title || prevDescription != card.Description {
		changes := map[string]string{}
		if prevTitle != card.Title {
			changes["old_title"] = prevTitle
			changes["new_title"] = card.Title
		}
		if prevDescription != card.Description {
			changes["description_changed"] = "true"
		}
		activity, actErr := domain.NewActivity(card.ID, boardID, userID, domain.ActivityCardUpdated,
			fmt.Sprintf("Карточка \"%s\" обновлена", card.Title), changes)
		if actErr == nil {
			if writeErr := uc.activityRepo.Create(ctx, activity); writeErr != nil {
				log.Printf("failed to write activity log: %v", writeErr)
			}
		}
	}

	// 8. Обновляем updated_at доски + публикуем события (async, non-blocking)
	go func() {
		_ = uc.boardRepo.TouchUpdatedAt(context.Background(), boardID)
		_ = uc.publisher.PublishCardUpdated(context.Background(), CardUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.UpdatedAt,
			CardID:       card.ID,
			ColumnID:     card.ColumnID,
			BoardID:      boardID,
			ActorID:      userID,
			Title:        card.Title,
			Description:  card.Description,
			AssigneeID:   card.AssigneeID,
		})

		// Детектим изменение assignee и публикуем отдельное событие
		assigneeChanged := !assigneesEqual(prevAssignee, card.AssigneeID)
		if assigneeChanged {
			if card.AssigneeID != nil && *card.AssigneeID != "" {
				_ = uc.publisher.PublishCardAssigned(context.Background(), CardAssigned{
					EventID:      generateEventID(),
					EventVersion: 1,
					OccurredAt:   card.UpdatedAt,
					CardID:       card.ID,
					BoardID:      boardID,
					ColumnID:     card.ColumnID,
					ActorID:      userID,
					AssigneeID:   *card.AssigneeID,
					PrevAssignee: prevAssignee,
					CardTitle:    card.Title,
				})
			} else if prevAssignee != nil {
				_ = uc.publisher.PublishCardUnassigned(context.Background(), CardUnassigned{
					EventID:      generateEventID(),
					EventVersion: 1,
					OccurredAt:   card.UpdatedAt,
					CardID:       card.ID,
					BoardID:      boardID,
					ColumnID:     card.ColumnID,
					ActorID:      userID,
					PrevAssignee: *prevAssignee,
					CardTitle:    card.Title,
				})
			}
		}
	}()

	return card, nil
}

// assigneesEqual сравнивает два *string assignee
func assigneesEqual(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
