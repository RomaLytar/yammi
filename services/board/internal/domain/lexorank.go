package domain

import (
	"fmt"
	"strings"
)

// Lexorank — алгоритм для позиционирования элементов без массовых UPDATE.
// Используется для card.Position (string вместо INT).
//
// Принцип: каждая карточка имеет лексикографическую позицию (строку).
// При перемещении карточки между двумя другими — генерируем позицию между ними.
//
// Пример:
//   Позиции: "aaaa", "ab", "b", "c"
//   Переместить между "aaaa" и "ab" → Between("aaaa", "ab") = "aaam"
//   Результат: "aaaa", "aaam", "ab", "b", "c"
//
// Плюсы:
//   - UPDATE только одной карточки (не всех)
//   - Нет race conditions при concurrent reorder
//   - O(1) для INSERT/MOVE операций

const (
	// Alphabet для lexorank (36 символов: a-z, 0-9)
	lexorankAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	lexorankBase     = len(lexorankAlphabet)

	// Начальные позиции для новых элементов
	LexorankFirst  = "n"       // Первая позиция (середина алфавита)
	LexorankMiddle = "n"
)

// LexorankBetween генерирует позицию между prev и next.
// Если prev пуст — возвращает позицию перед next.
// Если next пуст — возвращает позицию после prev.
func LexorankBetween(prev, next string) (string, error) {
	// Оба пусты — возвращаем начальную позицию
	if prev == "" && next == "" {
		return LexorankFirst, nil
	}

	// Только prev — добавляем в конец
	if next == "" {
		return incrementLexorank(prev), nil
	}

	// Только next — вставляем перед ним
	if prev == "" {
		return decrementLexorank(next), nil
	}

	// Оба заполнены — ищем середину
	if prev >= next {
		return "", fmt.Errorf("%w: prev (%s) must be < next (%s)", ErrInvalidLexorank, prev, next)
	}

	return midpoint(prev, next), nil
}

// midpoint вычисляет лексикографическую середину между a и b
func midpoint(a, b string) string {
	// Выравниваем длины строк (добавляем '0' в конец короткой)
	maxLen := max(len(a), len(b))
	a = padRight(a, maxLen)
	b = padRight(b, maxLen)

	var result strings.Builder
	carry := 0

	for i := 0; i < maxLen; i++ {
		aVal := charToValue(a[i])
		bVal := charToValue(b[i])

		// Среднее значение + carry из предыдущего разряда
		mid := (aVal + bVal + carry) / 2

		if (aVal+bVal+carry)%2 == 1 {
			carry = lexorankBase // перенос в следующий разряд
		} else {
			carry = 0
		}

		result.WriteByte(valueToChar(mid))
	}

	// Если carry остался — добавляем дополнительный символ
	if carry > 0 {
		result.WriteByte(valueToChar(carry / 2))
	}

	return strings.TrimRight(result.String(), "0")
}

// incrementLexorank увеличивает позицию (для добавления в конец)
func incrementLexorank(s string) string {
	if s == "" {
		return LexorankFirst
	}
	return s + "n" // добавляем середину алфавита
}

// decrementLexorank уменьшает позицию (для добавления в начало)
func decrementLexorank(s string) string {
	if s == "" {
		return "0"
	}

	// Берем на один символ меньше, чем s
	if len(s) == 1 {
		val := charToValue(s[0])
		if val == 0 {
			return "00n" // если '0', идем глубже
		}
		return string(valueToChar(val - 1))
	}

	return s[:len(s)-1] + "0" // уменьшаем последний символ
}

// charToValue конвертирует символ в числовое значение (0-35)
func charToValue(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'z' {
		return int(c-'a') + 10
	}
	return 0
}

// valueToChar конвертирует число (0-35) в символ
func valueToChar(v int) byte {
	if v < 10 {
		return byte('0' + v)
	}
	return byte('a' + (v - 10))
}

// padRight дополняет строку нулями справа до заданной длины
func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat("0", length-len(s))
}

// Validate проверяет валидность lexorank строки
func ValidateLexorank(s string) error {
	if s == "" {
		return fmt.Errorf("%w: empty position", ErrInvalidLexorank)
	}

	for _, c := range s {
		if !isValidLexorankChar(byte(c)) {
			return fmt.Errorf("%w: invalid character '%c'", ErrInvalidLexorank, c)
		}
	}

	return nil
}

// isValidLexorankChar проверяет, допустим ли символ в lexorank
func isValidLexorankChar(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z')
}
