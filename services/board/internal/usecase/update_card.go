package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateCardUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateCardUseCase(cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateCardUseCase {
	return &UpdateCardUseCase{
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
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

	// 2. Загружаем карточку
	card, err := uc.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем
	if err := card.Update(title, description, assigneeID); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	// 5. Публикуем событие
	go func() {
		_ = uc.publisher.PublishCardUpdated(context.Background(), CardUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   card.UpdatedAt,
			BoardID:      boardID,
			CardID:       card.ID,
			Title:        card.Title,
			Description:  card.Description,
			AssigneeID:   card.AssigneeID,
		})
	}()

	return card, nil
}
