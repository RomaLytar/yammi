package domain

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	// MaxFileSize — максимальный размер файла (50 МБ)
	MaxFileSize = 50 * 1024 * 1024
	// MaxAttachmentsPerCard — максимальное количество вложений на карточку
	MaxAttachmentsPerCard = 20
)

// Attachment — вложение к карточке
type Attachment struct {
	ID         string
	CardID     string
	BoardID    string
	FileName   string
	FileSize   int64
	MimeType   string
	StorageKey string // MinIO object key: boards/{boardID}/cards/{cardID}/{id}/{filename}
	UploaderID string
	CreatedAt  time.Time
}

// NewAttachment создает новое вложение с валидацией
func NewAttachment(cardID, boardID, fileName string, fileSize int64, mimeType, uploaderID string) (*Attachment, error) {
	if cardID == "" {
		return nil, ErrCardNotFound
	}
	if boardID == "" {
		return nil, ErrBoardNotFound
	}
	if fileName == "" {
		return nil, ErrEmptyFileName
	}
	if fileSize <= 0 {
		return nil, ErrFileTooLarge
	}
	if fileSize > MaxFileSize {
		return nil, ErrFileTooLarge
	}
	if uploaderID == "" {
		return nil, ErrAccessDenied
	}

	// Санитизация имени файла: оставляем только базовое имя
	fileName = sanitizeFileName(fileName)
	if fileName == "" {
		return nil, ErrEmptyFileName
	}

	id := uuid.NewString()
	storageKey := "boards/" + boardID + "/cards/" + cardID + "/" + id + "/" + fileName

	return &Attachment{
		ID:         id,
		CardID:     cardID,
		BoardID:    boardID,
		FileName:   fileName,
		FileSize:   fileSize,
		MimeType:   mimeType,
		StorageKey: storageKey,
		UploaderID: uploaderID,
		CreatedAt:  time.Now(),
	}, nil
}

// sanitizeFileName очищает имя файла от путей и опасных символов
func sanitizeFileName(name string) string {
	// Берем только базовое имя (убираем пути)
	name = filepath.Base(name)

	// Убираем нулевые байты и управляющие символы
	var b strings.Builder
	for _, r := range name {
		if r > 31 && r != 127 {
			b.WriteRune(r)
		}
	}
	name = b.String()

	// Убираем точки в начале (скрытые файлы / path traversal)
	name = strings.TrimLeft(name, ".")

	if name == "" || name == "." || name == ".." {
		return ""
	}

	return name
}
