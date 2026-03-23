package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type GetAutomationHistoryUseCase struct {
	ruleRepo   AutomationRuleRepository
	memberRepo MembershipRepository
}

func NewGetAutomationHistoryUseCase(ruleRepo AutomationRuleRepository, memberRepo MembershipRepository) *GetAutomationHistoryUseCase {
	return &GetAutomationHistoryUseCase{
		ruleRepo:   ruleRepo,
		memberRepo: memberRepo,
	}
}

func (uc *GetAutomationHistoryUseCase) Execute(ctx context.Context, ruleID, boardID, userID string, limit int) ([]*domain.AutomationExecution, error) {
	// 1. Проверка доступа (member может просматривать историю)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем историю выполнений
	return uc.ruleRepo.ListExecutionsByRuleID(ctx, ruleID, boardID, limit)
}
