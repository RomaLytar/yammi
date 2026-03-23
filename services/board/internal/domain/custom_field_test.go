package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCustomFieldDefinition_ValidText(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Sprint", FieldTypeText, nil, 0, false)
	if err != nil {
		t.Fatalf("NewCustomFieldDefinition() unexpected error: %v", err)
	}

	if def == nil {
		t.Fatal("NewCustomFieldDefinition() returned nil")
	}

	if def.ID == "" {
		t.Error("NewCustomFieldDefinition() ID is empty")
	}

	if def.BoardID != "board-123" {
		t.Errorf("NewCustomFieldDefinition() BoardID = %v, want board-123", def.BoardID)
	}

	if def.Name != "Sprint" {
		t.Errorf("NewCustomFieldDefinition() Name = %v, want Sprint", def.Name)
	}

	if def.FieldType != FieldTypeText {
		t.Errorf("NewCustomFieldDefinition() FieldType = %v, want text", def.FieldType)
	}

	if def.CreatedAt.IsZero() {
		t.Error("NewCustomFieldDefinition() CreatedAt is zero")
	}

	if def.UpdatedAt.IsZero() {
		t.Error("NewCustomFieldDefinition() UpdatedAt is zero")
	}
}

func TestNewCustomFieldDefinition_ValidNumber(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Story Points", FieldTypeNumber, nil, 1, true)
	if err != nil {
		t.Fatalf("NewCustomFieldDefinition() unexpected error: %v", err)
	}

	if def.FieldType != FieldTypeNumber {
		t.Errorf("NewCustomFieldDefinition() FieldType = %v, want number", def.FieldType)
	}

	if !def.Required {
		t.Error("NewCustomFieldDefinition() Required should be true")
	}

	if def.Position != 1 {
		t.Errorf("NewCustomFieldDefinition() Position = %v, want 1", def.Position)
	}
}

func TestNewCustomFieldDefinition_ValidDate(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Start Date", FieldTypeDate, nil, 2, false)
	if err != nil {
		t.Fatalf("NewCustomFieldDefinition() unexpected error: %v", err)
	}

	if def.FieldType != FieldTypeDate {
		t.Errorf("NewCustomFieldDefinition() FieldType = %v, want date", def.FieldType)
	}
}

func TestNewCustomFieldDefinition_ValidDropdown(t *testing.T) {
	options := []string{"Small", "Medium", "Large"}
	def, err := NewCustomFieldDefinition("", "board-123", "T-Shirt Size", FieldTypeDropdown, options, 3, false)
	if err != nil {
		t.Fatalf("NewCustomFieldDefinition() unexpected error: %v", err)
	}

	if def.FieldType != FieldTypeDropdown {
		t.Errorf("NewCustomFieldDefinition() FieldType = %v, want dropdown", def.FieldType)
	}

	if len(def.Options) != 3 {
		t.Errorf("NewCustomFieldDefinition() Options length = %v, want 3", len(def.Options))
	}

	if def.Options[0] != "Small" {
		t.Errorf("NewCustomFieldDefinition() Options[0] = %v, want Small", def.Options[0])
	}
}

func TestNewCustomFieldDefinition_EmptyName(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "", FieldTypeText, nil, 0, false)
	if !errors.Is(err, ErrEmptyFieldName) {
		t.Errorf("NewCustomFieldDefinition() error = %v, want ErrEmptyFieldName", err)
	}
	if def != nil {
		t.Error("NewCustomFieldDefinition() returned non-nil on error")
	}
}

func TestNewCustomFieldDefinition_InvalidType(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Field", FieldType("invalid"), nil, 0, false)
	if !errors.Is(err, ErrInvalidFieldType) {
		t.Errorf("NewCustomFieldDefinition() error = %v, want ErrInvalidFieldType", err)
	}
	if def != nil {
		t.Error("NewCustomFieldDefinition() returned non-nil on error")
	}
}

func TestNewCustomFieldDefinition_EmptyBoardID(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "", "Field", FieldTypeText, nil, 0, false)
	if !errors.Is(err, ErrBoardNotFound) {
		t.Errorf("NewCustomFieldDefinition() error = %v, want ErrBoardNotFound", err)
	}
	if def != nil {
		t.Error("NewCustomFieldDefinition() returned non-nil on error")
	}
}

func TestCustomFieldDefinition_Update(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Sprint", FieldTypeText, nil, 0, false)
	if err != nil {
		t.Fatalf("Failed to create test definition: %v", err)
	}

	err = def.Update("Sprint Number", []string{"1", "2", "3"}, true)
	if err != nil {
		t.Fatalf("CustomFieldDefinition.Update() unexpected error: %v", err)
	}

	if def.Name != "Sprint Number" {
		t.Errorf("CustomFieldDefinition.Update() Name = %v, want Sprint Number", def.Name)
	}

	if len(def.Options) != 3 {
		t.Errorf("CustomFieldDefinition.Update() Options length = %v, want 3", len(def.Options))
	}

	if !def.Required {
		t.Error("CustomFieldDefinition.Update() Required should be true")
	}

	// ID и BoardID не должны измениться
	if def.BoardID != "board-123" {
		t.Error("CustomFieldDefinition.Update() changed BoardID")
	}
}

func TestCustomFieldDefinition_Update_EmptyName(t *testing.T) {
	def, err := NewCustomFieldDefinition("", "board-123", "Sprint", FieldTypeText, nil, 0, false)
	if err != nil {
		t.Fatalf("Failed to create test definition: %v", err)
	}

	err = def.Update("", nil, false)
	if !errors.Is(err, ErrEmptyFieldName) {
		t.Errorf("CustomFieldDefinition.Update() error = %v, want ErrEmptyFieldName", err)
	}

	// При ошибке поля не должны измениться
	if def.Name != "Sprint" {
		t.Error("CustomFieldDefinition.Update() changed Name on error")
	}
}

func TestFieldType_IsValid(t *testing.T) {
	tests := []struct {
		ft    FieldType
		valid bool
	}{
		{FieldTypeText, true},
		{FieldTypeNumber, true},
		{FieldTypeDate, true},
		{FieldTypeDropdown, true},
		{FieldType(""), false},
		{FieldType("invalid"), false},
		{FieldType("boolean"), false},
		{FieldType("TEXT"), false},
	}

	for _, tt := range tests {
		t.Run("type_"+string(tt.ft), func(t *testing.T) {
			result := tt.ft.IsValid()
			if result != tt.valid {
				t.Errorf("FieldType(%q).IsValid() = %v, want %v", tt.ft, result, tt.valid)
			}
		})
	}
}

func TestCustomFieldValue_SetText(t *testing.T) {
	v := NewCustomFieldValue("", "card-123", "board-123", "field-123")

	v.SetText("hello world")

	if v.ValueText == nil || *v.ValueText != "hello world" {
		t.Errorf("SetText() ValueText = %v, want 'hello world'", v.ValueText)
	}
	if v.ValueNumber != nil {
		t.Error("SetText() should clear ValueNumber")
	}
	if v.ValueDate != nil {
		t.Error("SetText() should clear ValueDate")
	}
}

func TestCustomFieldValue_SetNumber(t *testing.T) {
	v := NewCustomFieldValue("", "card-123", "board-123", "field-123")

	v.SetNumber(42.5)

	if v.ValueNumber == nil || *v.ValueNumber != 42.5 {
		t.Errorf("SetNumber() ValueNumber = %v, want 42.5", v.ValueNumber)
	}
	if v.ValueText != nil {
		t.Error("SetNumber() should clear ValueText")
	}
	if v.ValueDate != nil {
		t.Error("SetNumber() should clear ValueDate")
	}
}

func TestCustomFieldValue_SetDate(t *testing.T) {
	v := NewCustomFieldValue("", "card-123", "board-123", "field-123")

	now := time.Now()
	v.SetDate(now)

	if v.ValueDate == nil || !v.ValueDate.Equal(now) {
		t.Errorf("SetDate() ValueDate = %v, want %v", v.ValueDate, now)
	}
	if v.ValueText != nil {
		t.Error("SetDate() should clear ValueText")
	}
	if v.ValueNumber != nil {
		t.Error("SetDate() should clear ValueNumber")
	}
}

func TestNewCustomFieldValue(t *testing.T) {
	v := NewCustomFieldValue("", "card-123", "board-123", "field-123")

	if v.ID == "" {
		t.Error("NewCustomFieldValue() ID is empty")
	}

	if v.CardID != "card-123" {
		t.Errorf("NewCustomFieldValue() CardID = %v, want card-123", v.CardID)
	}

	if v.BoardID != "board-123" {
		t.Errorf("NewCustomFieldValue() BoardID = %v, want board-123", v.BoardID)
	}

	if v.FieldID != "field-123" {
		t.Errorf("NewCustomFieldValue() FieldID = %v, want field-123", v.FieldID)
	}

	if v.CreatedAt.IsZero() {
		t.Error("NewCustomFieldValue() CreatedAt is zero")
	}
}
