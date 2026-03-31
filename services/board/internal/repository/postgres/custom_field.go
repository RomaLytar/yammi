package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CustomFieldRepository struct {
	db *sql.DB
}

func NewCustomFieldRepository(db *sql.DB) *CustomFieldRepository {
	return &CustomFieldRepository{db: db}
}

// CreateDefinition создает новое определение кастомного поля
func (r *CustomFieldRepository) CreateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(def.Options)
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}

	query := `
		INSERT INTO custom_field_definitions (id, board_id, name, field_type, options, position, required, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err = r.db.ExecContext(ctx, query,
		def.ID, def.BoardID, def.Name, string(def.FieldType), optionsJSON, def.Position, def.Required, def.CreatedAt, def.UpdatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrCustomFieldExists
		}
		return fmt.Errorf("insert custom field definition: %w", err)
	}

	return nil
}

// GetDefinitionByID возвращает определение по ID (фильтруется по boardID для защиты от IDOR)
func (r *CustomFieldRepository) GetDefinitionByID(ctx context.Context, defID, boardID string) (*domain.CustomFieldDefinition, error) {
	query := `
		SELECT id, board_id, name, field_type, options, position, required, created_at, updated_at
		FROM custom_field_definitions
		WHERE id = $1 AND board_id = $2
	`

	var def domain.CustomFieldDefinition
	var fieldType string
	var optionsJSON []byte

	err := r.db.QueryRowContext(ctx, query, defID, boardID).Scan(
		&def.ID, &def.BoardID, &def.Name, &fieldType, &optionsJSON, &def.Position, &def.Required, &def.CreatedAt, &def.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCustomFieldNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select custom field definition: %w", err)
	}

	def.FieldType = domain.FieldType(fieldType)

	if optionsJSON != nil {
		if err := json.Unmarshal(optionsJSON, &def.Options); err != nil {
			return nil, fmt.Errorf("unmarshal options: %w", err)
		}
	}

	return &def, nil
}

// ListDefinitionsByBoardID возвращает все определения кастомных полей доски
func (r *CustomFieldRepository) ListDefinitionsByBoardID(ctx context.Context, boardID string) ([]*domain.CustomFieldDefinition, error) {
	query := `
		SELECT id, board_id, name, field_type, options, position, required, created_at, updated_at
		FROM custom_field_definitions
		WHERE board_id = $1
		ORDER BY position ASC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("select custom field definitions: %w", err)
	}
	defer rows.Close()

	var defs []*domain.CustomFieldDefinition
	for rows.Next() {
		var def domain.CustomFieldDefinition
		var fieldType string
		var optionsJSON []byte

		if err := rows.Scan(&def.ID, &def.BoardID, &def.Name, &fieldType, &optionsJSON, &def.Position, &def.Required, &def.CreatedAt, &def.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan custom field definition: %w", err)
		}

		def.FieldType = domain.FieldType(fieldType)

		if optionsJSON != nil {
			if err := json.Unmarshal(optionsJSON, &def.Options); err != nil {
				return nil, fmt.Errorf("unmarshal options: %w", err)
			}
		}

		defs = append(defs, &def)
	}

	return defs, rows.Err()
}

// UpdateDefinition обновляет определение кастомного поля
func (r *CustomFieldRepository) UpdateDefinition(ctx context.Context, def *domain.CustomFieldDefinition) error {
	optionsJSON, err := json.Marshal(def.Options)
	if err != nil {
		return fmt.Errorf("marshal options: %w", err)
	}

	query := `
		UPDATE custom_field_definitions
		SET name = $1, options = $2, required = $3, updated_at = $4
		WHERE id = $5 AND board_id = $6
	`

	result, err := r.db.ExecContext(ctx, query, def.Name, optionsJSON, def.Required, def.UpdatedAt, def.ID, def.BoardID)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrCustomFieldExists
		}
		return fmt.Errorf("update custom field definition: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCustomFieldNotFound
	}

	return nil
}

// DeleteDefinition удаляет определение кастомного поля (CASCADE удалит значения, фильтруется по boardID)
func (r *CustomFieldRepository) DeleteDefinition(ctx context.Context, defID, boardID string) error {
	query := `DELETE FROM custom_field_definitions WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, defID, boardID)
	if err != nil {
		return fmt.Errorf("delete custom field definition: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCustomFieldNotFound
	}

	return nil
}

// CountDefinitionsByBoardID возвращает количество определений кастомных полей доски
func (r *CustomFieldRepository) CountDefinitionsByBoardID(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM custom_field_definitions WHERE board_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count custom field definitions: %w", err)
	}

	return count, nil
}

// SetValue создает или обновляет значение кастомного поля (upsert по board_id, card_id, field_id)
func (r *CustomFieldRepository) SetValue(ctx context.Context, value *domain.CustomFieldValue) error {
	query := `
		INSERT INTO custom_field_values (id, card_id, board_id, field_id, value_text, value_number, value_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (board_id, card_id, field_id)
		DO UPDATE SET value_text = $5, value_number = $6, value_date = $7, updated_at = $9
	`

	_, err := r.db.ExecContext(ctx, query,
		value.ID, value.CardID, value.BoardID, value.FieldID,
		value.ValueText, value.ValueNumber, value.ValueDate,
		value.CreatedAt, value.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("set custom field value: %w", err)
	}

	return nil
}

// GetCardValues возвращает все значения кастомных полей карточки
func (r *CustomFieldRepository) GetCardValues(ctx context.Context, cardID, boardID string) ([]*domain.CustomFieldValue, error) {
	query := `
		SELECT id, card_id, board_id, field_id, value_text, value_number, value_date, created_at, updated_at
		FROM custom_field_values
		WHERE card_id = $1 AND board_id = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select custom field values: %w", err)
	}
	defer rows.Close()

	var values []*domain.CustomFieldValue
	for rows.Next() {
		var v domain.CustomFieldValue
		if err := rows.Scan(&v.ID, &v.CardID, &v.BoardID, &v.FieldID, &v.ValueText, &v.ValueNumber, &v.ValueDate, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan custom field value: %w", err)
		}
		values = append(values, &v)
	}

	return values, rows.Err()
}

// DeleteValue удаляет значение кастомного поля карточки
func (r *CustomFieldRepository) DeleteValue(ctx context.Context, cardID, boardID, fieldID string) error {
	query := `DELETE FROM custom_field_values WHERE card_id = $1 AND board_id = $2 AND field_id = $3`
	result, err := r.db.ExecContext(ctx, query, cardID, boardID, fieldID)
	if err != nil {
		return fmt.Errorf("delete custom field value: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCustomFieldValueNotFound
	}

	return nil
}
