package domain

import (
	"errors"
	"testing"
)

func TestNewUserLabel_Valid(t *testing.T) {
	label, err := NewUserLabel("", "user-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("NewUserLabel() unexpected error: %v", err)
	}

	if label == nil {
		t.Fatal("NewUserLabel() returned nil label")
	}

	if label.ID == "" {
		t.Error("NewUserLabel() ID is empty")
	}

	if label.UserID != "user-123" {
		t.Errorf("NewUserLabel() UserID = %v, want user-123", label.UserID)
	}

	if label.Name != "Bug" {
		t.Errorf("NewUserLabel() Name = %v, want Bug", label.Name)
	}

	if label.Color != "#ef4444" {
		t.Errorf("NewUserLabel() Color = %v, want #ef4444", label.Color)
	}

	if label.CreatedAt.IsZero() {
		t.Error("NewUserLabel() CreatedAt is zero")
	}
}

func TestNewUserLabel_WithID(t *testing.T) {
	label, err := NewUserLabel("my-id", "user-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("NewUserLabel() unexpected error: %v", err)
	}

	if label.ID != "my-id" {
		t.Errorf("NewUserLabel() ID = %v, want my-id", label.ID)
	}
}

func TestNewUserLabel_EmptyUserID(t *testing.T) {
	label, err := NewUserLabel("", "", "Bug", "#ef4444")
	if err == nil {
		t.Error("NewUserLabel() expected error for empty userID")
	}
	if label != nil {
		t.Error("NewUserLabel() returned non-nil label on error")
	}
}

func TestNewUserLabel_EmptyName(t *testing.T) {
	label, err := NewUserLabel("", "user-123", "", "#ef4444")
	if !errors.Is(err, ErrEmptyLabelName) {
		t.Errorf("NewUserLabel() error = %v, want ErrEmptyLabelName", err)
	}
	if label != nil {
		t.Error("NewUserLabel() returned non-nil label on error")
	}
}

func TestNewUserLabel_InvalidColor(t *testing.T) {
	invalidColors := []string{
		"",
		"red",
		"#fff",
		"#gggggg",
		"ef4444",
		"#ef444",
		"#ef44444",
	}

	for _, color := range invalidColors {
		t.Run("color_"+color, func(t *testing.T) {
			label, err := NewUserLabel("", "user-123", "Bug", color)
			if !errors.Is(err, ErrInvalidColor) {
				t.Errorf("NewUserLabel() with color %q error = %v, want ErrInvalidColor", color, err)
			}
			if label != nil {
				t.Error("NewUserLabel() returned non-nil label on error")
			}
		})
	}
}

func TestNewUserLabel_ValidColors(t *testing.T) {
	validColors := []string{
		"#000000",
		"#ffffff",
		"#FFFFFF",
		"#ef4444",
		"#3b82f6",
		"#6b7280",
	}

	for _, color := range validColors {
		t.Run("color_"+color, func(t *testing.T) {
			label, err := NewUserLabel("", "user-123", "Test", color)
			if err != nil {
				t.Errorf("NewUserLabel() with valid color %q returned error: %v", color, err)
			}
			if label == nil {
				t.Error("NewUserLabel() returned nil label for valid color")
			}
		})
	}
}

func TestUserLabel_Update(t *testing.T) {
	label, err := NewUserLabel("", "user-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("Feature", "#3b82f6")
	if err != nil {
		t.Fatalf("UserLabel.Update() unexpected error: %v", err)
	}

	if label.Name != "Feature" {
		t.Errorf("UserLabel.Update() Name = %v, want Feature", label.Name)
	}

	if label.Color != "#3b82f6" {
		t.Errorf("UserLabel.Update() Color = %v, want #3b82f6", label.Color)
	}

	// UserID не должен измениться
	if label.UserID != "user-123" {
		t.Error("UserLabel.Update() changed UserID")
	}
}

func TestUserLabel_Update_EmptyName(t *testing.T) {
	label, err := NewUserLabel("", "user-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("", "#3b82f6")
	if !errors.Is(err, ErrEmptyLabelName) {
		t.Errorf("UserLabel.Update() error = %v, want ErrEmptyLabelName", err)
	}

	// При ошибке поля не должны измениться
	if label.Name != "Bug" {
		t.Error("UserLabel.Update() changed Name on error")
	}
}

func TestUserLabel_Update_InvalidColor(t *testing.T) {
	label, err := NewUserLabel("", "user-123", "Bug", "#ef4444")
	if err != nil {
		t.Fatalf("Failed to create test label: %v", err)
	}

	err = label.Update("Feature", "invalid")
	if !errors.Is(err, ErrInvalidColor) {
		t.Errorf("UserLabel.Update() error = %v, want ErrInvalidColor", err)
	}

	// При ошибке поля не должны измениться
	if label.Color != "#ef4444" {
		t.Error("UserLabel.Update() changed Color on error")
	}
}
