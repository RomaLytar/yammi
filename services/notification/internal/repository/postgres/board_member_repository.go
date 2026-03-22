package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type BoardMemberRepo struct {
	db *sql.DB
}

func NewBoardMemberRepo(db *sql.DB) *BoardMemberRepo {
	return &BoardMemberRepo{db: db}
}

func (r *BoardMemberRepo) AddMember(ctx context.Context, boardID, userID string) error {
	query := `INSERT INTO board_members (board_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, boardID, userID)
	if err != nil {
		return fmt.Errorf("add board member: %w", err)
	}
	return nil
}

func (r *BoardMemberRepo) RemoveMember(ctx context.Context, boardID, userID string) error {
	query := `DELETE FROM board_members WHERE board_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, boardID, userID)
	if err != nil {
		return fmt.Errorf("remove board member: %w", err)
	}
	return nil
}

func (r *BoardMemberRepo) RemoveAllByBoard(ctx context.Context, boardID string) error {
	query := `DELETE FROM board_members WHERE board_id = $1`
	_, err := r.db.ExecContext(ctx, query, boardID)
	if err != nil {
		return fmt.Errorf("remove all board members: %w", err)
	}
	return nil
}

func (r *BoardMemberRepo) ListMemberIDs(ctx context.Context, boardID string) ([]string, error) {
	query := `SELECT user_id FROM board_members WHERE board_id = $1`
	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("list board members: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan board member: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *BoardMemberRepo) ListBoardIDsByUser(ctx context.Context, userID string) ([]string, error) {
	query := `SELECT board_id FROM board_members WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list boards by user: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan board id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *BoardMemberRepo) TruncateCache(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM board_members")
	if err != nil {
		return fmt.Errorf("truncate board_members: %w", err)
	}
	return nil
}
