package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type AttachmentRepository struct {
	db *sql.DB
}

func NewAttachmentRepository(db *sql.DB) *AttachmentRepository {
	return &AttachmentRepository{db: db}
}

// Create создает запись о вложении
func (r *AttachmentRepository) Create(ctx context.Context, attachment *domain.Attachment) error {
	query := `
		INSERT INTO attachments (id, card_id, board_id, file_name, file_size, mime_type, storage_key, uploader_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(ctx, query,
		attachment.ID, attachment.CardID, attachment.BoardID,
		attachment.FileName, attachment.FileSize, attachment.MimeType,
		attachment.StorageKey, attachment.UploaderID, attachment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert attachment: %w", err)
	}

	return nil
}

// GetByID возвращает вложение по ID с проверкой принадлежности к доске (IDOR protection)
func (r *AttachmentRepository) GetByID(ctx context.Context, attachmentID, boardID string) (*domain.Attachment, error) {
	query := `
		SELECT id, card_id, board_id, file_name, file_size, mime_type, storage_key, uploader_id, created_at
		FROM attachments
		WHERE id = $1 AND board_id = $2
	`

	var att domain.Attachment
	err := r.db.QueryRowContext(ctx, query, attachmentID, boardID).Scan(
		&att.ID, &att.CardID, &att.BoardID,
		&att.FileName, &att.FileSize, &att.MimeType,
		&att.StorageKey, &att.UploaderID, &att.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrAttachmentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select attachment: %w", err)
	}

	return &att, nil
}

// ListByCardID возвращает все вложения карточки
func (r *AttachmentRepository) ListByCardID(ctx context.Context, cardID, boardID string) ([]*domain.Attachment, error) {
	query := `
		SELECT id, card_id, board_id, file_name, file_size, mime_type, storage_key, uploader_id, created_at
		FROM attachments
		WHERE card_id = $1 AND board_id = $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select attachments: %w", err)
	}
	defer rows.Close()

	var attachments []*domain.Attachment
	for rows.Next() {
		var att domain.Attachment
		if err := rows.Scan(
			&att.ID, &att.CardID, &att.BoardID,
			&att.FileName, &att.FileSize, &att.MimeType,
			&att.StorageKey, &att.UploaderID, &att.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan attachment: %w", err)
		}
		attachments = append(attachments, &att)
	}

	return attachments, rows.Err()
}

// Delete удаляет вложение по ID
func (r *AttachmentRepository) Delete(ctx context.Context, attachmentID, boardID string) error {
	query := `DELETE FROM attachments WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, attachmentID, boardID)
	if err != nil {
		return fmt.Errorf("delete attachment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrAttachmentNotFound
	}

	return nil
}

// CountByCardID возвращает количество вложений карточки
func (r *AttachmentRepository) CountByCardID(ctx context.Context, cardID, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM attachments WHERE card_id = $1 AND board_id = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, cardID, boardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count attachments: %w", err)
	}

	return count, nil
}
