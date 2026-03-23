package domain

import (
	"time"

	"github.com/google/uuid"
)

// FieldType — тип кастомного поля
type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeDropdown FieldType = "dropdown"
)

// IsValid проверяет допустимость типа поля
func (ft FieldType) IsValid() bool {
	switch ft {
	case FieldTypeText, FieldTypeNumber, FieldTypeDate, FieldTypeDropdown:
		return true
	default:
		return false
	}
}

// CustomFieldDefinition — определение кастомного поля доски
type CustomFieldDefinition struct {
	ID        string
	BoardID   string
	Name      string
	FieldType FieldType
	Options   []string // dropdown options (JSON в БД)
	Position  int
	Required  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCustomFieldDefinition создает новое определение кастомного поля с валидацией
func NewCustomFieldDefinition(id, boardID, name string, fieldType FieldType, options []string, position int, required bool) (*CustomFieldDefinition, error) {
	if boardID == "" {
		return nil, ErrBoardNotFound
	}

	if name == "" {
		return nil, ErrEmptyFieldName
	}

	if !fieldType.IsValid() {
		return nil, ErrInvalidFieldType
	}

	if id == "" {
		id = uuid.NewString()
	}

	now := time.Now()

	return &CustomFieldDefinition{
		ID:        id,
		BoardID:   boardID,
		Name:      name,
		FieldType: fieldType,
		Options:   options,
		Position:  position,
		Required:  required,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update обновляет метаданные определения кастомного поля
func (d *CustomFieldDefinition) Update(name string, options []string, required bool) error {
	if name == "" {
		return ErrEmptyFieldName
	}

	d.Name = name
	d.Options = options
	d.Required = required
	d.UpdatedAt = time.Now()

	return nil
}

// CustomFieldValue — значение кастомного поля для карточки
type CustomFieldValue struct {
	ID          string
	CardID      string
	BoardID     string
	FieldID     string
	ValueText   *string
	ValueNumber *float64
	ValueDate   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewCustomFieldValue создает новое значение кастомного поля
func NewCustomFieldValue(id, cardID, boardID, fieldID string) *CustomFieldValue {
	if id == "" {
		id = uuid.NewString()
	}

	now := time.Now()

	return &CustomFieldValue{
		ID:        id,
		CardID:    cardID,
		BoardID:   boardID,
		FieldID:   fieldID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// SetText устанавливает текстовое значение
func (v *CustomFieldValue) SetText(val string) {
	v.ValueText = &val
	v.ValueNumber = nil
	v.ValueDate = nil
	v.UpdatedAt = time.Now()
}

// SetNumber устанавливает числовое значение
func (v *CustomFieldValue) SetNumber(val float64) {
	v.ValueNumber = &val
	v.ValueText = nil
	v.ValueDate = nil
	v.UpdatedAt = time.Now()
}

// SetDate устанавливает значение даты
func (v *CustomFieldValue) SetDate(val time.Time) {
	v.ValueDate = &val
	v.ValueText = nil
	v.ValueNumber = nil
	v.UpdatedAt = time.Now()
}
