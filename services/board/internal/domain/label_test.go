package domain

import (
	"errors"
	"testing"
)

func TestNewLabel_Valid(t *testing.T) {
	label, err := NewLabel("", "board-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("NewLabel() unexpected error: %v", err)
	}

	if label == nil {
		t.Fatal("NewLabel() returned nil label")
	}

	if label.ID == "" {
		t.Error("NewLabel() ID is empty")
	}

	if label.BoardID != "board-123" {
		t.Errorf("NewLabel() BoardID = %v, want board-123", label.BoardID)
	}

	if label.Name != "Bug" {
		t.Errorf("NewLabel() Name = %v, want Bug", label.Name)
	}

	if label.Color != "#ef4444" {
		t.Errorf("NewLabel() Color = %v, want #ef4444", label.Color)
	}

	if label.CreatedAt.IsZero() {
		t.Error("NewLabel() CreatedAt is zero")
	}
}

func TestNewLabel_EmptyName(t *testing.T) {
	label, err := NewLabel("", "board-123", "", "#ef4444")
	if !errors.Is(err, ErrEmptyLabelName) {
		t.Errorf("NewLabel() error = %v, want ErrEmptyLabelName", err)
	}
	if label != nil {
		t.Error("NewLabel() returned non-nil label on error")
	}
}

func TestNewLabel_InvalidColor(t *testing.T) {
	invalidColors := []string{
		"",
		"red",
		"#fff",
		"#gggggg",
		"ef4444",
		"#ef444",
		"#ef44444",
		"123456",
		"#12345g",
	}

	for _, color := range invalidColors {
		t.Run("color_"+color, func(t *testing.T) {
			label, err := NewLabel("", "board-123", "Bug", color)
			if !errors.Is(err, ErrInvalidColor) {
				t.Errorf("NewLabel() with color %q error = %v, want ErrInvalidColor", color, err)
			}
			if label != nil {
				t.Error("NewLabel() returned non-nil label on error")
			}
		})
	}
}

func TestNewLabel_ValidColors(t *testing.T) {
	validColors := []string{
		"#000000",
		"#ffffff",
		"#FFFFFF",
		"#ef4444",
		"#3b82f6",
		"#22c55e",
		"#eab308",
		"#6b7280",
		"#a855f7",
		"#Aa11Bb",
	}

	for _, color := range validColors {
		t.Run("color_"+color, func(t *testing.T) {
			label, err := NewLabel("", "board-123", "Test", color)
			if err != nil {
				t.Errorf("NewLabel() with valid color %q returned error: %v", color, err)
			}
			if label == nil {
				t.Error("NewLabel() returned nil label for valid color")
			}
		})
	}
}

func TestNewLabel_EmptyBoardID(t *testing.T) {
	label, err := NewLabel("", "", "Bug", "#ef4444")
	if !errors.Is(err, ErrBoardNotFound) {
		t.Errorf("NewLabel() error = %v, want ErrBoardNotFound", err)
	}
	if label != nil {
		t.Error("NewLabel() returned non-nil label on error")
	}
}

func TestLabel_Update(t *testing.T) {
	label, err := NewLabel("", "board-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("Feature", "#3b82f6")
	if err != nil {
		t.Fatalf("Label.Update() unexpected error: %v", err)
	}

	if label.Name != "Feature" {
		t.Errorf("Label.Update() Name = %v, want Feature", label.Name)
	}

	if label.Color != "#3b82f6" {
		t.Errorf("Label.Update() Color = %v, want #3b82f6", label.Color)
	}

	// ID и BoardID не должны измениться
	if label.BoardID != "board-123" {
		t.Error("Label.Update() changed BoardID")
	}
}

func TestLabel_Update_EmptyName(t *testing.T) {
	label, err := NewLabel("", "board-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("", "#3b82f6")
	if !errors.Is(err, ErrEmptyLabelName) {
		t.Errorf("Label.Update() error = %v, want ErrEmptyLabelName", err)
	}

	// При ошибке поля не должны измениться
	if label.Name != "Bug" {
		t.Error("Label.Update() changed Name on error")
	}
}

func TestLabel_Update_InvalidColor(t *testing.T) {
	label, err := NewLabel("", "board-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("Feature", "invalid")
	if !errors.Is(err, ErrInvalidColor) {
		t.Errorf("Label.Update() error = %v, want ErrInvalidColor", err)
	}

	// При ошибке поля не должны измениться
	if label.Color != "#ef4444" {
		t.Error("Label.Update() changed Color on error")
	}
}

func TestValidateColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#000000", true},
		{"#ffffff", true},
		{"#FFFFFF", true},
		{"#ef4444", true},
		{"#aAbBcC", true},
		{"", false},
		{"red", false},
		{"#fff", false},
		{"#gggggg", false},
		{"ef4444", false},
		{"#ef444", false},
		{"#ef44444", false},
	}

	for _, tt := range tests {
		t.Run("validate_"+tt.color, func(t *testing.T) {
			result := ValidateColor(tt.color)
			if result != tt.valid {
				t.Errorf("ValidateColor(%q) = %v, want %v", tt.color, result, tt.valid)
			}
		})
	}
}
