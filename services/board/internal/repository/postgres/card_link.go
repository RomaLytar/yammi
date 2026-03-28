package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type CardLinkRepository struct {
	db *sql.DB
}

func NewCardLinkRepository(db *sql.DB) *CardLinkRepository {
	return &CardLinkRepository{db: db}
}

// Create создает новую связь между карточками
func (r *CardLinkRepository) Create(ctx context.Context, link *domain.CardLink) error {
	query := `
		INSERT INTO card_links (id, parent_id, child_id, board_id, link_type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(ctx, query,
		link.ID, link.ParentID, link.ChildID, link.BoardID, string(link.LinkType), link.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrLinkAlreadyExists
		}
		// Check constraint violation (self-link) — SQLSTATE 23514
		if isCheckConstraintError(err) {
			return domain.ErrSelfLink
		}
		return fmt.Errorf("insert card link: %w", err)
	}

	return nil
}

// CreateVerified создает связь, проверяя существование parent card в одном запросе
func (r *CardLinkRepository) CreateVerified(ctx context.Context, link *domain.CardLink) error {
	query := `
		INSERT INTO card_links (id, parent_id, child_id, board_id, link_type, created_at)
		SELECT $1, $2, $3, $4, $5, $6
		WHERE EXISTS (SELECT 1 FROM cards WHERE id = $2 AND board_id = $4)
	`

	result, err := r.db.ExecContext(ctx, query,
		link.ID, link.ParentID, link.ChildID, link.BoardID, string(link.LinkType), link.CreatedAt,
	)
	if err != nil {
		if isDuplicateKeyError(err) {
			return domain.ErrLinkAlreadyExists
		}
		if isCheckConstraintError(err) {
			return domain.ErrSelfLink
		}
		return fmt.Errorf("insert card link verified: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardNotFound
	}

	return nil
}

// Delete удаляет связь по ID (boardID для partition pruning)
func (r *CardLinkRepository) Delete(ctx context.Context, linkID, boardID string) error {
	query := `DELETE FROM card_links WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, linkID, boardID)
	if err != nil {
		return fmt.Errorf("delete card link: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrCardLinkNotFound
	}

	return nil
}

// GetByID возвращает связь по ID (boardID для partition pruning)
func (r *CardLinkRepository) GetByID(ctx context.Context, linkID, boardID string) (*domain.CardLink, error) {
	query := `
		SELECT id, parent_id, child_id, board_id, link_type, created_at
		FROM card_links
		WHERE id = $1 AND board_id = $2
	`

	var link domain.CardLink
	var linkType string
	err := r.db.QueryRowContext(ctx, query, linkID, boardID).Scan(
		&link.ID, &link.ParentID, &link.ChildID, &link.BoardID, &linkType, &link.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrCardLinkNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select card link: %w", err)
	}

	link.LinkType = domain.CardLinkType(linkType)
	return &link, nil
}

// ListChildren возвращает все дочерние связи карточки (boardID для partition pruning)
func (r *CardLinkRepository) ListChildren(ctx context.Context, parentID, boardID string) ([]*domain.CardLink, error) {
	query := `
		SELECT id, parent_id, child_id, board_id, link_type, created_at
		FROM card_links
		WHERE parent_id = $1 AND board_id = $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, parentID, boardID)
	if err != nil {
		return nil, fmt.Errorf("select card link children: %w", err)
	}
	defer rows.Close()

	var links []*domain.CardLink
	for rows.Next() {
		var l domain.CardLink
		var linkType string
		if err := rows.Scan(&l.ID, &l.ParentID, &l.ChildID, &l.BoardID, &linkType, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan card link: %w", err)
		}
		l.LinkType = domain.CardLinkType(linkType)
		links = append(links, &l)
	}

	return links, rows.Err()
}

// ListParents возвращает все родительские связи карточки (без boardID — child может быть на любой доске)
func (r *CardLinkRepository) ListParents(ctx context.Context, childID string) ([]*domain.CardLink, error) {
	query := `
		SELECT id, parent_id, child_id, board_id, link_type, created_at
		FROM card_links
		WHERE child_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, childID)
	if err != nil {
		return nil, fmt.Errorf("select card link parents: %w", err)
	}
	defer rows.Close()

	var links []*domain.CardLink
	for rows.Next() {
		var l domain.CardLink
		var linkType string
		if err := rows.Scan(&l.ID, &l.ParentID, &l.ChildID, &l.BoardID, &linkType, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan card link: %w", err)
		}
		l.LinkType = domain.CardLinkType(linkType)
		links = append(links, &l)
	}

	return links, rows.Err()
}

// Exists проверяет существование связи между двумя карточками
func (r *CardLinkRepository) Exists(ctx context.Context, parentID, childID, boardID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM card_links WHERE parent_id = $1 AND child_id = $2 AND board_id = $3)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, parentID, childID, boardID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check card link exists: %w", err)
	}

	return exists, nil
}

// isCheckConstraintError проверяет, является ли ошибка check constraint violation
// Работает для PostgreSQL (SQLSTATE 23514)
func isCheckConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "check constraint") || contains(err.Error(), "23514")
}
