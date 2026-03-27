package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
	"github.com/google/uuid"
)

// ExecuteAutomationsUseCase находит и выполняет правила автоматизации
// для заданного триггер-события (например, перемещение карточки в колонку).
type ExecuteAutomationsUseCase struct {
	automationRepo AutomationRuleRepository
	cardRepo       CardRepository
	labelRepo      LabelRepository
}

func NewExecuteAutomationsUseCase(
	automationRepo AutomationRuleRepository,
	cardRepo CardRepository,
	labelRepo LabelRepository,
) *ExecuteAutomationsUseCase {
	return &ExecuteAutomationsUseCase{
		automationRepo: automationRepo,
		cardRepo:       cardRepo,
		labelRepo:      labelRepo,
	}
}

// Execute находит и выполняет подходящие правила автоматизации.
// Вызывается async после операций с карточками — ошибки логируются, не пробрасываются.
func (uc *ExecuteAutomationsUseCase) Execute(
	ctx context.Context,
	boardID, cardID string,
	triggerType domain.TriggerType,
	triggerData map[string]string,
) error {
	// 1. Получаем активные правила для доски и типа триггера
	rules, err := uc.automationRepo.ListEnabledByBoardAndTrigger(ctx, boardID, triggerType)
	if err != nil {
		slog.Error("failed to list automation rules", "error", err, "board_id", boardID, "trigger", triggerType)
		return err
	}

	if len(rules) == 0 {
		return nil
	}

	// 2. Для каждого правила проверяем соответствие trigger config и выполняем action
	for _, rule := range rules {
		if !uc.matchesTrigger(rule, triggerData) {
			continue
		}

		// Выполняем действие и записываем результат
		actionErr := uc.executeAction(ctx, rule, boardID, cardID)

		// 3. Записываем execution (success или failed)
		exec := &domain.AutomationExecution{
			ID:         uuid.NewString(),
			RuleID:     rule.ID,
			BoardID:    boardID,
			CardID:     cardID,
			Status:     "success",
			ExecutedAt: time.Now(),
		}

		if actionErr != nil {
			exec.Status = "failed"
			exec.ErrorMessage = actionErr.Error()
			slog.Error("automation action failed",
				"error", actionErr,
				"rule_id", rule.ID,
				"board_id", boardID,
				"card_id", cardID,
				"action_type", rule.ActionType,
			)
		}

		if err := uc.automationRepo.CreateExecution(ctx, exec); err != nil {
			slog.Error("failed to record automation execution",
				"error", err,
				"rule_id", rule.ID,
				"board_id", boardID,
			)
		}
	}

	return nil
}

// matchesTrigger проверяет, соответствуют ли данные триггера конфигурации правила.
func (uc *ExecuteAutomationsUseCase) matchesTrigger(rule *domain.AutomationRule, triggerData map[string]string) bool {
	switch rule.TriggerType {
	case domain.TriggerCardMovedToColumn:
		// Для card_moved_to_column проверяем column_id в trigger config
		expectedColumnID, ok := rule.TriggerConfig["column_id"]
		if !ok {
			return true // если column_id не указан — матчим любое перемещение
		}
		return triggerData["to_column_id"] == expectedColumnID

	case domain.TriggerCardCreated:
		// Для card_created проверяем column_id в trigger config
		expectedColumnID, ok := rule.TriggerConfig["column_id"]
		if !ok {
			return true // если column_id не указан — матчим создание в любой колонке
		}
		return triggerData["column_id"] == expectedColumnID

	default:
		// Для остальных триггеров: все ключи из config должны совпадать
		for key, expectedValue := range rule.TriggerConfig {
			if triggerData[key] != expectedValue {
				return false
			}
		}
		return true
	}
}

// executeAction выполняет действие правила автоматизации.
func (uc *ExecuteAutomationsUseCase) executeAction(ctx context.Context, rule *domain.AutomationRule, boardID, cardID string) error {
	switch rule.ActionType {
	case domain.ActionAssignMember:
		return uc.actionAssignMember(ctx, rule, boardID, cardID)

	case domain.ActionAddLabel:
		return uc.actionAddLabel(ctx, rule, boardID, cardID)

	case domain.ActionSetPriority:
		return uc.actionSetPriority(ctx, rule, boardID, cardID)

	case domain.ActionMoveCard:
		// Пропускаем move_card для предотвращения рекурсивных петель автоматизации
		slog.Warn("skipping move_card action to prevent recursive loops",
			"rule_id", rule.ID,
			"board_id", boardID,
			"card_id", cardID,
		)
		return nil

	default:
		slog.Warn("unknown action type", "action_type", rule.ActionType, "rule_id", rule.ID)
		return nil
	}
}

// actionAssignMember назначает участника на карточку.
func (uc *ExecuteAutomationsUseCase) actionAssignMember(ctx context.Context, rule *domain.AutomationRule, boardID, cardID string) error {
	userID := rule.ActionConfig["user_id"]
	if userID == "" {
		return errors.New("user_id is required in action config")
	}

	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return err
	}

	card.AssigneeID = &userID
	card.UpdatedAt = time.Now()

	return uc.cardRepo.Update(ctx, card)
}

// actionAddLabel добавляет метку на карточку.
func (uc *ExecuteAutomationsUseCase) actionAddLabel(ctx context.Context, rule *domain.AutomationRule, boardID, cardID string) error {
	labelID := rule.ActionConfig["label_id"]
	if labelID == "" {
		return domain.ErrLabelNotFound
	}

	return uc.labelRepo.AddToCard(ctx, cardID, boardID, labelID)
}

// actionSetPriority устанавливает приоритет карточки.
func (uc *ExecuteAutomationsUseCase) actionSetPriority(ctx context.Context, rule *domain.AutomationRule, boardID, cardID string) error {
	priority := domain.Priority(rule.ActionConfig["priority"])
	if !priority.IsValid() {
		return domain.ErrInvalidPriority
	}

	card, err := uc.cardRepo.GetByID(ctx, cardID, boardID)
	if err != nil {
		return err
	}

	if err := card.Update(card.Title, card.Description, card.AssigneeID, card.DueDate, priority, card.TaskType); err != nil {
		return err
	}

	return uc.cardRepo.Update(ctx, card)
}
