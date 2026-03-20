package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	
)

type MoveCardUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewMoveCardUseCase(cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *MoveCardUseCase {
	return &MoveCardUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *MoveCardUseCase) Execute(ctx context.Context, cardID, boardID, fromColumnID, toColumnID, userID string, targetPosition int) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточку
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// 3. Получаем карточки в целевой колонке
	cardsInColumn, err := uc.cardRepo.ListByColumnID(ctx, toColumnID)
	if err != nil {
		return nil, err
	}

	// 4. Вычисляем новую lexorank позицию
	var prevPosition, nextPosition string
	if targetPosition == 0 {
		// В начало колонки
		if len(cardsInColumn) > 0 {
			nextPosition = cardsInColumn[0].Position
		}
	} else if targetPosition >= len(cardsInColumn) {
		// В конец колонки
		if len(cardsInColumn) > 0 {
			prevPosition = cardsInColumn[len(cardsInColumn)-1].Position
		}
	} else {
		// Между карточками
		prevPosition = cardsInColumn[targetPosition-1].Position
		nextPosition = cardsInColumn[targetPosition].Position
	}

	newPosition, err := domain.LexorankBetween(prevPosition, nextPosition)
	if err != nil {
		return nil, err
	}

	// 5. Перемещаем карточку (domain logic)
	if err := card.Move(toColumnID, newPosition); err != nil {
		return nil, err
	}

	// 6. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 7. Публикуем событие
	go func() {
		_ = uc.publisher.PublishCardMoved(context.Background(), CardMoved{
			EventID:        generateEventID(),
			EventVersion:   1,
			OccurredAt:     card.UpdatedAt,
			CardID:         cardID,
			BoardID:        boardID,
			FromColumnID:   fromColumnID,
			ToColumnID:     toColumnID,
			NewPosition:    newPosition,
		})
	}()

	return card, nil
}
