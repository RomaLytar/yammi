package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	
)

type MoveCardUseCase struct {
	cardRepo   CardRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewMoveCardUseCase(cardRepo CardRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *MoveCardUseCase {
	return &MoveCardUseCase{
		cardRepo:   cardRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *MoveCardUseCase) Execute(ctx context.Context, cardID, boardID, fromColumnID, toColumnID, userID, newPosition string) (*domain.Card, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем карточку (фильтр по boardID — IDOR protection)
	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Валидируем lexorank позицию
	if err := domain.ValidateLexorank(newPosition); err != nil {
		return nil, err
	}

	// 4. Перемещаем карточку (domain logic)
	if err := card.Move(toColumnID, newPosition); err != nil {
		return nil, err
	}

	// 5. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 6. Обновляем updated_at доски + публикуем событие (async, non-blocking)
	go func() {
		_ = uc.boardRepo.TouchUpdatedAt(context.Background(), boardID)
		_ = uc.publisher.PublishCardMoved(context.Background(), CardMoved{
			EventID:        generateEventID(),
			EventVersion:   1,
			OccurredAt:     card.UpdatedAt,
			CardID:         cardID,
			BoardID:        boardID,
			ActorID:        userID,
			FromColumnID:   fromColumnID,
			ToColumnID:     toColumnID,
			NewPosition:    newPosition,
		})
	}()

	return card, nil
}
