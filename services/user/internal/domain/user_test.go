package domain

import (
	"errors"
	"testing"
	"time"
)

func TestUser_Update(t *testing.T) {
	t.Run("valid update works", func(t *testing.T) {
		user := &User{
			ID:        "user-1",
			Email:     "test@example.com",
			Name:      "Old Name",
			AvatarURL: "https://old-avatar.com/img.png",
			Bio:       "Old bio",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		}
		before := user.UpdatedAt

		err := user.Update("New Name", "https://new-avatar.com/img.png", "New bio")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Name != "New Name" {
			t.Fatalf("expected name New Name, got %s", user.Name)
		}
		if user.AvatarURL != "https://new-avatar.com/img.png" {
			t.Fatalf("expected new avatar URL, got %s", user.AvatarURL)
		}
		if user.Bio != "New bio" {
			t.Fatalf("expected bio New bio, got %s", user.Bio)
		}
		if !user.UpdatedAt.After(before) {
			t.Fatal("expected UpdatedAt to be updated")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		user := &User{
			ID:    "user-2",
			Email: "test@example.com",
			Name:  "Original Name",
		}

		err := user.Update("", "https://avatar.com/img.png", "Some bio")
		if !errors.Is(err, ErrEmptyName) {
			t.Fatalf("expected ErrEmptyName, got %v", err)
		}
		// Name should remain unchanged after failed update
		if user.Name != "Original Name" {
			t.Fatalf("expected name to remain Original Name, got %s", user.Name)
		}
	})

	t.Run("partial update only bio works", func(t *testing.T) {
		user := &User{
			ID:        "user-3",
			Email:     "test@example.com",
			Name:      "Keep Name",
			AvatarURL: "",
			Bio:       "Old bio",
		}

		err := user.Update("Keep Name", "", "Updated bio")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if user.Name != "Keep Name" {
			t.Fatalf("expected name Keep Name, got %s", user.Name)
		}
		if user.AvatarURL != "" {
			t.Fatalf("expected empty avatar URL, got %s", user.AvatarURL)
		}
		if user.Bio != "Updated bio" {
			t.Fatalf("expected bio Updated bio, got %s", user.Bio)
		}
	})
}

func TestNewUserFromEvent(t *testing.T) {
	user := NewUserFromEvent("id-1", "test@example.com", "Test User")

	if user.ID != "id-1" {
		t.Fatalf("expected ID id-1, got %s", user.ID)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected email test@example.com, got %s", user.Email)
	}
	if user.Name != "Test User" {
		t.Fatalf("expected name Test User, got %s", user.Name)
	}
	if user.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if user.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
	}
}
