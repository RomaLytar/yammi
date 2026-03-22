package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNewComment(t *testing.T) {
	parentID := "parent-123"

	tests := []struct {
		name     string
		cardID   string
		boardID  string
		authorID string
		content  string
		parentID *string
		wantErr  error
	}{
		{
			name:     "valid comment without parent",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "user-123",
			content:  "Hello, world!",
			parentID: nil,
			wantErr:  nil,
		},
		{
			name:     "valid comment with parent",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "user-456",
			content:  "This is a reply",
			parentID: &parentID,
			wantErr:  nil,
		},
		{
			name:     "empty card_id",
			cardID:   "",
			boardID:  "board-123",
			authorID: "user-123",
			content:  "Hello",
			parentID: nil,
			wantErr:  ErrEmptyCardID,
		},
		{
			name:     "empty board_id",
			cardID:   "card-123",
			boardID:  "",
			authorID: "user-123",
			content:  "Hello",
			parentID: nil,
			wantErr:  ErrEmptyBoardID,
		},
		{
			name:     "empty author_id",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "",
			content:  "Hello",
			parentID: nil,
			wantErr:  ErrEmptyAuthorID,
		},
		{
			name:     "empty text",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "user-123",
			content:  "",
			parentID: nil,
			wantErr:  ErrEmptyText,
		},
		{
			name:     "content too long",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "user-123",
			content:  strings.Repeat("a", 10001),
			parentID: nil,
			wantErr:  ErrContentTooLong,
		},
		{
			name:     "content at max length",
			cardID:   "card-123",
			boardID:  "board-123",
			authorID: "user-123",
			content:  strings.Repeat("a", 10000),
			parentID: nil,
			wantErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comment, err := NewComment(tt.cardID, tt.boardID, tt.authorID, tt.content, tt.parentID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if comment != nil {
					t.Errorf("NewComment() returned comment when error expected")
				}
				return
			}

			// Проверяем корректность созданного комментария
			if comment == nil {
				t.Fatal("NewComment() returned nil comment")
			}

			if comment.ID == "" {
				t.Error("NewComment() ID is empty")
			}

			if comment.CardID != tt.cardID {
				t.Errorf("NewComment() CardID = %v, want %v", comment.CardID, tt.cardID)
			}

			if comment.BoardID != tt.boardID {
				t.Errorf("NewComment() BoardID = %v, want %v", comment.BoardID, tt.boardID)
			}

			if comment.AuthorID != tt.authorID {
				t.Errorf("NewComment() AuthorID = %v, want %v", comment.AuthorID, tt.authorID)
			}

			if comment.Content != tt.content {
				t.Errorf("NewComment() Content length = %v, want %v", len(comment.Content), len(tt.content))
			}

			// Проверяем parent_id
			if tt.parentID == nil {
				if comment.ParentID != nil {
					t.Errorf("NewComment() ParentID = %v, want nil", *comment.ParentID)
				}
			} else {
				if comment.ParentID == nil {
					t.Error("NewComment() ParentID is nil, want non-nil")
				} else if *comment.ParentID != *tt.parentID {
					t.Errorf("NewComment() ParentID = %v, want %v", *comment.ParentID, *tt.parentID)
				}
			}

			if comment.CreatedAt.IsZero() {
				t.Error("NewComment() CreatedAt is zero")
			}

			if comment.UpdatedAt.IsZero() {
				t.Error("NewComment() UpdatedAt is zero")
			}

			// CreatedAt и UpdatedAt должны быть примерно одинаковыми
			if comment.UpdatedAt.Sub(comment.CreatedAt) > time.Second {
				t.Error("NewComment() CreatedAt and UpdatedAt differ too much")
			}
		})
	}
}

func TestComment_Update(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr error
	}{
		{
			name:    "valid update",
			content: "Updated comment text",
			wantErr: nil,
		},
		{
			name:    "empty text",
			content: "",
			wantErr: ErrEmptyText,
		},
		{
			name:    "content too long",
			content: strings.Repeat("b", 10001),
			wantErr: ErrContentTooLong,
		},
		{
			name:    "content at max length",
			content: strings.Repeat("b", 10000),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый комментарий
			comment, err := NewComment("card-123", "board-123", "user-123", "Original text", nil)
			if err != nil {
				t.Fatalf("Failed to create test comment: %v", err)
			}

			originalUpdatedAt := comment.UpdatedAt
			time.Sleep(10 * time.Millisecond)

			// Обновляем комментарий
			err = comment.Update(tt.content)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Comment.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				// При ошибке поля не должны измениться
				if comment.Content != "Original text" {
					t.Error("Comment.Update() changed Content on error")
				}
				return
			}

			// Проверяем успешное обновление
			if comment.Content != tt.content {
				t.Errorf("Comment.Update() Content length = %v, want %v", len(comment.Content), len(tt.content))
			}

			// UpdatedAt должна обновиться
			if !comment.UpdatedAt.After(originalUpdatedAt) {
				t.Error("Comment.Update() did not update UpdatedAt")
			}

			// Другие поля не должны измениться
			if comment.CardID != "card-123" {
				t.Error("Comment.Update() changed CardID")
			}

			if comment.BoardID != "board-123" {
				t.Error("Comment.Update() changed BoardID")
			}

			if comment.AuthorID != "user-123" {
				t.Error("Comment.Update() changed AuthorID")
			}
		})
	}
}

func TestComment_IsReply(t *testing.T) {
	// Корневой комментарий
	rootComment, err := NewComment("card-123", "board-123", "user-123", "Root comment", nil)
	if err != nil {
		t.Fatalf("Failed to create root comment: %v", err)
	}

	if rootComment.IsReply() {
		t.Error("Root comment should not be a reply")
	}

	// Ответ на комментарий
	parentID := "parent-123"
	replyComment, err := NewComment("card-123", "board-123", "user-456", "Reply comment", &parentID)
	if err != nil {
		t.Fatalf("Failed to create reply comment: %v", err)
	}

	if !replyComment.IsReply() {
		t.Error("Reply comment should be a reply")
	}
}

func TestComment_ValidateReplyDepth(t *testing.T) {
	// Тестируем создание комментария с parent_id — это валидно
	parentID := "some-parent-id"
	comment, err := NewComment("card-123", "board-123", "user-123", "Reply", &parentID)
	if err != nil {
		t.Fatalf("NewComment() with parent_id should succeed: %v", err)
	}

	if !comment.IsReply() {
		t.Error("Comment with parent_id should be a reply")
	}

	// Проверка глубины вложенности (replies to replies) выполняется в usecase,
	// а не в domain — domain не имеет доступа к репозиторию.
	// Тест документирует это поведение.
	if comment.ParentID == nil {
		t.Error("Comment ParentID should not be nil")
	}
	if *comment.ParentID != parentID {
		t.Errorf("Comment ParentID = %v, want %v", *comment.ParentID, parentID)
	}
}
