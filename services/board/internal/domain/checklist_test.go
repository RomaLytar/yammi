package domain

import (
	"errors"
	"testing"
)

func TestNewChecklist_Valid(t *testing.T) {
	checklist, err := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)
	if err != nil {
		t.Fatalf("NewChecklist() unexpected error: %v", err)
	}

	if checklist == nil {
		t.Fatal("NewChecklist() returned nil checklist")
	}

	if checklist.ID == "" {
		t.Error("NewChecklist() ID is empty")
	}

	if checklist.CardID != "card-123" {
		t.Errorf("NewChecklist() CardID = %v, want card-123", checklist.CardID)
	}

	if checklist.BoardID != "board-123" {
		t.Errorf("NewChecklist() BoardID = %v, want board-123", checklist.BoardID)
	}

	if checklist.Title != "Review Tasks" {
		t.Errorf("NewChecklist() Title = %v, want Review Tasks", checklist.Title)
	}

	if checklist.Position != 0 {
		t.Errorf("NewChecklist() Position = %v, want 0", checklist.Position)
	}

	if checklist.CreatedAt.IsZero() {
		t.Error("NewChecklist() CreatedAt is zero")
	}

	if checklist.UpdatedAt.IsZero() {
		t.Error("NewChecklist() UpdatedAt is zero")
	}
}

func TestNewChecklist_EmptyTitle(t *testing.T) {
	checklist, err := NewChecklist("", "card-123", "board-123", "", 0)
	if !errors.Is(err, ErrEmptyChecklistTitle) {
		t.Errorf("NewChecklist() error = %v, want ErrEmptyChecklistTitle", err)
	}
	if checklist != nil {
		t.Error("NewChecklist() returned non-nil checklist on error")
	}
}

func TestNewChecklist_EmptyCardID(t *testing.T) {
	checklist, err := NewChecklist("", "", "board-123", "Review Tasks", 0)
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("NewChecklist() error = %v, want ErrCardNotFound", err)
	}
	if checklist != nil {
		t.Error("NewChecklist() returned non-nil checklist on error")
	}
}

func TestNewChecklist_EmptyBoardID(t *testing.T) {
	checklist, err := NewChecklist("", "card-123", "", "Review Tasks", 0)
	if !errors.Is(err, ErrBoardNotFound) {
		t.Errorf("NewChecklist() error = %v, want ErrBoardNotFound", err)
	}
	if checklist != nil {
		t.Error("NewChecklist() returned non-nil checklist on error")
	}
}

func TestChecklist_Update(t *testing.T) {
	checklist, err := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)
	if err != nil {
		t.Fatalf("Failed to create test checklist: %v", err)
	}

	err = checklist.Update("Deploy Tasks")
	if err != nil {
		t.Fatalf("Checklist.Update() unexpected error: %v", err)
	}

	if checklist.Title != "Deploy Tasks" {
		t.Errorf("Checklist.Update() Title = %v, want Deploy Tasks", checklist.Title)
	}

	// ID, CardID, BoardID не должны измениться
	if checklist.CardID != "card-123" {
		t.Error("Checklist.Update() changed CardID")
	}
	if checklist.BoardID != "board-123" {
		t.Error("Checklist.Update() changed BoardID")
	}
}

func TestChecklist_Update_EmptyTitle(t *testing.T) {
	checklist, err := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)
	if err != nil {
		t.Fatalf("Failed to create test checklist: %v", err)
	}

	err = checklist.Update("")
	if !errors.Is(err, ErrEmptyChecklistTitle) {
		t.Errorf("Checklist.Update() error = %v, want ErrEmptyChecklistTitle", err)
	}

	// При ошибке поля не должны измениться
	if checklist.Title != "Review Tasks" {
		t.Error("Checklist.Update() changed Title on error")
	}
}

func TestChecklist_Progress_NoItems(t *testing.T) {
	checklist, _ := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)

	progress := checklist.Progress()
	if progress != 0 {
		t.Errorf("Checklist.Progress() = %v, want 0 (no items)", progress)
	}
}

func TestChecklist_Progress_SomeChecked(t *testing.T) {
	checklist, _ := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)
	checklist.Items = []ChecklistItem{
		{ID: "item-1", IsChecked: true},
		{ID: "item-2", IsChecked: false},
		{ID: "item-3", IsChecked: true},
		{ID: "item-4", IsChecked: false},
	}

	progress := checklist.Progress()
	if progress != 50 {
		t.Errorf("Checklist.Progress() = %v, want 50", progress)
	}
}

func TestChecklist_Progress_AllChecked(t *testing.T) {
	checklist, _ := NewChecklist("", "card-123", "board-123", "Review Tasks", 0)
	checklist.Items = []ChecklistItem{
		{ID: "item-1", IsChecked: true},
		{ID: "item-2", IsChecked: true},
		{ID: "item-3", IsChecked: true},
	}

	progress := checklist.Progress()
	if progress != 100 {
		t.Errorf("Checklist.Progress() = %v, want 100", progress)
	}
}

func TestNewChecklistItem_Valid(t *testing.T) {
	item, err := NewChecklistItem("", "checklist-123", "board-123", "Write tests", 0)
	if err != nil {
		t.Fatalf("NewChecklistItem() unexpected error: %v", err)
	}

	if item == nil {
		t.Fatal("NewChecklistItem() returned nil item")
	}

	if item.ID == "" {
		t.Error("NewChecklistItem() ID is empty")
	}

	if item.ChecklistID != "checklist-123" {
		t.Errorf("NewChecklistItem() ChecklistID = %v, want checklist-123", item.ChecklistID)
	}

	if item.BoardID != "board-123" {
		t.Errorf("NewChecklistItem() BoardID = %v, want board-123", item.BoardID)
	}

	if item.Title != "Write tests" {
		t.Errorf("NewChecklistItem() Title = %v, want Write tests", item.Title)
	}

	if item.IsChecked {
		t.Error("NewChecklistItem() IsChecked should be false by default")
	}

	if item.CreatedAt.IsZero() {
		t.Error("NewChecklistItem() CreatedAt is zero")
	}
}

func TestNewChecklistItem_EmptyTitle(t *testing.T) {
	item, err := NewChecklistItem("", "checklist-123", "board-123", "", 0)
	if !errors.Is(err, ErrEmptyItemTitle) {
		t.Errorf("NewChecklistItem() error = %v, want ErrEmptyItemTitle", err)
	}
	if item != nil {
		t.Error("NewChecklistItem() returned non-nil item on error")
	}
}

func TestNewChecklistItem_EmptyChecklistID(t *testing.T) {
	item, err := NewChecklistItem("", "", "board-123", "Write tests", 0)
	if !errors.Is(err, ErrChecklistNotFound) {
		t.Errorf("NewChecklistItem() error = %v, want ErrChecklistNotFound", err)
	}
	if item != nil {
		t.Error("NewChecklistItem() returned non-nil item on error")
	}
}

func TestChecklistItem_Update(t *testing.T) {
	item, _ := NewChecklistItem("", "checklist-123", "board-123", "Write tests", 0)

	err := item.Update("Write unit tests")
	if err != nil {
		t.Fatalf("ChecklistItem.Update() unexpected error: %v", err)
	}

	if item.Title != "Write unit tests" {
		t.Errorf("ChecklistItem.Update() Title = %v, want Write unit tests", item.Title)
	}
}

func TestChecklistItem_Update_EmptyTitle(t *testing.T) {
	item, _ := NewChecklistItem("", "checklist-123", "board-123", "Write tests", 0)

	err := item.Update("")
	if !errors.Is(err, ErrEmptyItemTitle) {
		t.Errorf("ChecklistItem.Update() error = %v, want ErrEmptyItemTitle", err)
	}

	// При ошибке поля не должны измениться
	if item.Title != "Write tests" {
		t.Error("ChecklistItem.Update() changed Title on error")
	}
}

func TestChecklistItem_Toggle(t *testing.T) {
	item, _ := NewChecklistItem("", "checklist-123", "board-123", "Write tests", 0)

	if item.IsChecked {
		t.Fatal("IsChecked should be false initially")
	}

	// Toggle on
	item.Toggle()
	if !item.IsChecked {
		t.Error("After first Toggle(), IsChecked should be true")
	}

	// Toggle off
	item.Toggle()
	if item.IsChecked {
		t.Error("After second Toggle(), IsChecked should be false")
	}
}
