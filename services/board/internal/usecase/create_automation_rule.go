package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

const maxRulesPerBoard = 25

type CreateAutomationRuleUseCase struct {
	ruleRepo   AutomationRuleRepository
	memberRepo MembershipRepository
	publisher  EventPublisher
}

func NewCreateAutomationRuleUseCase(ruleRepo AutomationRuleRepository, memberRepo MembershipRepository, publisher EventPublisher) *CreateAutomationRuleUseCase {
	return &CreateAutomationRuleUseCase{
		ruleRepo:   ruleRepo,
		memberRepo: memberRepo,
		publisher:  publisher,
	}
}

func (uc *CreateAutomationRuleUseCase) Execute(ctx context.Context, boardID, userID, name string, triggerType domain.TriggerType, triggerConfig map[string]string, actionType domain.ActionType, actionConfig map[string]string) (*domain.AutomationRule, error) {
	// 1. Проверка доступа (только owner может создавать правила)
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

	// 2. Проверка лимита правил на доску
	count, err := uc.ruleRepo.CountByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}
	if count >= maxRulesPerBoard {
		return nil, domain.ErrMaxRulesReached
	}

	// 3. Создаем правило (валидация внутри)
	rule, err := domain.NewAutomationRule("", boardID, name, triggerType, triggerConfig, actionType, actionConfig, userID)
	if err != nil {
		return nil, err
	}

	// 4. Сохраняем
	if err := uc.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}

	// 5. Публикуем событие (async, non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := uc.publisher.PublishAutomationRuleCreated(ctx, AutomationRuleCreated{
			EventID:      generateEventID(),
			EventVersion: 1,
			OccurredAt:   time.Now(),
			RuleID:       rule.ID,
			BoardID:      boardID,
			ActorID:      userID,
			Name:         rule.Name,
			TriggerType:  string(rule.TriggerType),
			ActionType:   string(rule.ActionType),
		}); err != nil {
			slog.Error("failed to publish AutomationRuleCreated", "error", err, "rule_id", rule.ID, "board_id", boardID)
		}
	}()

	return rule, nil
}
