package postgres

import "strings"

// IsUniqueViolation проверяет что ошибка PostgreSQL связана с нарушением уникального ограничения.
func IsUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "duplicate key") ||
		strings.Contains(err.Error(), "unique constraint")
}
