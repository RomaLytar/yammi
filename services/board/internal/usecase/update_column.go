package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateColumnUseCase struct {
	columnRepo ColumnRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateColumnUseCase(columnRepo ColumnRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateColumnUseCase {
	return &UpdateColumnUseCase{
		columnRepo: columnRepo,
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

	// 5. Публикуем событие
	go func() {
		_ = uc.publisher.PublishColumnUpdated(context.Background(), ColumnUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ColumnID:     column.ID,
			Title:        column.Title,
		})
	}()

	return column, nil
}
