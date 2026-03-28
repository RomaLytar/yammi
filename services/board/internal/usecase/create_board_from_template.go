package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CreateBoardFromTemplateUseCase struct {
	boardTmplRepo BoardTemplateRepository
	boardRepo     BoardRepository
	memberRepo    MembershipRepository
	columnRepo    ColumnRepository
	labelRepo     LabelRepository
	publisher     EventPublisher
}

func NewCreateBoardFromTemplateUseCase(
	boardTmplRepo BoardTemplateRepository,
	boardRepo BoardRepository,
	memberRepo MembershipRepository,
	columnRepo ColumnRepository,
	labelRepo LabelRepository,
	publisher EventPublisher,
) *CreateBoardFromTemplateUseCase {
	return &CreateBoardFromTemplateUseCase{
		boardTmplRepo: boardTmplRepo,
		boardRepo:     boardRepo,
		memberRepo:    memberRepo,
		columnRepo:    columnRepo,
		labelRepo:     labelRepo,
		publisher:     publisher,
	}
}

func (uc *CreateBoardFromTemplateUseCase) Execute(ctx context.Context, templateID, title, userID string) (*domain.Board, error) {
	// 1. Загружаем шаблон
	tmpl, err := uc.boardTmplRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// 2. Создаем доску (используем title из параметра, description из шаблона)
	board, err := domain.NewBoard(title, tmpl.Description, userID)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем доску (автоматически создает owner в board_members)
	if err := uc.boardRepo.Create(ctx, board); err != nil {
		return nil, err
	}

	// 4. Создаем колонки из шаблона (batch insert — один запрос вместо N)
	if len(tmpl.ColumnsData) > 0 {
		columns := make([]*domain.Column, 0, len(tmpl.ColumnsData))
		for _, colData := range tmpl.ColumnsData {
			col, err := domain.NewColumn(board.ID, colData.Title, colData.Position)
			if err != nil {
				slog.Error("failed to create column from template", "error", err, "board_id", board.ID)
				continue
			}
			columns = append(columns, col)
		}
		if err := uc.columnRepo.BatchCreate(ctx, columns); err != nil {
			slog.Error("failed to batch create columns from template", "error", err, "board_id", board.ID)
		}
	}

	// 5. Создаем метки из шаблона (batch insert — один запрос вместо M)
	if len(tmpl.LabelsData) > 0 {
		labels := make([]*domain.Label, 0, len(tmpl.LabelsData))
		for _, labelData := range tmpl.LabelsData {
			label, err := domain.NewLabel("", board.ID, labelData.Name, labelData.Color)
			if err != nil {
				slog.Error("failed to create label from template", "error", err, "board_id", board.ID)
				continue
			}
			labels = append(labels, label)
		}
		if err := uc.labelRepo.BatchCreate(ctx, labels); err != nil {
			slog.Error("failed to batch create labels from template", "error", err, "board_id", board.ID)
		}
	}

	// 6. Публикуем события (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishBoardCreated(ctx, BoardCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.CreatedAt,
			BoardID:      board.ID,
			OwnerID:      board.OwnerID,
			Title:        board.Title,
			Description:  board.Description,
		}); err != nil {
			slog.Error("failed to publish BoardCreated", "error", err, "board_id", board.ID)
		}
		if err := uc.publisher.PublishMemberAdded(ctx, MemberAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   board.CreatedAt,
			BoardID:      board.ID,
			UserID:       board.OwnerID,
			ActorID:      board.OwnerID,
			Role:         string(domain.RoleOwner),
			BoardTitle:   board.Title,
		}); err != nil {
			slog.Error("failed to publish MemberAdded", "error", err, "board_id", board.ID)
		}
	}()

	return board, nil
}
