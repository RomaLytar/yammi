package domain

import (
	"errors"
	"testing"
)

func TestNewCardLink_Valid(t *testing.T) {
	link, err := NewCardLink("", "parent-123", "child-456", "board-123", LinkTypeSubtask)
	if err != nil {
		t.Fatalf("NewCardLink() unexpected error: %v", err)
	}

	if link == nil {
		t.Fatal("NewCardLink() returned nil link")
	}

	if link.ID == "" {
		t.Error("NewCardLink() ID is empty")
	}

	if link.ParentID != "parent-123" {
		t.Errorf("NewCardLink() ParentID = %v, want parent-123", link.ParentID)
	}

	if link.ChildID != "child-456" {
		t.Errorf("NewCardLink() ChildID = %v, want child-456", link.ChildID)
	}

	if link.BoardID != "board-123" {
		t.Errorf("NewCardLink() BoardID = %v, want board-123", link.BoardID)
	}

	if link.LinkType != LinkTypeSubtask {
		t.Errorf("NewCardLink() LinkType = %v, want subtask", link.LinkType)
	}

	if link.CreatedAt.IsZero() {
		t.Error("NewCardLink() CreatedAt is zero")
	}
}

func TestNewCardLink_WithID(t *testing.T) {
	link, err := NewCardLink("custom-id", "parent-123", "child-456", "board-123", LinkTypeSubtask)
	if err != nil {
		t.Fatalf("NewCardLink() unexpected error: %v", err)
	}

	if link.ID != "custom-id" {
		t.Errorf("NewCardLink() ID = %v, want custom-id", link.ID)
	}
}

func TestNewCardLink_SelfLink(t *testing.T) {
	link, err := NewCardLink("", "card-123", "card-123", "board-123", LinkTypeSubtask)
	if !errors.Is(err, ErrSelfLink) {
		t.Errorf("NewCardLink() error = %v, want ErrSelfLink", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestNewCardLink_EmptyParentID(t *testing.T) {
	link, err := NewCardLink("", "", "child-456", "board-123", LinkTypeSubtask)
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("NewCardLink() error = %v, want ErrCardNotFound", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestNewCardLink_EmptyChildID(t *testing.T) {
	link, err := NewCardLink("", "parent-123", "", "board-123", LinkTypeSubtask)
	if !errors.Is(err, ErrCardNotFound) {
		t.Errorf("NewCardLink() error = %v, want ErrCardNotFound", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestNewCardLink_EmptyBoardID(t *testing.T) {
	link, err := NewCardLink("", "parent-123", "child-456", "", LinkTypeSubtask)
	if !errors.Is(err, ErrBoardNotFound) {
		t.Errorf("NewCardLink() error = %v, want ErrBoardNotFound", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestNewCardLink_InvalidType(t *testing.T) {
	link, err := NewCardLink("", "parent-123", "child-456", "board-123", CardLinkType("invalid"))
	if !errors.Is(err, ErrInvalidLinkType) {
		t.Errorf("NewCardLink() error = %v, want ErrInvalidLinkType", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestNewCardLink_EmptyType(t *testing.T) {
	link, err := NewCardLink("", "parent-123", "child-456", "board-123", CardLinkType(""))
	if !errors.Is(err, ErrInvalidLinkType) {
		t.Errorf("NewCardLink() error = %v, want ErrInvalidLinkType", err)
	}
	if link != nil {
		t.Error("NewCardLink() returned non-nil link on error")
	}
}

func TestCardLinkType_IsValid(t *testing.T) {
	tests := []struct {
		linkType CardLinkType
		valid    bool
	}{
		{LinkTypeSubtask, true},
		{CardLinkType("subtask"), true},
		{CardLinkType(""), false},
		{CardLinkType("invalid"), false},
		{CardLinkType("parent"), false},
		{CardLinkType("depends_on"), false},
	}

	for _, tt := range tests {
		t.Run("type_"+string(tt.linkType), func(t *testing.T) {
			result := tt.linkType.IsValid()
			if result != tt.valid {
				t.Errorf("CardLinkType(%q).IsValid() = %v, want %v", tt.linkType, result, tt.valid)
			}
		})
	}
}
