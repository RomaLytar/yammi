package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateColumnUseCase struct {
	columnRepo ColumnRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateColumnUseCase(columnRepo ColumnRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateColumnUseCase {
	return &UpdateColumnUseCase{
		columnRepo: columnRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *UpdateColumnUseCase) Execute(ctx context.Context, columnID, boardID, userID, title string, version int) (*domain.Column, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем колонку
	column, err := uc.columnRepo.GetByID(ctx, columnID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем
	if err := column.Update(title); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.columnRepo.Update(ctx, column); err != nil {
		return nil, err
	}

	// 5. Обновляем updated_at доски
	_ = uc.boardRepo.TouchUpdatedAt(ctx, boardID)

	// 6. Публикуем событие
	go func() {
		_ = uc.publisher.PublishColumnUpdated(context.Background(), ColumnUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ColumnID:     column.ID,
			ActorID:      userID,
			Title:        column.Title,
		})
	}()

	return column, nil
}
