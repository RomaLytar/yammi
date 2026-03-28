package domain

import (
	"errors"
	"testing"
)

// ============================================================================
// BoardTemplate Tests
// ============================================================================

func TestNewBoardTemplate_Valid(t *testing.T) {
	tmpl, err := NewBoardTemplate("", "user-123", "Project Board", "Standard project board",
		[]BoardColumnTemplateData{
			{Title: "Backlog", Position: 0},
			{Title: "Sprint", Position: 1},
			{Title: "Done", Position: 2},
		},
		[]LabelTemplateData{
			{Name: "Bug", Color: "#ef4444"},
			{Name: "Feature", Color: "#3b82f6"},
		})
	if err != nil {
		t.Fatalf("NewBoardTemplate() unexpected error: %v", err)
	}

	if tmpl == nil {
		t.Fatal("NewBoardTemplate() returned nil")
	}

	if tmpl.ID == "" {
		t.Error("NewBoardTemplate() ID is empty")
	}

	if tmpl.UserID != "user-123" {
		t.Errorf("NewBoardTemplate() UserID = %v, want user-123", tmpl.UserID)
	}

	if tmpl.Name != "Project Board" {
		t.Errorf("NewBoardTemplate() Name = %v, want Project Board", tmpl.Name)
	}

	if tmpl.Description != "Standard project board" {
		t.Errorf("NewBoardTemplate() Description = %v, want Standard project board", tmpl.Description)
	}

	if len(tmpl.ColumnsData) != 3 {
		t.Errorf("NewBoardTemplate() ColumnsData length = %d, want 3", len(tmpl.ColumnsData))
	}

	if len(tmpl.LabelsData) != 2 {
		t.Errorf("NewBoardTemplate() LabelsData length = %d, want 2", len(tmpl.LabelsData))
	}

	if tmpl.CreatedAt.IsZero() {
		t.Error("NewBoardTemplate() CreatedAt is zero")
	}
}

func TestNewBoardTemplate_EmptyName(t *testing.T) {
	tmpl, err := NewBoardTemplate("", "user-123", "", "desc", nil, nil)
	if !errors.Is(err, ErrEmptyTemplateName) {
		t.Errorf("NewBoardTemplate() error = %v, want ErrEmptyTemplateName", err)
	}
	if tmpl != nil {
		t.Error("NewBoardTemplate() returned non-nil on error")
	}
}

func TestNewBoardTemplate_EmptyUserID(t *testing.T) {
	tmpl, err := NewBoardTemplate("", "", "Template", "desc", nil, nil)
	if !errors.Is(err, ErrEmptyOwnerID) {
		t.Errorf("NewBoardTemplate() error = %v, want ErrEmptyOwnerID", err)
	}
	if tmpl != nil {
		t.Error("NewBoardTemplate() returned non-nil on error")
	}
}

func TestNewBoardTemplate_NilSlicesDefaultToEmpty(t *testing.T) {
	tmpl, err := NewBoardTemplate("", "user-123", "Template", "desc", nil, nil)
	if err != nil {
		t.Fatalf("NewBoardTemplate() unexpected error: %v", err)
	}

	if tmpl.ColumnsData == nil {
		t.Error("NewBoardTemplate() ColumnsData is nil, want empty slice")
	}

	if tmpl.LabelsData == nil {
		t.Error("NewBoardTemplate() LabelsData is nil, want empty slice")
	}
}
