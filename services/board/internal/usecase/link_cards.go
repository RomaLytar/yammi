package usecase

import (
	"context"
	"log/slog"
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
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishCardLinked(ctx, CardLinked{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			LinkID:       link.ID,
			ParentID:     parentID,
			ChildID:      childID,
			BoardID:      boardID,
			LinkType:     string(link.LinkType),
			ActorID:      userID,
		}); err != nil {
			slog.Error("failed to publish CardLinked", "error", err, "link_id", link.ID, "board_id", boardID)
		}
	}()

	return link, nil
}
