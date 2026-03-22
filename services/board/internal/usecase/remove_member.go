package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type RemoveMemberUseCase struct {
	boardRepo  BoardRepository
	cardRepo   CardRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewRemoveMemberUseCase(boardRepo BoardRepository, cardRepo CardRepository, memberRepo MembershipRepository, publisher EventPublisher) *RemoveMemberUseCase {
	return &RemoveMemberUseCase{
		boardRepo:  boardRepo,
		cardRepo:   cardRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *RemoveMemberUseCase) Execute(ctx context.Context, boardID, userID, memberUserID string) error {
	// 1. Проверка: только owner может удалять участников
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember || role != domain.RoleOwner {
		return domain.ErrAccessDenied
	}

	// 2. Загружаем доску для BoardTitle
	board, err := uc.boardRepo.GetByID(ctx, boardID)
	if err != nil {
		return err
	}

	// 3. Удаляем участника
	if err := uc.memberRepo.RemoveMember(ctx, boardID, memberUserID); err != nil {
		return err
	}

	// 4. Снимаем assignee со всех карточек удалённого участника
	_, _ = uc.cardRepo.UnassignByUser(ctx, boardID, memberUserID)

	// 5. Публикуем событие
	go func() {
		_ = uc.publisher.PublishMemberRemoved(context.Background(), MemberRemoved{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   getCurrentTime(),
			BoardID:      boardID,
			UserID:       memberUserID,
			ActorID:      userID,
			BoardTitle:   board.Title,
		})
	}()

	return nil
}
