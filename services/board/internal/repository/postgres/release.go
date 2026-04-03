package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/RomaLytar/yammi/services/board/internal/domain"
)

type ReleaseRepository struct {
	db *sql.DB
}

func NewReleaseRepository(db *sql.DB) *ReleaseRepository {
	return &ReleaseRepository{db: db}
}

// scanRelease сканирует релиз из строки результата (DRY helper)
func scanRelease(scanner interface{ Scan(dest ...interface{}) error }) (*domain.Release, error) {
	var r domain.Release
	var startDate, endDate, startedAt, completedAt sql.NullTime
	var status string
	err := scanner.Scan(&r.ID, &r.BoardID, &r.Name, &r.Description, &status, &startDate, &endDate, &startedAt, &completedAt, &r.CreatedBy, &r.Version, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.Status = domain.ReleaseStatus(status)
	if startDate.Valid {
		r.StartDate = &startDate.Time
	}
	if endDate.Valid {
		r.EndDate = &endDate.Time
	}
	if startedAt.Valid {
		r.StartedAt = &startedAt.Time
	}
	if completedAt.Valid {
		r.CompletedAt = &completedAt.Time
	}
	return &r, nil
}

// Create создает новый релиз
func (r *ReleaseRepository) Create(ctx context.Context, release *domain.Release) error {
	query := `
		INSERT INTO releases (id, board_id, name, description, status, start_date, end_date, started_at, completed_at, created_by, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(ctx, query,
		release.ID, release.BoardID, release.Name, release.Description,
		string(release.Status), release.StartDate, release.EndDate,
		release.StartedAt, release.CompletedAt,
		release.CreatedBy, release.Version, release.CreatedAt, release.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert release: %w", err)
	}

	return nil
}

// GetByID возвращает релиз по ID (фильтруется по boardID для защиты от IDOR)
func (r *ReleaseRepository) GetByID(ctx context.Context, releaseID, boardID string) (*domain.Release, error) {
	query := `
		SELECT id, board_id, name, description, status, start_date, end_date, started_at, completed_at, created_by, version, created_at, updated_at
		FROM releases
		WHERE id = $1 AND board_id = $2
	`

	release, err := scanRelease(r.db.QueryRowContext(ctx, query, releaseID, boardID))
	if err == sql.ErrNoRows {
		return nil, domain.ErrReleaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select release: %w", err)
	}

	return release, nil
}

// ListByBoardID возвращает все релизы доски
func (r *ReleaseRepository) ListByBoardID(ctx context.Context, boardID string) ([]*domain.Release, error) {
	query := `
		SELECT id, board_id, name, description, status, start_date, end_date, started_at, completed_at, created_by, version, created_at, updated_at
		FROM releases
		WHERE board_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, boardID)
	if err != nil {
		return nil, fmt.Errorf("select releases: %w", err)
	}
	defer rows.Close()

	var releases []*domain.Release
	for rows.Next() {
		release, err := scanRelease(rows)
		if err != nil {
			return nil, fmt.Errorf("scan release: %w", err)
		}
		releases = append(releases, release)
	}

	return releases, rows.Err()
}

// GetActiveByBoardID возвращает активный релиз доски (или ErrReleaseNotFound)
func (r *ReleaseRepository) GetActiveByBoardID(ctx context.Context, boardID string) (*domain.Release, error) {
	query := `
		SELECT id, board_id, name, description, status, start_date, end_date, started_at, completed_at, created_by, version, created_at, updated_at
		FROM releases
		WHERE board_id = $1 AND status = 'active'
	`

	release, err := scanRelease(r.db.QueryRowContext(ctx, query, boardID))
	if err == sql.ErrNoRows {
		return nil, domain.ErrReleaseNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select active release: %w", err)
	}

	return release, nil
}

// Update обновляет релиз
func (r *ReleaseRepository) Update(ctx context.Context, release *domain.Release) error {
	query := `
		UPDATE releases
		SET name = $1, description = $2, status = $3, start_date = $4, end_date = $5, started_at = $6, completed_at = $7, version = $8, updated_at = $9
		WHERE id = $10 AND board_id = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		release.Name, release.Description, string(release.Status),
		release.StartDate, release.EndDate,
		release.StartedAt, release.CompletedAt, release.Version, release.UpdatedAt,
		release.ID, release.BoardID,
	)
	if err != nil {
		return fmt.Errorf("update release: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrReleaseNotFound
	}

	return nil
}

// Delete удаляет релиз по ID (фильтруется по boardID для защиты от IDOR)
func (r *ReleaseRepository) Delete(ctx context.Context, releaseID, boardID string) error {
	query := `DELETE FROM releases WHERE id = $1 AND board_id = $2`
	result, err := r.db.ExecContext(ctx, query, releaseID, boardID)
	if err != nil {
		return fmt.Errorf("delete release: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrReleaseNotFound
	}

	return nil
}

// CountByBoardID возвращает количество релизов доски
func (r *ReleaseRepository) CountByBoardID(ctx context.Context, boardID string) (int, error) {
	query := `SELECT COUNT(*) FROM releases WHERE board_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, boardID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count releases: %w", err)
	}

	return count, nil
}
