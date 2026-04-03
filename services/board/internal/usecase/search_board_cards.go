package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type SearchBoardCardsUseCase struct {
	cardRepo   CardRepository
	memberRepo MembershipRepository
}

func NewSearchBoardCardsUseCase(cardRepo CardRepository, memberRepo MembershipRepository) *SearchBoardCardsUseCase {
	return &SearchBoardCardsUseCase{cardRepo: cardRepo, memberRepo: memberRepo}
}

func (uc *SearchBoardCardsUseCase) Execute(ctx context.Context, boardID, userID, search, assigneeID, priority, taskType string) ([]*domain.Card, error) {
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// Валидация priority (если указан)
	if priority != "" && !domain.Priority(priority).IsValid() {
		return nil, domain.ErrInvalidPriority
	}

	// Валидация task_type (если указан)
	if taskType != "" && !domain.TaskType(taskType).IsValid() {
		return nil, domain.ErrInvalidTaskType
	}

	return uc.cardRepo.SearchByBoardID(ctx, boardID, search, assigneeID, priority, taskType)
}
