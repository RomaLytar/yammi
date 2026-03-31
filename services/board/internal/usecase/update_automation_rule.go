package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type UpdateAutomationRuleUseCase struct {
	ruleRepo   AutomationRuleRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewUpdateAutomationRuleUseCase(ruleRepo AutomationRuleRepository, memberRepo MembershipRepository, publisher EventPublisher) *UpdateAutomationRuleUseCase {
	return &UpdateAutomationRuleUseCase{
		ruleRepo:   ruleRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *UpdateAutomationRuleUseCase) Execute(ctx context.Context, ruleID, boardID, userID, name string, enabled bool, triggerConfig, actionConfig map[string]string) (*domain.AutomationRule, error) {
	// 1. Проверка доступа (только owner может обновлять правила)
	isMember, role, err := uc.memberRepo.IsMember(ctx, boardID, userID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrAccessDenied
	}
	if role != domain.RoleOwner {
		return nil, domain.ErrNotOwner
	}

	// 2. Загружаем правило (с проверкой boardID)
	rule, err := uc.ruleRepo.GetByID(ctx, ruleID, boardID)
	if err != nil {
		return nil, err
	}

	// 3. Обновляем (валидация внутри)
	if err := rule.Update(name, enabled, triggerConfig, actionConfig); err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishAutomationRuleUpdated(ctx, AutomationRuleUpdated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			RuleID:       rule.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         rule.Name,
			Enabled:      rule.Enabled,
		}); err != nil {
			slog.Error("failed to publish AutomationRuleUpdated", "error", err, "rule_id", rule.ID, "board_id", boardID)
		}
	}()

	return rule, nil
}
