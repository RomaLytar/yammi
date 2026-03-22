package postgres

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

const (
	maxRetries = 3
	retryDelay = 50 * time.Millisecond
)

// isRetryableError проверяет, является ли ошибка временной (PgBouncer transient).
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "server closed") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "no more connections") ||
		strings.Contains(msg, "driver: bad connection") ||
		err == sql.ErrConnDone
}

// retryExec выполняет ExecContext с retry при transient ошибках.
func retryExec(ctx context.Context, db *sql.DB, query string, args ...interface{}) (sql.Result, error) {
	var result sql.Result
	var err error

	for i := 0; i < maxRetries; i++ {
		result, err = db.ExecContext(ctx, query, args...)
		if err == nil || !isRetryableError(err) {
			return result, err
		}
		log.Printf("retryable DB error (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay * time.Duration(i+1))
	}

	return result, err
}

// retryQueryRow выполняет QueryRowContext с retry при transient ошибках.
func retryQueryRow(ctx context.Context, db *sql.DB, query string, args ...interface{}) *retriableRow {
	return &retriableRow{ctx: ctx, db: db, query: query, args: args}
}

type retriableRow struct {
	ctx   context.Context
	db    *sql.DB
	query string
	args  []interface{}
}

func (r *retriableRow) Scan(dest ...interface{}) error {
	var err error

	for i := 0; i < maxRetries; i++ {
		err = r.db.QueryRowContext(r.ctx, r.query, r.args...).Scan(dest...)
		if err == nil || !isRetryableError(err) {
			return err
		}
		log.Printf("retryable DB error (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(retryDelay * time.Duration(i+1))
	}

	return err
}
