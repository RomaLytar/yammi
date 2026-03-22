package domain

import (
	"errors"
	"testing"
)

func TestNewAttachment(t *testing.T) {
	tests := []struct {
		name       string
		cardID     string
		boardID    string
		fileName   string
		fileSize   int64
		mimeType   string
		uploaderID string
		wantErr    error
	}{
		{
			name:       "valid attachment",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "document.pdf",
			fileSize:   1024,
			mimeType:   "application/pdf",
			uploaderID: "user-789",
			wantErr:    nil,
		},
		{
			name:       "empty card_id",
			cardID:     "",
			boardID:    "board-456",
			fileName:   "document.pdf",
			fileSize:   1024,
			mimeType:   "application/pdf",
			uploaderID: "user-789",
			wantErr:    ErrCardNotFound,
		},
		{
			name:       "empty board_id",
			cardID:     "card-123",
			boardID:    "",
			fileName:   "document.pdf",
			fileSize:   1024,
			mimeType:   "application/pdf",
			uploaderID: "user-789",
			wantErr:    ErrBoardNotFound,
		},
		{
			name:       "empty filename",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "",
			fileSize:   1024,
			mimeType:   "application/pdf",
			uploaderID: "user-789",
			wantErr:    ErrEmptyFileName,
		},
		{
			name:       "file too large",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "huge.zip",
			fileSize:   MaxFileSize + 1,
			mimeType:   "application/zip",
			uploaderID: "user-789",
			wantErr:    ErrFileTooLarge,
		},
		{
			name:       "at max size",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "max.bin",
			fileSize:   MaxFileSize,
			mimeType:   "application/octet-stream",
			uploaderID: "user-789",
			wantErr:    nil,
		},
		{
			name:       "zero file size",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "empty.txt",
			fileSize:   0,
			mimeType:   "text/plain",
			uploaderID: "user-789",
			wantErr:    ErrFileTooLarge,
		},
		{
			name:       "negative file size",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "bad.txt",
			fileSize:   -1,
			mimeType:   "text/plain",
			uploaderID: "user-789",
			wantErr:    ErrFileTooLarge,
		},
		{
			name:       "empty uploader_id",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "file.txt",
			fileSize:   100,
			mimeType:   "text/plain",
			uploaderID: "",
			wantErr:    ErrAccessDenied,
		},
		{
			name:       "path traversal in filename",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "../../../etc/passwd",
			fileSize:   100,
			mimeType:   "text/plain",
			uploaderID: "user-789",
			wantErr:    nil, // sanitizeFileName strips path
		},
		{
			name:       "hidden file dots only",
			cardID:     "card-123",
			boardID:    "board-456",
			fileName:   "...",
			fileSize:   100,
			mimeType:   "text/plain",
			uploaderID: "user-789",
			wantErr:    ErrEmptyFileName, // sanitizeFileName returns empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			att, err := NewAttachment(tt.cardID, tt.boardID, tt.fileName, tt.fileSize, tt.mimeType, tt.uploaderID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewAttachment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr != nil {
				if att != nil {
					t.Errorf("NewAttachment() returned attachment when error expected")
				}
				return
			}

			// Проверяем корректность созданного вложения
			if att == nil {
				t.Fatal("NewAttachment() returned nil attachment")
			}

			if att.ID == "" {
				t.Error("NewAttachment() ID is empty")
			}

			if att.CardID != tt.cardID {
				t.Errorf("NewAttachment() CardID = %v, want %v", att.CardID, tt.cardID)
			}

			if att.BoardID != tt.boardID {
				t.Errorf("NewAttachment() BoardID = %v, want %v", att.BoardID, tt.boardID)
			}

			if att.FileSize != tt.fileSize {
				t.Errorf("NewAttachment() FileSize = %v, want %v", att.FileSize, tt.fileSize)
			}

			if att.MimeType != tt.mimeType {
				t.Errorf("NewAttachment() MimeType = %v, want %v", att.MimeType, tt.mimeType)
			}

			if att.UploaderID != tt.uploaderID {
				t.Errorf("NewAttachment() UploaderID = %v, want %v", att.UploaderID, tt.uploaderID)
			}

			if att.StorageKey == "" {
				t.Error("NewAttachment() StorageKey is empty")
			}

			if att.CreatedAt.IsZero() {
				t.Error("NewAttachment() CreatedAt is zero")
			}
		})
	}
}
