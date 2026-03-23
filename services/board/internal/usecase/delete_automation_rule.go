package usecase

import (
	"context"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type DeleteAutomationRuleUseCase struct {
	ruleRepo   AutomationRuleRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewDeleteAutomationRuleUseCase(ruleRepo AutomationRuleRepository, memberRepo MembershipRepository, publisher EventPublisher) *DeleteAutomationRuleUseCase {
	return &DeleteAutomationRuleUseCase{
		ruleRepo:   ruleRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *DeleteAutomationRuleUseCase) Execute(ctx context.Context, ruleID, boardID, userID string) error {
	// 1. Проверка доступа (только owner может удалять правила)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return err
	}
	if !isMember {
		return domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return domain.ErrNotOwner
	}

	// 2. Удаляем правило (CASCADE удалит executions)
	if err := uc.ruleRepo.Delete(ctx, ruleID); err != nil {
		return err
	}

	// 3. Публикуем событие (async, non-blocking)
	go func() {
		_ = uc.publisher.PublishAutomationRuleDeleted(context.Background(), AutomationRuleDeleted{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			RuleID:       ruleID,
			BoardID:      boardID,
			ActorID:      userID,
		})
	}()

	return nil
}
