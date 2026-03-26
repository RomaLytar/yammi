package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateCardUseCase struct {
	cardRepo     CardRepository
	boardRepo    BoardRepository
	memberRepo   MembershipRepository
	activityRepo ActivityRepository
	publisher    EventPublisher
}

func NewCreateCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, activityRepo ActivityRepository, publisher EventPublisher) *CreateCardUseCase {
	return &CreateCardUseCase{
		cardRepo:     cardRepo,
		boardRepo:    boardRepo,
		memberRepo:   memberRepo,
		activityRepo: activityRepo,
		publisher:    publisher,
	}
}

func (uc *CreateCardUseCase) Execute(ctx context.Context, columnID, boardID, userID, title, description, position string, assigneeID *string, dueDate *time.Time, priority domain.Priority, taskType domain.TaskType) (*domain.Card, error) {
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

	// 3. Если position пустой — генерируем (в конец колонки)
	if position == "" {
		lastCard, err := uc.cardRepo.GetLastInColumn(ctx, columnID)
		if err != nil && err != domain.ErrCardNotFound {
			return nil, err
		}
		if lastCard != nil {
			position, _ = domain.LexorankBetween(lastCard.Position, "")
		} else {
			position = domain.LexorankFirst // первая карточка
		}
	}

	// 4. Создаем карточку (валидация lexorank внутри)
	card, err := domain.NewCard(columnID, title, description, position, assigneeID, userID, dueDate, priority, taskType)
	if err != nil {
		return nil, err
	}

	// 5. Сохраняем
	if err := uc.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	// 6. Записываем активность (async, non-blocking)
	if activity, err := domain.NewActivity(card.ID, boardID, userID, domain.ActivityCardCreated,
		fmt.Sprintf("Карточка \"%s\" создана", card.Title), nil); err == nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := uc.activityRepo.Create(ctx, activity); err != nil {
				slog.Error("failed to write activity log", "error", err, "card_id", card.ID)
			}
		}()
	}

	// 7. Обновляем updated_at доски + публикуем события (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		if err := uc.publisher.PublishCardCreated(ctx, CardCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.CreatedAt,
			CardID:       card.ID,
			ColumnID:     card.ColumnID,
			BoardID:      boardID,
			ActorID:      userID,
			Title:        card.Title,
			Description:  card.Description,
			Position:     card.Position,
			AssigneeID:   card.AssigneeID,
			DueDate:      card.DueDate,
			Priority:     string(card.Priority),
			TaskType:     string(card.TaskType),
		}); err != nil {
			slog.Error("failed to publish CardCreated", "error", err, "card_id", card.ID, "board_id", boardID)
		}

		// Если карточка создана сразу с assignee — отправляем событие
		if card.AssigneeID != nil && *card.AssigneeID != "" {
			if err := uc.publisher.PublishCardAssigned(ctx, CardAssigned{
				EventID:      generateEventID(),
				EventVersion: 1,
				OccurredAt:   card.CreatedAt,
				CardID:       card.ID,
				BoardID:      boardID,
				ColumnID:     card.ColumnID,
				ActorID:      userID,
				AssigneeID:   *card.AssigneeID,
				CardTitle:    card.Title,
			}); err != nil {
				slog.Error("failed to publish CardAssigned", "error", err, "card_id", card.ID, "board_id", boardID)
			}
		}
	}()

	return card, nil
}
