package domain

import (
	"testing"
)

func TestNewNotification_Valid(t *testing.T) {
	metadata := map[string]string{"board_id": "123", "board_title": "Test Board"}
	n, err := NewNotification("user-1", TypeBoardCreated, "Board created", "Your board was created", metadata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.ID == "" {
		t.Error("expected non-empty ID")
	}
	if n.UserID != "user-1" {
		t.Errorf("expected UserID=user-1, got %s", n.UserID)
	}
	if n.Type != TypeBoardCreated {
		t.Errorf("expected Type=board_created, got %s", n.Type)
	}
	if n.Title != "Board created" {
		t.Errorf("expected Title=Board created, got %s", n.Title)
	}
	if n.Message != "Your board was created" {
		t.Errorf("expected Message=Your board was created, got %s", n.Message)
	}
	if n.IsRead {
		t.Error("expected IsRead=false")
	}
	if n.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if n.Metadata["board_id"] != "123" {
		t.Errorf("expected metadata board_id=123, got %s", n.Metadata["board_id"])
	}
	if n.Metadata["board_title"] != "Test Board" {
		t.Errorf("expected metadata board_title=Test Board, got %s", n.Metadata["board_title"])
	}
}

func TestNewNotification_EmptyUserID(t *testing.T) {
	_, err := NewNotification("", TypeBoardCreated, "title", "msg", nil)
	if err == nil {
		t.Fatal("expected error for empty userID")
	}
	if err != ErrEmptyUserID {
		t.Errorf("expected ErrEmptyUserID, got %v", err)
	}
}

func TestNewNotification_EmptyType(t *testing.T) {
	_, err := NewNotification("user-1", "", "title", "msg", nil)
	if err == nil {
		t.Fatal("expected error for empty type")
	}
	if err != ErrEmptyType {
		t.Errorf("expected ErrEmptyType, got %v", err)
	}
}

func TestNewNotification_EmptyTitle(t *testing.T) {
	_, err := NewNotification("user-1", TypeWelcome, "", "msg", nil)
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	if err != ErrEmptyTitle {
		t.Errorf("expected ErrEmptyTitle, got %v", err)
	}
}

func TestNewNotification_NilMetadata(t *testing.T) {
	n, err := NewNotification("user-1", TypeWelcome, "Welcome", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Metadata == nil {
		t.Error("expected non-nil metadata map")
	}
	if len(n.Metadata) != 0 {
		t.Errorf("expected empty metadata, got %d entries", len(n.Metadata))
	}
}

func TestNewNotification_EmptyMessage(t *testing.T) {
	n, err := NewNotification("user-1", TypeWelcome, "Welcome", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Message != "" {
		t.Errorf("expected empty message, got %s", n.Message)
	}
}

func TestNewNotification_UniqueIDs(t *testing.T) {
	n1, err := NewNotification("user-1", TypeWelcome, "Title", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n2, err := NewNotification("user-1", TypeWelcome, "Title", "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n1.ID == n2.ID {
		t.Error("expected unique IDs for different notifications")
	}
}

func TestDefaultSettings(t *testing.T) {
	s := DefaultSettings("user-1")

	if s.UserID != "user-1" {
		t.Errorf("expected UserID=user-1, got %s", s.UserID)
	}
	if !s.Enabled {
		t.Error("expected Enabled=true")
	}
	if !s.RealtimeEnabled {
		t.Error("expected RealtimeEnabled=true")
	}
	if s.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if s.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}
