package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/RomaLytar/yammi/services/board/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoardRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	board := &domain.Board{
		ID:          "board-123",
		Title:       "Test Board",
		Description: "Description",
		OwnerID:     "user-123",
		Version:     1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Ожидаем начало транзакции
	mock.ExpectBegin()

	// Ожидаем INSERT в boards
	mock.ExpectExec("INSERT INTO boards").
		WithArgs(board.ID, board.Title, board.Description, board.OwnerID, board.Version, board.CreatedAt, board.UpdatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем INSERT в board_members
	mock.ExpectExec("INSERT INTO board_members").
		WithArgs(board.ID, board.OwnerID, board.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Ожидаем commit
	mock.ExpectCommit()

	err = repo.Create(context.Background(), board)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_GetByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	boardID := "board-123"
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "title", "description", "owner_id", "version", "created_at", "updated_at"}).
		AddRow(boardID, "Test Board", "Description", "user-123", 1, now, now)

	mock.ExpectQuery("SELECT (.+) FROM boards WHERE id").
		WithArgs(boardID).
		WillReturnRows(rows)

	board, err := repo.GetByID(context.Background(), boardID)
	assert.NoError(t, err)
	assert.NotNil(t, board)
	assert.Equal(t, boardID, board.ID)
	assert.Equal(t, "Test Board", board.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_GetByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	mock.ExpectQuery("SELECT (.+) FROM boards WHERE id").
		WithArgs("non-existent").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "description", "owner_id", "version", "created_at", "updated_at"}))

	board, err := repo.GetByID(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBoardNotFound)
	assert.Nil(t, board)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_Update_OptimisticLocking(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	board := &domain.Board{
		ID:          "board-123",
		Title:       "Updated Title",
		Description: "Updated Description",
		Version:     2, // Новая версия
		UpdatedAt:   time.Now(),
	}

	// Ожидаем UPDATE с проверкой старой версии (version = 1)
	mock.ExpectExec("UPDATE boards").
		WithArgs(board.Title, board.Description, board.Version, board.UpdatedAt, board.ID, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Update(context.Background(), board)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_Update_ConflictVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	board := &domain.Board{
		ID:          "board-123",
		Title:       "Updated Title",
		Description: "Updated Description",
		Version:     2,
		UpdatedAt:   time.Now(),
	}

	// Возвращаем 0 затронутых строк (version conflict)
	mock.ExpectExec("UPDATE boards").
		WithArgs(board.Title, board.Description, board.Version, board.UpdatedAt, board.ID, 1).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Update(context.Background(), board)
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidVersion)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_Delete(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	boardID := "board-123"

	mock.ExpectExec("DELETE FROM boards WHERE id").
		WithArgs(boardID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Delete(context.Background(), boardID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBoardRepository_Delete_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewBoardRepository(db)

	mock.ExpectExec("DELETE FROM boards WHERE id").
		WithArgs("non-existent").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Delete(context.Background(), "non-existent")
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrBoardNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}
