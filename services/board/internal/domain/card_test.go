package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewCard(t *testing.T) {
	assignee1 := "user-123"
	assignee2 := "user-456"

	tests := []struct {
		name        string
		columnID    string
		title       string
		description string
		position    string
		assigneeID  *string
		wantErr     error
	}{
		{
			name:        "valid card with assignee",
			columnID:    "column-123",
			title:       "Task 1",
			description: "Description",
			position:    "n",
			assigneeID:  &assignee1,
			wantErr:     nil,
		},
		{
			name:        "valid card without assignee",
			columnID:    "column-123",
			title:       "Task 2",
			description: "Description",
			position:    "n",
			assigneeID:  nil,
			wantErr:     nil,
		},
		{
			name:        "valid card with different position",
			columnID:    "column-123",
			title:       "Task 3",
			description: "",
			position:    "aaa",
			assigneeID:  &assignee2,
			wantErr:     nil,
		},
		{
			name:        "empty column ID",
			columnID:    "",
			title:       "Task",
			description: "Description",
			position:    "n",
			assigneeID:  nil,
			wantErr:     ErrColumnNotFound,
		},
		{
			name:        "empty title",
			columnID:    "column-123",
			title:       "",
			description: "Description",
			position:    "n",
			assigneeID:  nil,
			wantErr:     ErrEmptyCardTitle,
		},
		{
			name:        "empty position",
			columnID:    "column-123",
			title:       "Task",
			description: "Description",
			position:    "",
			assigneeID:  nil,
			wantErr:     ErrInvalidLexorank,
		},
		{
			name:        "invalid lexorank characters",
			columnID:    "column-123",
			title:       "Task",
			description: "Description",
			position:    "abc!@#",
			assigneeID:  nil,
			wantErr:     ErrInvalidLexorank,
		},
		{
			name:        "uppercase letters in lexorank",
			columnID:    "column-123",
			title:       "Task",
			description: "Description",
			position:    "ABC",
			assigneeID:  nil,
			wantErr:     ErrInvalidLexorank,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card, err := NewCard(tt.columnID, tt.title, tt.description, tt.position, tt.assigneeID, "test-creator")

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewCard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if card != nil {
					t.Errorf("NewCard() returned card when error expected")
				}
				return
			}

			// Проверяем корректность созданной карточки
			if card == nil {
				t.Fatal("NewCard() returned nil card")
			}

			if card.ID == "" {
				t.Error("NewCard() ID is empty")
			}

			if card.ColumnID != tt.columnID {
				t.Errorf("NewCard() ColumnID = %v, want %v", card.ColumnID, tt.columnID)
			}

			if card.Title != tt.title {
				t.Errorf("NewCard() Title = %v, want %v", card.Title, tt.title)
			}

			if card.Description != tt.description {
				t.Errorf("NewCard() Description = %v, want %v", card.Description, tt.description)
			}

			if card.Position != tt.position {
				t.Errorf("NewCard() Position = %v, want %v", card.Position, tt.position)
			}

			// Проверяем assignee
			if tt.assigneeID == nil {
				if card.AssigneeID != nil {
					t.Errorf("NewCard() AssigneeID = %v, want nil", *card.AssigneeID)
				}
			} else {
				if card.AssigneeID == nil {
					t.Error("NewCard() AssigneeID is nil, want non-nil")
				} else if *card.AssigneeID != *tt.assigneeID {
					t.Errorf("NewCard() AssigneeID = %v, want %v", *card.AssigneeID, *tt.assigneeID)
				}
			}

			if card.CreatedAt.IsZero() {
				t.Error("NewCard() CreatedAt is zero")
			}

			if card.UpdatedAt.IsZero() {
				t.Error("NewCard() UpdatedAt is zero")
			}

			// CreatedAt и UpdatedAt должны быть примерно одинаковыми
			if card.UpdatedAt.Sub(card.CreatedAt) > time.Second {
				t.Error("NewCard() CreatedAt and UpdatedAt differ too much")
			}
		})
	}
}

func TestCard_Update(t *testing.T) {
	assignee1 := "user-123"
	assignee2 := "user-456"

	tests := []struct {
		name        string
		title       string
		description string
		assigneeID  *string
		wantErr     error
	}{
		{
			name:        "valid update with assignee",
			title:       "Updated Task",
			description: "Updated description",
			assigneeID:  &assignee1,
			wantErr:     nil,
		},
		{
			name:        "valid update without assignee",
			title:       "Updated Task",
			description: "Updated description",
			assigneeID:  nil,
			wantErr:     nil,
		},
		{
			name:        "change assignee",
			title:       "Task",
			description: "Description",
			assigneeID:  &assignee2,
			wantErr:     nil,
		},
		{
			name:        "remove assignee",
			title:       "Task",
			description: "Description",
			assigneeID:  nil,
			wantErr:     nil,
		},
		{
			name:        "empty title",
			title:       "",
			description: "Description",
			assigneeID:  &assignee1,
			wantErr:     ErrEmptyCardTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую карточку
			originalAssignee := "original-user"
			card, err := NewCard("column-123", "Original Task", "Original description", "n", &originalAssignee, "test-creator")
			if err != nil {
				t.Fatalf("Failed to create test card: %v", err)
			}

			originalUpdatedAt := card.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			// Обновляем карточку
			err = card.Update(tt.title, tt.description, tt.assigneeID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Card.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке поля не должны измениться
				if card.Title != "Original Task" {
					t.Error("Card.Update() changed Title on error")
				}
				return
			}

			// Проверяем успешное обновление
			if card.Title != tt.title {
				t.Errorf("Card.Update() Title = %v, want %v", card.Title, tt.title)
			}

			if card.Description != tt.description {
				t.Errorf("Card.Update() Description = %v, want %v", card.Description, tt.description)
			}

			// Проверяем assignee
			if tt.assigneeID == nil {
				if card.AssigneeID != nil {
					t.Errorf("Card.Update() AssigneeID = %v, want nil", *card.AssigneeID)
				}
			} else {
				if card.AssigneeID == nil {
					t.Error("Card.Update() AssigneeID is nil, want non-nil")
				} else if *card.AssigneeID != *tt.assigneeID {
					t.Errorf("Card.Update() AssigneeID = %v, want %v", *card.AssigneeID, *tt.assigneeID)
				}
			}

			// UpdatedAt должна обновиться
			if !card.UpdatedAt.After(originalUpdatedAt) {
				t.Error("Card.Update() did not update UpdatedAt")
			}

			// Другие поля не должны измениться
			if card.ColumnID != "column-123" {
				t.Error("Card.Update() changed ColumnID")
			}

			if card.Position != "n" {
				t.Error("Card.Update() changed Position")
			}
		})
	}
}

func TestCard_Move(t *testing.T) {
	tests := []struct {
		name            string
		targetColumnID  string
		newPosition     string
		wantErr         error
	}{
		{
			name:           "valid move to different column",
			targetColumnID: "column-456",
			newPosition:    "aaa",
			wantErr:        nil,
		},
		{
			name:           "move to same column",
			targetColumnID: "column-123",
			newPosition:    "zzz",
			wantErr:        nil,
		},
		{
			name:           "empty target column ID",
			targetColumnID: "",
			newPosition:    "n",
			wantErr:        ErrColumnNotFound,
		},
		{
			name:           "empty position",
			targetColumnID: "column-456",
			newPosition:    "",
			wantErr:        ErrInvalidLexorank,
		},
		{
			name:           "invalid lexorank",
			targetColumnID: "column-456",
			newPosition:    "ABC",
			wantErr:        ErrInvalidLexorank,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую карточку
			card, err := NewCard("column-123", "Test Task", "Description", "n", nil, "test-creator")
			if err != nil {
				t.Fatalf("Failed to create test card: %v", err)
			}

			originalColumnID := card.ColumnID
			originalPosition := card.Position
			originalUpdatedAt := card.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			// Перемещаем карточку
			err = card.Move(tt.targetColumnID, tt.newPosition)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Card.Move() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке поля не должны измениться
				if card.ColumnID != originalColumnID {
					t.Error("Card.Move() changed ColumnID on error")
				}
				if card.Position != originalPosition {
					t.Error("Card.Move() changed Position on error")
				}
				return
			}

			// Проверяем успешное перемещение
			if card.ColumnID != tt.targetColumnID {
				t.Errorf("Card.Move() ColumnID = %v, want %v", card.ColumnID, tt.targetColumnID)
			}

			if card.Position != tt.newPosition {
				t.Errorf("Card.Move() Position = %v, want %v", card.Position, tt.newPosition)
			}

			// UpdatedAt должна обновиться
			if !card.UpdatedAt.After(originalUpdatedAt) {
				t.Error("Card.Move() did not update UpdatedAt")
			}

			// Другие поля не должны измениться
			if card.Title != "Test Task" {
				t.Error("Card.Move() changed Title")
			}
		})
	}
}

func TestCard_Reorder(t *testing.T) {
	tests := []struct {
		name        string
		newPosition string
		wantErr     error
	}{
		{
			name:        "valid reorder",
			newPosition: "aaa",
			wantErr:     nil,
		},
		{
			name:        "reorder to same position",
			newPosition: "n",
			wantErr:     nil,
		},
		{
			name:        "reorder with complex lexorank",
			newPosition: "abc123xyz",
			wantErr:     nil,
		},
		{
			name:        "empty position",
			newPosition: "",
			wantErr:     ErrInvalidLexorank,
		},
		{
			name:        "invalid lexorank with uppercase",
			newPosition: "ABC",
			wantErr:     ErrInvalidLexorank,
		},
		{
			name:        "invalid lexorank with special chars",
			newPosition: "a-b-c",
			wantErr:     ErrInvalidLexorank,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую карточку
			card, err := NewCard("column-123", "Test Task", "Description", "n", nil, "test-creator")
			if err != nil {
				t.Fatalf("Failed to create test card: %v", err)
			}

			originalPosition := card.Position
			originalColumnID := card.ColumnID
			originalUpdatedAt := card.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			// Изменяем позицию
			err = card.Reorder(tt.newPosition)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Card.Reorder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке position не должна измениться
				if card.Position != originalPosition {
					t.Error("Card.Reorder() changed Position on error")
				}
				return
			}

			// Проверяем успешное изменение позиции
			if card.Position != tt.newPosition {
				t.Errorf("Card.Reorder() Position = %v, want %v", card.Position, tt.newPosition)
			}

			// UpdatedAt должна обновиться
			if !card.UpdatedAt.After(originalUpdatedAt) {
				t.Error("Card.Reorder() did not update UpdatedAt")
			}

			// ColumnID не должна измениться (в отличие от Move)
			if card.ColumnID != originalColumnID {
				t.Error("Card.Reorder() changed ColumnID")
			}

			// Другие поля не должны измениться
			if card.Title != "Test Task" {
				t.Error("Card.Reorder() changed Title")
			}
		})
	}
}

func TestCard_LexorankValidation(t *testing.T) {
	// Тест различных валидных и невалидных lexorank позиций
	validPositions := []string{"a", "n", "z", "0", "9", "abc", "xyz", "000", "999", "a0b1c2"}
	invalidPositions := []string{"", "A", "Z", "abc!", "xyz@", "a b", "a-b", "абв"}

	for _, pos := range validPositions {
		t.Run("valid_"+pos, func(t *testing.T) {
			card, err := NewCard("column-123", "Task", "Desc", pos, nil, "test-creator")
			if err != nil {
				t.Errorf("NewCard() with valid position %q returned error: %v", pos, err)
			}
			if card == nil {
				t.Error("NewCard() returned nil card for valid position")
			}
		})
	}

	for _, pos := range invalidPositions {
		t.Run("invalid_"+pos, func(t *testing.T) {
			card, err := NewCard("column-123", "Task", "Desc", pos, nil, "test-creator")
			if !errors.Is(err, ErrInvalidLexorank) {
				t.Errorf("NewCard() with invalid position %q error = %v, want ErrInvalidLexorank", pos, err)
			}
			if card != nil {
				t.Error("NewCard() returned non-nil card for invalid position")
			}
		})
	}
}
