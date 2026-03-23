package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type LinkCardsUseCase struct {
	cardLinkRepo CardLinkRepository
	cardRepo     CardRepository
	memberRepo   MembershipRepository
	publisher    EventPublisher
}

func NewLinkCardsUseCase(cardLinkRepo CardLinkRepository, cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *LinkCardsUseCase {
	return &LinkCardsUseCase{
		cardLinkRepo: cardLinkRepo,
		cardRepo:     cardRepo,
		memberRepo:   memberRepo,
		publisher:    publisher,
	}
}

func (uc *LinkCardsUseCase) Execute(ctx context.Context, parentID, childID, boardID, userID string) (*domain.CardLink, error) {
	// 1. Проверка доступа (member может создавать связи)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Проверяем что родительская карточка существует
	_, err = uc.cardRepo.GetByID(ctx, parentID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Создаем связь (валидация self-link внутри domain)
	link, err := domain.NewCardLink("", parentID, childID, boardID, domain.LinkTypeSubtask)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем (duplicate constraint в БД вернёт ErrLinkAlreadyExists)
	if err := uc.cardLinkRepo.Create(ctx, link); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishCardLinked(context.Background(), CardLinked{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LinkID:       link.ID,
			ParentID:     parentID,
			ChildID:      childID,
			BoardID:      boardID,
			LinkType:     string(link.LinkType),
			ActorID:      userID,
		})
	}()

	return link, nil
}
