package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AddMemberUseCase struct {
	boardRepo  BoardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewAddMemberUseCase(boardRepo BoardRepository, memberRepo MembershipRepository, publisher EventPublisher) *AddMemberUseCase {
	return &AddMemberUseCase{
		boardRepo:  boardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *AddMemberUseCase) Execute(ctx context.Context, boardID, userID, memberUserID string, role domain.Role) error {
	// 1. Загружаем доску
	board, err := uc.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	// 2. Проверка: только owner может добавлять members
	if !board.IsOwner(userID) {
		return domain.ErrNotOwner
	}

	// 3. Валидация роли
	if !role.IsValid() {
		return domain.ErrInvalidRole
	}

	// 4. Добавляем member
	if err := uc.memberRepo.AddMember(ctx, boardID, memberUserID, role); err != nil {
		return err
	}

	// 5. Публикуем событие
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishMemberAdded(ctx, MemberAdded{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			BoardID:      boardID,
			UserID:       memberUserID,
			ActorID:      userID,
			Role:         string(role),
			BoardTitle:   board.Title,
		}); err != nil {
			slog.Error("failed to publish MemberAdded", "error", err, "board_id", boardID)
		}
	}()

	return nil
}
