package domain

import (
	"testing"
	"time"
)

func TestNewBoard(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		ownerID     string
		wantErr     error
	}{
		{
			name:        "valid board",
			title:       "My Board",
			description: "Board description",
			ownerID:     "user-123",
			wantErr:     nil,
		},
		{
			name:        "valid board without description",
			title:       "Simple Board",
			description: "",
			ownerID:     "user-456",
			wantErr:     nil,
		},
		{
			name:        "empty title",
			title:       "",
			description: "Description",
			ownerID:     "user-123",
			wantErr:     ErrEmptyTitle,
		},
		{
			name:        "empty owner ID",
			title:       "Board Title",
			description: "Description",
			ownerID:     "",
			wantErr:     ErrEmptyOwnerID,
		},
		{
			name:        "both empty title and owner ID",
			title:       "",
			description: "Description",
			ownerID:     "",
			wantErr:     ErrEmptyTitle, // title проверяется первым
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			board, err := NewBoard(tt.title, tt.description, tt.ownerID)

			if err != tt.wantErr {
				t.Errorf("NewBoard() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if board != nil {
					t.Errorf("NewBoard() returned board when error expected")
				}
				return
			}

			// Проверяем корректность созданной доски
			if board == nil {
				t.Fatal("NewBoard() returned nil board")
			}

			if board.ID == "" {
				t.Error("NewBoard() ID is empty")
			}

			if board.Title != tt.title {
				t.Errorf("NewBoard() Title = %v, want %v", board.Title, tt.title)
			}

			if board.Description != tt.description {
				t.Errorf("NewBoard() Description = %v, want %v", board.Description, tt.description)
			}

			if board.OwnerID != tt.ownerID {
				t.Errorf("NewBoard() OwnerID = %v, want %v", board.OwnerID, tt.ownerID)
			}

			if board.Version != 1 {
				t.Errorf("NewBoard() Version = %v, want 1", board.Version)
			}

			if board.CreatedAt.IsZero() {
				t.Error("NewBoard() CreatedAt is zero")
			}

			if board.UpdatedAt.IsZero() {
				t.Error("NewBoard() UpdatedAt is zero")
			}

			// CreatedAt и UpdatedAt должны быть примерно одинаковыми
			if board.UpdatedAt.Sub(board.CreatedAt) > time.Second {
				t.Error("NewBoard() CreatedAt and UpdatedAt differ too much")
			}
		})
	}
}

func TestBoard_Update(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		wantErr     error
	}{
		{
			name:        "valid update with description",
			title:       "Updated Title",
			description: "Updated description",
			wantErr:     nil,
		},
		{
			name:        "valid update without description",
			title:       "Updated Title",
			description: "",
			wantErr:     nil,
		},
		{
			name:        "empty title",
			title:       "",
			description: "Description",
			wantErr:     ErrEmptyTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую доску
			board, err := NewBoard("Original Title", "Original description", "user-123")
			if err != nil {
				t.Fatalf("Failed to create test board: %v", err)
			}

			originalVersion := board.Version
			originalUpdatedAt := board.UpdatedAt
			time.Sleep(10 * time.Millisecond) // небольшая задержка для проверки UpdatedAt

			// Обновляем доску
			err = board.Update(tt.title, tt.description)

			if err != tt.wantErr {
				t.Errorf("Board.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке поля не должны измениться
				if board.Title != "Original Title" {
					t.Error("Board.Update() changed Title on error")
				}
				if board.Description != "Original description" {
					t.Error("Board.Update() changed Description on error")
				}
				if board.Version != originalVersion {
					t.Error("Board.Update() changed Version on error")
				}
				return
			}

			// Проверяем успешное обновление
			if board.Title != tt.title {
				t.Errorf("Board.Update() Title = %v, want %v", board.Title, tt.title)
			}

			if board.Description != tt.description {
				t.Errorf("Board.Update() Description = %v, want %v", board.Description, tt.description)
			}

			// Version должна увеличиться
			if board.Version != originalVersion+1 {
				t.Errorf("Board.Update() Version = %v, want %v", board.Version, originalVersion+1)
			}

			// UpdatedAt должна обновиться
			if !board.UpdatedAt.After(originalUpdatedAt) {
				t.Error("Board.Update() did not update UpdatedAt")
			}
		})
	}
}

func TestBoard_IsOwner(t *testing.T) {
	board, err := NewBoard("Test Board", "Description", "owner-123")
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	tests := []struct {
		name   string
		userID string
		want   bool
	}{
		{
			name:   "is owner",
			userID: "owner-123",
			want:   true,
		},
		{
			name:   "not owner",
			userID: "user-456",
			want:   false,
		},
		{
			name:   "empty user ID",
			userID: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := board.IsOwner(tt.userID)
			if got != tt.want {
				t.Errorf("Board.IsOwner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBoard_IncrementVersion(t *testing.T) {
	board, err := NewBoard("Test Board", "Description", "owner-123")
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	originalVersion := board.Version
	originalUpdatedAt := board.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	board.IncrementVersion()

	if board.Version != originalVersion+1 {
		t.Errorf("Board.IncrementVersion() Version = %v, want %v", board.Version, originalVersion+1)
	}

	if !board.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Board.IncrementVersion() did not update UpdatedAt")
	}

	// Проверяем множественные инкременты
	board.IncrementVersion()
	board.IncrementVersion()

	if board.Version != originalVersion+3 {
		t.Errorf("Board.IncrementVersion() (multiple) Version = %v, want %v", board.Version, originalVersion+3)
	}
}
