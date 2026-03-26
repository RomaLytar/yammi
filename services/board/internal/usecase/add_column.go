package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AddColumnUseCase struct {
	columnRepo ColumnRepository
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewAddColumnUseCase(columnRepo ColumnRepository, boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *AddColumnUseCase {
	return &AddColumnUseCase{
		columnRepo: columnRepo,
		boardRepo:  boardRepo,
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

	// 4. Обновляем updated_at доски + публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.boardRepo.TouchUpdatedAt(ctx, boardID); err != nil {
			slog.Error("failed to touch board updated_at", "error", err, "board_id", boardID)
		}
		if err := uc.publisher.PublishColumnCreated(ctx, ColumnAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			ColumnID:     column.ID,
			BoardID:      column.BoardID,
			ActorID:      userID,
			Title:        column.Title,
			Position:     column.Position,
		}); err != nil {
			slog.Error("failed to publish ColumnAdded", "error", err, "board_id", boardID)
		}
	}()

	return column, nil
}
