package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/RomaLytar/yammi/services/comment/internal/domain"
)

type CommentRepository struct {
	db *sql.DB
}

func NewCommentRepository(db *sql.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create создает новый комментарий
func (r *CommentRepository) Create(ctx context.Context, comment *domain.Comment) error {
	query := `
		INSERT INTO comments (id, card_id, board_id, author_id, parent_id, content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.ExecContext(ctx, query,
		comment.ID, comment.CardID, comment.BoardID, comment.AuthorID,
		comment.ParentID, comment.Content, comment.CreatedAt, comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}

	return nil
}

// GetByID возвращает комментарий по ID
func (r *CommentRepository) GetByID(ctx context.Context, commentID string) (*domain.Comment, error) {
	query := `
		SELECT id, card_id, board_id, author_id, parent_id, content,
		       COALESCE((SELECT COUNT(*) FROM comments WHERE parent_id = c.id), 0) AS reply_count,
		       created_at, updated_at
		FROM comments c
		WHERE id = $1
	`

	var comment domain.Comment
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, query, commentID).Scan(
		&comment.ID, &comment.CardID, &comment.BoardID, &comment.AuthorID,
		&parentID, &comment.Content, &comment.ReplyCount,
		&comment.CreatedAt, &comment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCommentNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select comment: %w", err)
	}

	if parentID.Valid {
		comment.ParentID = &parentID.String
	}

	return &comment, nil
}

// ListByCardID возвращает комментарии карточки с курсорной пагинацией
// Cursor формат: "2024-01-01T00:00:00Z|uuid"
func (r *CommentRepository) ListByCardID(ctx context.Context, cardID string, limit int, cursor string) ([]*domain.Comment, string, error) {
	var rows *sql.Rows
	var err error

	if cursor != "" {
		// Парсим курсор: created_at|id
		parts := strings.SplitN(cursor, "|", 2)
		if len(parts) != 2 {
			return nil, "", fmt.Errorf("invalid cursor format")
		}
		cursorTime, parseErr := time.Parse(time.RFC3339Nano, parts[0])
		if parseErr != nil {
			return nil, "", fmt.Errorf("invalid cursor time: %w", parseErr)
		}
		cursorID := parts[1]

		query := `
			SELECT id, card_id, board_id, author_id, parent_id, content,
			       COALESCE((SELECT COUNT(*) FROM comments WHERE parent_id = c.id), 0) AS reply_count,
			       created_at, updated_at
			FROM comments c
			WHERE card_id = $1 AND (created_at, id) > ($2, $3)
			ORDER BY created_at ASC, id ASC
			LIMIT $4
		`
		rows, err = r.db.QueryContext(ctx, query, cardID, cursorTime, cursorID, limit+1)
	} else {
		query := `
			SELECT id, card_id, board_id, author_id, parent_id, content,
			       COALESCE((SELECT COUNT(*) FROM comments WHERE parent_id = c.id), 0) AS reply_count,
			       created_at, updated_at
			FROM comments c
			WHERE card_id = $1
			ORDER BY created_at ASC, id ASC
			LIMIT $2
		`
		rows, err = r.db.QueryContext(ctx, query, cardID, limit+1)
	}

	if err != nil {
		return nil, "", fmt.Errorf("select comments: %w", err)
	}
	defer rows.Close()

	var comments []*domain.Comment
	for rows.Next() {
		var c domain.Comment
		var parentID sql.NullString

		if err := rows.Scan(&c.ID, &c.CardID, &c.BoardID, &c.AuthorID, &parentID, &c.Content, &c.ReplyCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, "", fmt.Errorf("scan comment: %w", err)
		}

		if parentID.Valid {
			c.ParentID = &parentID.String
		}

		comments = append(comments, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	// Определяем следующий курсор
	var nextCursor string
	if len(comments) > limit {
		comments = comments[:limit]
		last := comments[len(comments)-1]
		nextCursor = last.CreatedAt.Format(time.RFC3339Nano) + "|" + last.ID
	}

	return comments, nextCursor, nil
}

// Update обновляет комментарий
func (r *CommentRepository) Update(ctx context.Context, comment *domain.Comment) error {
	query := `
		UPDATE comments
		SET content = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID)
	if err != nil {
		return fmt.Errorf("update comment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}

// Delete удаляет комментарий по ID
func (r *CommentRepository) Delete(ctx context.Context, commentID string) error {
	query := `DELETE FROM comments WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, commentID)
	if err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCommentNotFound
	}

	return nil
}

// CountByCardID возвращает количество комментариев к карточке
func (r *CommentRepository) CountByCardID(ctx context.Context, cardID string) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE card_id = $1`

	var count int
	if err := r.db.QueryRowContext(ctx, query, cardID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count comments: %w", err)
	}

	return count, nil
}

// IncrementReplyCount увеличивает счётчик ответов у родительского комментария
func (r *CommentRepository) IncrementReplyCount(ctx context.Context, commentID string) error {
	// reply_count вычисляется динамически в GetByID/ListByCardID через подзапрос,
	// но для оптимизации можно хранить материализованное значение.
	// В текущей реализации используем подзапрос — этот метод no-op.
	return nil
}

// DecrementReplyCount уменьшает счётчик ответов у родительского комментария
func (r *CommentRepository) DecrementReplyCount(ctx context.Context, commentID string) error {
	// reply_count вычисляется динамически — этот метод no-op.
	return nil
}
