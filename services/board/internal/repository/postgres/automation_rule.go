package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AutomationRuleRepository struct {
	db *sql.DB
}

func NewAutomationRuleRepository(db *sql.DB) *AutomationRuleRepository {
	return &AutomationRuleRepository{db: db}
}

// Create создает новое правило автоматизации
func (r *AutomationRuleRepository) Create(ctx context.Context, rule *domain.AutomationRule) error {
	triggerConfigJSON, err := json.Marshal(rule.TriggerConfig)
	if err != nil {
		return fmt.Errorf("marshal trigger_config: %w", err)
	}

	actionConfigJSON, err := json.Marshal(rule.ActionConfig)
	if err != nil {
		return fmt.Errorf("marshal action_config: %w", err)
	}

	query := `
		INSERT INTO automation_rules (id, board_id, name, enabled, trigger_type, trigger_config, action_type, action_config, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = r.db.ExecContext(ctx, query,
		rule.ID, rule.BoardID, rule.Name, rule.Enabled,
		string(rule.TriggerType), triggerConfigJSON,
		string(rule.ActionType), actionConfigJSON,
		rule.CreatedBy, rule.CreatedAt, rule.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert automation_rule: %w", err)
	}

	return nil
}

// GetByID возвращает правило по ID (фильтруется по boardID для защиты от IDOR)
func (r *AutomationRuleRepository) GetByID(ctx context.Context, ruleID, boardID string) (*domain.AutomationRule, error) {
	query := `
		SELECT id, board_id, name, enabled, trigger_type, trigger_config, action_type, action_config, created_by, created_at, updated_at
		FROM automation_rules
		WHERE id = $1 AND board_id = $2
	`

	var rule domain.AutomationRule
	var triggerType, actionType string
	var triggerConfigJSON, actionConfigJSON []byte

	err := r.db.QueryRowContext(ctx, query, ruleID, boardID).Scan(
		&rule.ID, &rule.BoardID, &rule.Name, &rule.Enabled,
		&triggerType, &triggerConfigJSON,
		&actionType, &actionConfigJSON,
		&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAutomationRuleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select automation_rule: %w", err)
	}

	rule.TriggerType = domain.TriggerType(triggerType)
	rule.ActionType = domain.ActionType(actionType)

	if err := json.Unmarshal(triggerConfigJSON, &rule.TriggerConfig); err != nil {
		return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
	}
	if err := json.Unmarshal(actionConfigJSON, &rule.ActionConfig); err != nil {
		return nil, fmt.Errorf("unmarshal action_config: %w", err)
	}

	return &rule, nil
}

// ListByBoardID возвращает все правила доски
func (r *AutomationRuleRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.AutomationRule, error) {
	query := `
		SELECT id, board_id, name, enabled, trigger_type, trigger_config, action_type, action_config, created_by, created_at, updated_at
		FROM automation_rules
		WHERE board_id = $1
		ORDER BY created_at ASC
	`

	return r.scanRules(ctx, query, boardID)
}

// ListEnabledByBoardAndTrigger возвращает активные правила по доске и типу триггера
func (r *AutomationRuleRepository) ListEnabledByBoardAndTrigger(ctx context.Context, boardID string, triggerType domain.TriggerType) ([]*domain.AutomationRule, error) {
	query := `
		SELECT id, board_id, name, enabled, trigger_type, trigger_config, action_type, action_config, created_by, created_at, updated_at
		FROM automation_rules
		WHERE board_id = $1 AND trigger_type = $2 AND enabled = TRUE
		ORDER BY created_at ASC
	`

	return r.scanRules(ctx, query, boardID, string(triggerType))
}

// Update обновляет правило автоматизации
func (r *AutomationRuleRepository) Update(ctx context.Context, rule *domain.AutomationRule) error {
	triggerConfigJSON, err := json.Marshal(rule.TriggerConfig)
	if err != nil {
		return fmt.Errorf("marshal trigger_config: %w", err)
	}

	actionConfigJSON, err := json.Marshal(rule.ActionConfig)
	if err != nil {
		return fmt.Errorf("marshal action_config: %w", err)
	}

	query := `
		UPDATE automation_rules
		SET name = $1, enabled = $2, trigger_config = $3, action_config = $4, updated_at = $5
		WHERE id = $6 AND board_id = $7
	`

	result, err := r.db.ExecContext(ctx, query,
		rule.Name, rule.Enabled,
		triggerConfigJSON, actionConfigJSON,
		rule.UpdatedAt, rule.ID, rule.BoardID,
	)
	if err != nil {
		return fmt.Errorf("update automation_rule: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrAutomationRuleNotFound
	}

	return nil
}

// Delete удаляет правило по ID (фильтруется по boardID для защиты от IDOR)
func (r *AutomationRuleRepository) Delete(ctx context.Context, ruleID, boardID string) error {
	query := `DELETE FROM automation_rules WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, ruleID, boardID)
	if err != nil {
		return fmt.Errorf("delete automation_rule: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrAutomationRuleNotFound
	}

	return nil
}

// CountByBoardID возвращает количество правил доски
func (r *AutomationRuleRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM automation_rules WHERE board_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count automation_rules: %w", err)
	}

	return count, nil
}

// CreateExecution создает запись о выполнении правила
func (r *AutomationRuleRepository) CreateExecution(ctx context.Context, exec *domain.AutomationExecution) error {
	query := `
		INSERT INTO automation_executions (id, rule_id, board_id, card_id, trigger_event_id, status, error_message, executed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var cardID, triggerEventID interface{}
	if exec.CardID != "" {
		cardID = exec.CardID
	}
	if exec.TriggerEventID != "" {
		triggerEventID = exec.TriggerEventID
	}

	_, err := r.db.ExecContext(ctx, query,
		exec.ID, exec.RuleID, exec.BoardID, cardID, triggerEventID,
		exec.Status, exec.ErrorMessage, exec.ExecutedAt,
	)
	if err != nil {
		return fmt.Errorf("insert automation_execution: %w", err)
	}

	return nil
}

// ListExecutionsByRuleID возвращает историю выполнений правила
func (r *AutomationRuleRepository) ListExecutionsByRuleID(ctx context.Context, ruleID, boardID string, limit int) ([]*domain.AutomationExecution, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, rule_id, board_id, COALESCE(card_id::text, ''), COALESCE(trigger_event_id::text, ''), status, COALESCE(error_message, ''), executed_at
		FROM automation_executions
		WHERE rule_id = $1 AND board_id = $2
		ORDER BY executed_at DESC
		LIMIT $3
	`

	rows, err := r.db.QueryContext(ctx, query, ruleID, boardID, limit)
	if err != nil {
		return nil, fmt.Errorf("select automation_executions: %w", err)
	}
	defer rows.Close()

	var executions []*domain.AutomationExecution
	for rows.Next() {
		var exec domain.AutomationExecution
		if err := rows.Scan(
			&exec.ID, &exec.RuleID, &exec.BoardID, &exec.CardID, &exec.TriggerEventID,
			&exec.Status, &exec.ErrorMessage, &exec.ExecutedAt,
		); err != nil {
			return nil, fmt.Errorf("scan automation_execution: %w", err)
		}
		executions = append(executions, &exec)
	}

	return executions, rows.Err()
}

// scanRules — вспомогательный метод для сканирования списка правил
func (r *AutomationRuleRepository) scanRules(ctx context.Context, query string, args ...interface{}) ([]*domain.AutomationRule, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("select automation_rules: %w", err)
	}
	defer rows.Close()

	var rules []*domain.AutomationRule
	for rows.Next() {
		var rule domain.AutomationRule
		var triggerType, actionType string
		var triggerConfigJSON, actionConfigJSON []byte

		if err := rows.Scan(
			&rule.ID, &rule.BoardID, &rule.Name, &rule.Enabled,
			&triggerType, &triggerConfigJSON,
			&actionType, &actionConfigJSON,
			&rule.CreatedBy, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan automation_rule: %w", err)
		}

		rule.TriggerType = domain.TriggerType(triggerType)
		rule.ActionType = domain.ActionType(actionType)

		if err := json.Unmarshal(triggerConfigJSON, &rule.TriggerConfig); err != nil {
			return nil, fmt.Errorf("unmarshal trigger_config: %w", err)
		}
		if err := json.Unmarshal(actionConfigJSON, &rule.ActionConfig); err != nil {
			return nil, fmt.Errorf("unmarshal action_config: %w", err)
		}

		rules = append(rules, &rule)
	}

	return rules, rows.Err()
}
