package postgres

import (
	"context"
	"database/sql"
)

type NameCacheRepo struct {
	db *sql.DB
}

func NewNameCacheRepo(db *sql.DB) *NameCacheRepo {
	return &NameCacheRepo{db: db}
}

// --- Board names ---

func (r *NameCacheRepo) SetBoardName(ctx context.Context, boardID, title string) error {
	query := `INSERT INTO board_names (board_id, title) VALUES ($1, $2)
		ON CONFLICT (board_id) DO UPDATE SET title = $2`
	_, err := r.db.ExecContext(ctx, query, boardID, title)
	return err
}

func (r *NameCacheRepo) GetBoardName(ctx context.Context, boardID string) string {
	var title string
	r.db.QueryRowContext(ctx, `SELECT title FROM board_names WHERE board_id = $1`, boardID).Scan(&title)
	return title
}

func (r *NameCacheRepo) DeleteBoardName(ctx context.Context, boardID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM board_names WHERE board_id = $1`, boardID)
	return err
}

// --- User names ---

func (r *NameCacheRepo) SetUserName(ctx context.Context, userID, name string) error {
	query := `INSERT INTO user_names (user_id, name) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET name = $2`
	_, err := r.db.ExecContext(ctx, query, userID, name)
	return err
}

func (r *NameCacheRepo) GetUserName(ctx context.Context, userID string) string {
	var name string
	r.db.QueryRowContext(ctx, `SELECT name FROM user_names WHERE user_id = $1`, userID).Scan(&name)
	return name
}

// --- Card names ---

func (r *NameCacheRepo) SetCardName(ctx context.Context, cardID, title string) error {
	query := `INSERT INTO card_names (card_id, title) VALUES ($1, $2)
		ON CONFLICT (card_id) DO UPDATE SET title = $2`
	_, err := r.db.ExecContext(ctx, query, cardID, title)
	return err
}

func (r *NameCacheRepo) GetCardName(ctx context.Context, cardID string) string {
	var title string
	r.db.QueryRowContext(ctx, `SELECT title FROM card_names WHERE card_id = $1`, cardID).Scan(&title)
	return title
}

func (r *NameCacheRepo) DeleteCardName(ctx context.Context, cardID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM card_names WHERE card_id = $1`, cardID)
	return err
}

// --- Column names ---

func (r *NameCacheRepo) SetColumnName(ctx context.Context, columnID, title string) error {
	query := `INSERT INTO column_names (column_id, title) VALUES ($1, $2)
		ON CONFLICT (column_id) DO UPDATE SET title = $2`
	_, err := r.db.ExecContext(ctx, query, columnID, title)
	return err
}

func (r *NameCacheRepo) GetColumnName(ctx context.Context, columnID string) string {
	var title string
	r.db.QueryRowContext(ctx, `SELECT title FROM column_names WHERE column_id = $1`, columnID).Scan(&title)
	return title
}

func (r *NameCacheRepo) DeleteColumnName(ctx context.Context, columnID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM column_names WHERE column_id = $1`, columnID)
	return err
}
