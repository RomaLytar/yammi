package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type MembershipRepository struct {
	db *sql.DB
}

func NewMembershipRepository(db *sql.DB) *MembershipRepository {
	return &MembershipRepository{db: db}
}

// AddMember добавляет пользователя в доску с указанной ролью и возвращает созданного участника
func (r *MembershipRepository) AddMember(ctx context.Context, boardID, userID string, role domain.Role) (*domain.Member, error) {
	// Проверяем, что роль валидна
	if !role.IsValid() {
		return nil, domain.ErrInvalidRole
	}

	query := `
		INSERT INTO board_members (board_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING joined_at
	`

	var m domain.Member
	m.UserID = userID
	m.Role = role
	err := r.db.QueryRowContext(ctx, query, boardID, userID, role.String()).Scan(&m.JoinedAt)
	if err != nil {
		// Проверяем duplicate key constraint (SQLSTATE 23505)
		if isDuplicateKeyError(err) {
			return nil, domain.ErrMemberExists
		}
		return nil, fmt.Errorf("insert member: %w", err)
	}

	return &m, nil
}

// RemoveMember удаляет пользователя из доски
func (r *MembershipRepository) RemoveMember(ctx context.Context, boardID, userID string) error {
	// Проверяем, что удаляемый пользователь не является owner
	var role string
	err := r.db.QueryRowContext(ctx, "SELECT role FROM board_members WHERE board_id = $1 AND user_id = $2", boardID, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return domain.ErrMemberNotFound
	}
	if err != nil {
		return fmt.Errorf("select member role: %w", err)
	}

	if role == string(domain.RoleOwner) {
		return domain.ErrCannotRemoveOwner
	}

	query := `DELETE FROM board_members WHERE board_id = $1 AND user_id = $2`
	result, err := r.db.ExecContext(ctx, query, boardID, userID)
	if err != nil {
		return fmt.Errorf("delete member: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrMemberNotFound
	}

	return nil
}

// IsMember проверяет, является ли пользователь членом доски и возвращает его роль
func (r *MembershipRepository) IsMember(ctx context.Context, boardID, userID string) (bool, domain.Role, error) {
	query := `SELECT role FROM board_members WHERE board_id = $1 AND user_id = $2`

	var roleStr string
	err := r.db.QueryRowContext(ctx, query, boardID, userID).Scan(&roleStr)

	if err == sql.ErrNoRows {
		return false, "", nil
	}
	if err != nil {
		return false, "", fmt.Errorf("select member: %w", err)
	}

	return true, domain.Role(roleStr), nil
}

// ListMembers возвращает список членов доски с пагинацией (offset-based)
func (r *MembershipRepository) ListMembers(ctx context.Context, boardID string, limit, offset int) ([]*domain.Member, error) {
	query := `
		SELECT user_id, role, joined_at
		FROM board_members
		WHERE board_id = $1
		ORDER BY joined_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, boardID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select members: %w", err)
	}
	defer rows.Close()

	var members []*domain.Member
	for rows.Next() {
		var m domain.Member
		var roleStr string

		if err := rows.Scan(&m.UserID, &roleStr, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}

		m.Role = domain.Role(roleStr)
		members = append(members, &m)
	}

	return members, rows.Err()
}

// isDuplicateKeyError проверяет, является ли ошибка duplicate key constraint
// Работает для PostgreSQL (SQLSTATE 23505)
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	// Упрощенная проверка — в production используйте библиотеку для парсинга PostgreSQL errors
	// Например: github.com/lib/pq или github.com/jackc/pgx/v5/pgconn
	return contains(err.Error(), "duplicate key") || contains(err.Error(), "23505")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
