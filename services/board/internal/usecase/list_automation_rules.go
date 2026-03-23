package usecase

import (
	"context"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ListAutomationRulesUseCase struct {
	ruleRepo   AutomationRuleRepository
	memberRepo MembershipRepository
}

func NewListAutomationRulesUseCase(ruleRepo AutomationRuleRepository, memberRepo MembershipRepository) *ListAutomationRulesUseCase {
	return &ListAutomationRulesUseCase{
		ruleRepo:   ruleRepo,
		memberRepo: memberRepo,
	}
}

func (uc *ListAutomationRulesUseCase) Execute(ctx context.Context, boardID, userID string) ([]*domain.AutomationRule, error) {
	// 1. Проверка доступа (member может просматривать правила)
	isMember, _, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}

	// 2. Получаем правила доски
	return uc.ruleRepo.ListByBoardID(ctx, boardID)
}
