package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	
)

type AddColumnUseCase struct {
	columnRepo ColumnRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewAddColumnUseCase(columnRepo ColumnRepository, memberRepo MembershipRepository, publisher EventPublisher) *AddColumnUseCase {
	return &AddColumnUseCase{
		columnRepo: columnRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *AddColumnUseCase) Execute(ctx context.Context, boardID, userID, title string, position int) (*domain.Column, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Создаем колонку
	column, err := domain.NewColumn(boardID, title, position)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем
	if err := uc.columnRepo.Create(ctx, column); err != nil {
		return nil, err
	}

	// 4. Публикуем событие
	go func() {
		_ = uc.publisher.PublishColumnCreated(context.Background(), ColumnAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ColumnID:     column.ID,
			BoardID:      column.BoardID,
			Title:        column.Title,
			Position:     column.Position,
		})
	}()

	return column, nil
}
