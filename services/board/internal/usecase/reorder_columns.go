package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ReorderColumnsUseCase struct {
	columnRepo ColumnRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewReorderColumnsUseCase(columnRepo ColumnRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *ReorderColumnsUseCase {
	return &ReorderColumnsUseCase{
		columnRepo: columnRepo,
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *ReorderColumnsUseCase) Execute(ctx context.Context, boardID, userID string, positions map[string]int, version int) ([]*domain.Column, error) {
	// 1. Проверка доступа
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Загружаем колонки
	columns, err := uc.columnRepo.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем позиции для каждой колонки
	for columnID, position := range positions {
		for _, col := range columns {
			if col.ID == columnID {
				if err := col.UpdatePosition(position); err != nil {
					return nil, err
				}
				if err := uc.columnRepo.Update(ctx, col); err != nil {
					return nil, err
				}
				break
			}
		}
	}

	// 4. Перезагружаем обновленный список
	columns, err = uc.columnRepo.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	// 5. Обновляем updated_at доски + публикуем событие (async, non-blocking)
	columnIDs := make([]string, len(columns))
	for i, col := range columns {
		columnIDs[i] = col.ID
	}

	go func() {
		_ = uc.boardRepo.TouchUpdatedAt(context.Background(), boardID)
		_ = uc.publisher.PublishColumnsReordered(context.Background(), ColumnsReordered{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			ActorID:      userID,
			Columns:      columnIDs,
		})
	}()

	return columns, nil
}
