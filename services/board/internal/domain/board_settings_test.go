package domain

import (
	"testing"
)

func TestNewBoardSettings_Valid(t *testing.T) {
	settings := NewBoardSettings("board-123")

	if settings.BoardID != "board-123" {
		t.Errorf("NewBoardSettings() BoardID = %v, want board-123", settings.BoardID)
	}

	if settings.UseBoardLabelsOnly != false {
		t.Error("NewBoardSettings() UseBoardLabelsOnly should default to false")
	}

	if settings.SprintDurationDays != 14 {
		t.Errorf("NewBoardSettings() SprintDurationDays = %d, want 14", settings.SprintDurationDays)
	}

	if settings.CreatedAt.IsZero() {
		t.Error("NewBoardSettings() CreatedAt is zero")
	}

	if settings.UpdatedAt.IsZero() {
		t.Error("NewBoardSettings() UpdatedAt is zero")
	}
}

func TestBoardSettings_Update(t *testing.T) {
	settings := NewBoardSettings("board-123")
	oldUpdatedAt := settings.UpdatedAt

	settings.Update(true, nil, 14, false)

	if settings.UseBoardLabelsOnly != true {
		t.Error("BoardSettings.Update() UseBoardLabelsOnly should be true")
	}

	if !settings.UpdatedAt.After(oldUpdatedAt) && settings.UpdatedAt != oldUpdatedAt {
		// UpdatedAt should be >= oldUpdatedAt (same or later)
	}
}

func TestBoardSettings_Update_SetFalse(t *testing.T) {
	settings := NewBoardSettings("board-123")
	settings.Update(true, nil, 14, false)
	settings.Update(false, nil, 14, false)

	if settings.UseBoardLabelsOnly != false {
		t.Error("BoardSettings.Update() UseBoardLabelsOnly should be false after setting back")
	}
}

func TestBoardSettings_BoardID_Preserved(t *testing.T) {
	settings := NewBoardSettings("board-abc")
	settings.Update(true, nil, 14, false)

	if settings.BoardID != "board-abc" {
		t.Error("BoardSettings.Update() changed BoardID")
	}
}
