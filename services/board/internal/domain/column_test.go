package domain

import (
	"testing"
	"time"
)

func TestNewColumn(t *testing.T) {
	tests := []struct {
		name     string
		boardID  string
		title    string
		position int
		wantErr  error
	}{
		{
			name:     "valid column at position 0",
			boardID:  "board-123",
			title:    "To Do",
			position: 0,
			wantErr:  nil,
		},
		{
			name:     "valid column at position 5",
			boardID:  "board-123",
			title:    "In Progress",
			position: 5,
			wantErr:  nil,
		},
		{
			name:     "empty board ID",
			boardID:  "",
			title:    "To Do",
			position: 0,
			wantErr:  ErrBoardNotFound,
		},
		{
			name:     "empty title",
			boardID:  "board-123",
			title:    "",
			position: 0,
			wantErr:  ErrEmptyColumnTitle,
		},
		{
			name:     "negative position",
			boardID:  "board-123",
			title:    "To Do",
			position: -1,
			wantErr:  ErrInvalidPosition,
		},
		{
			name:     "multiple validation errors",
			boardID:  "",
			title:    "",
			position: -1,
			wantErr:  ErrBoardNotFound, // boardID проверяется первым
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column, err := NewColumn(tt.boardID, tt.title, tt.position)

			if err != tt.wantErr {
				t.Errorf("NewColumn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if column != nil {
					t.Errorf("NewColumn() returned column when error expected")
				}
				return
			}

			// Проверяем корректность созданной колонки
			if column == nil {
				t.Fatal("NewColumn() returned nil column")
			}

			if column.ID == "" {
				t.Error("NewColumn() ID is empty")
			}

			if column.BoardID != tt.boardID {
				t.Errorf("NewColumn() BoardID = %v, want %v", column.BoardID, tt.boardID)
			}

			if column.Title != tt.title {
				t.Errorf("NewColumn() Title = %v, want %v", column.Title, tt.title)
			}

			if column.Position != tt.position {
				t.Errorf("NewColumn() Position = %v, want %v", column.Position, tt.position)
			}

			if column.CreatedAt.IsZero() {
				t.Error("NewColumn() CreatedAt is zero")
			}

			// CreatedAt должна быть близка к текущему времени
			if time.Since(column.CreatedAt) > time.Second {
				t.Error("NewColumn() CreatedAt is too far in the past")
			}
		})
	}
}

func TestColumn_Update(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr error
	}{
		{
			name:    "valid update",
			title:   "Updated Title",
			wantErr: nil,
		},
		{
			name:    "update with same title",
			title:   "Original Title",
			wantErr: nil,
		},
		{
			name:    "empty title",
			title:   "",
			wantErr: ErrEmptyColumnTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую колонку
			column, err := NewColumn("board-123", "Original Title", 0)
			if err != nil {
				t.Fatalf("Failed to create test column: %v", err)
			}

			originalTitle := column.Title

			// Обновляем колонку
			err = column.Update(tt.title)

			if err != tt.wantErr {
				t.Errorf("Column.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке title не должен измениться
				if column.Title != originalTitle {
					t.Error("Column.Update() changed Title on error")
				}
				return
			}

			// Проверяем успешное обновление
			if column.Title != tt.title {
				t.Errorf("Column.Update() Title = %v, want %v", column.Title, tt.title)
			}

			// Другие поля не должны измениться
			if column.BoardID != "board-123" {
				t.Error("Column.Update() changed BoardID")
			}

			if column.Position != 0 {
				t.Error("Column.Update() changed Position")
			}
		})
	}
}

func TestColumn_UpdatePosition(t *testing.T) {
	tests := []struct {
		name     string
		position int
		wantErr  error
	}{
		{
			name:     "update to position 0",
			position: 0,
			wantErr:  nil,
		},
		{
			name:     "update to position 10",
			position: 10,
			wantErr:  nil,
		},
		{
			name:     "update to same position",
			position: 5,
			wantErr:  nil,
		},
		{
			name:     "negative position",
			position: -1,
			wantErr:  ErrInvalidPosition,
		},
		{
			name:     "large position",
			position: 1000,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовую колонку
			column, err := NewColumn("board-123", "Test Column", 5)
			if err != nil {
				t.Fatalf("Failed to create test column: %v", err)
			}

			originalPosition := column.Position

			// Обновляем позицию
			err = column.UpdatePosition(tt.position)

			if err != tt.wantErr {
				t.Errorf("Column.UpdatePosition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке position не должна измениться
				if column.Position != originalPosition {
					t.Error("Column.UpdatePosition() changed Position on error")
				}
				return
			}

			// Проверяем успешное обновление
			if column.Position != tt.position {
				t.Errorf("Column.UpdatePosition() Position = %v, want %v", column.Position, tt.position)
			}

			// Другие поля не должны измениться
			if column.Title != "Test Column" {
				t.Error("Column.UpdatePosition() changed Title")
			}

			if column.BoardID != "board-123" {
				t.Error("Column.UpdatePosition() changed BoardID")
			}
		})
	}
}

func TestColumn_PositionBoundaries(t *testing.T) {
	// Тест граничных значений позиции
	column, err := NewColumn("board-123", "Test Column", 0)
	if err != nil {
		t.Fatalf("Failed to create test column: %v", err)
	}

	// Проверяем максимальные значения
	err = column.UpdatePosition(int(^uint(0) >> 1)) // max int
	if err != nil {
		t.Errorf("Column.UpdatePosition() failed on max int: %v", err)
	}

	// Проверяем, что отрицательные значения отклоняются
	err = column.UpdatePosition(-100)
	if err != ErrInvalidPosition {
		t.Errorf("Column.UpdatePosition() error = %v, want ErrInvalidPosition", err)
	}
}
