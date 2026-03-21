package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	
)

type CreateCardUseCase struct {
	cardRepo   CardRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewCreateCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateCardUseCase {
	return &CreateCardUseCase{
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *CreateCardUseCase) Execute(ctx context.Context, columnID, boardID, userID, title, description, position string, assigneeID *string) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Если position пустой — генерируем (в конец колонки)
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

	// 3. Создаем карточку (валидация lexorank внутри)
	card, err := domain.NewCard(columnID, title, description, position, assigneeID, userID)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	// 5. Обновляем updated_at доски
	_ = uc.boardRepo.TouchUpdatedAt(ctx, boardID)

	// 6. Публикуем событие
	go func() {
		_ = uc.publisher.PublishCardCreated(context.Background(), CardCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.CreatedAt,
			CardID:       card.ID,
			ColumnID:     card.ColumnID,
			BoardID:      boardID,
			Title:        card.Title,
			Description:  card.Description,
			Position:     card.Position,
		})
	}()

	return card, nil
}
