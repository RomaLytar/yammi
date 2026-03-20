package domain

import (
	"testing"
)

func TestLexorankBetween(t *testing.T) {
	tests := []struct {
		name    string
		prev    string
		next    string
		want    string
		wantErr bool
	}{
		{
			name: "both empty - returns first position",
			prev: "",
			next: "",
			want: "n",
		},
		{
			name: "only prev - append after",
			prev: "a",
			next: "",
			want: "an",
		},
		{
			name: "only next - insert before",
			prev: "",
			next: "z",
			want: "0",
		},
		{
			name: "between two positions",
			prev: "aaaa",
			next: "ab",
			want: "aaam",
		},
		{
			name: "between adjacent characters",
			prev: "a",
			next: "b",
			want: "am",
		},
		{
			name: "between same length strings",
			prev: "abc",
			next: "abd",
			want: "abcm",
		},
		{
			name: "between different length strings",
			prev: "a",
			next: "abc",
			want: "aam",
		},
		{
			name:    "prev >= next - error",
			prev:    "z",
			next:    "a",
			wantErr: true,
		},
		{
			name:    "prev == next - error",
			prev:    "a",
			next:    "a",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LexorankBetween(tt.prev, tt.next)
			if (err != nil) != tt.wantErr {
				t.Errorf("LexorankBetween() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("LexorankBetween() = %v, want %v", got, tt.want)
			}

			// Дополнительная проверка: результат должен быть между prev и next
			if !tt.wantErr && tt.prev != "" && tt.next != "" {
				if got <= tt.prev || got >= tt.next {
					t.Errorf("LexorankBetween() result %v not between %v and %v", got, tt.prev, tt.next)
				}
			}
		})
	}
}

func TestLexorankOrdering(t *testing.T) {
	// Тест проверяет, что позиции сортируются лексикографически
	positions := []string{"a", "am", "b", "c"}

	// Генерируем новую позицию между "a" и "am"
	between, err := LexorankBetween("a", "am")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверяем, что она меньше "am" и больше "a"
	if between <= "a" || between >= "am" {
		t.Errorf("generated position %v not between 'a' and 'am'", between)
	}

	// Вставляем в середину массива
	result := []string{"a", between, "am", "b", "c"}

	// Проверяем сортировку
	for i := 0; i < len(result)-1; i++ {
		if result[i] >= result[i+1] {
			t.Errorf("positions not ordered: %v >= %v", result[i], result[i+1])
		}
	}
}

func TestLexorankSequence(t *testing.T) {
	// Симуляция последовательных вставок (как в реальном использовании)
	var positions []string

	// Вставляем первую карточку
	first, _ := LexorankBetween("", "")
	positions = append(positions, first)

	// Вставляем вторую карточку в конец
	second, _ := LexorankBetween(first, "")
	positions = append(positions, second)

	// Вставляем третью карточку в конец
	third, _ := LexorankBetween(second, "")
	positions = append(positions, third)

	// Вставляем карточку между первой и второй
	between, _ := LexorankBetween(first, second)
	positions = []string{first, between, second, third}

	// Проверяем сортировку
	for i := 0; i < len(positions)-1; i++ {
		if positions[i] >= positions[i+1] {
			t.Errorf("positions not ordered after insertions: %v >= %v", positions[i], positions[i+1])
		}
	}

	t.Logf("Generated positions: %v", positions)
}

func TestValidateLexorank(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid lowercase", "abc", false},
		{"valid digits", "123", false},
		{"valid mixed", "a1b2c3", false},
		{"empty string", "", true},
		{"invalid uppercase", "ABC", true},
		{"invalid special char", "a-b", true},
		{"invalid space", "a b", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLexorank(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateLexorank() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Benchmark для проверки производительности
func BenchmarkLexorankBetween(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = LexorankBetween("aaaa", "ab")
	}
}

func TestLexorankEdgeCases(t *testing.T) {
	t.Run("very long positions", func(t *testing.T) {
		prev := "aaaaaaaaaaaaaaaa"
		next := "aaaaaaaaaaaaaaab"
		result, err := LexorankBetween(prev, next)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result <= prev || result >= next {
			t.Errorf("result %v not between %v and %v", result, prev, next)
		}
	})

	t.Run("positions with trailing zeros", func(t *testing.T) {
		prev := "a000"
		next := "a100"
		result, err := LexorankBetween(prev, next)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result <= prev || result >= next {
			t.Errorf("result %v not between %v and %v", result, prev, next)
		}
	})

	t.Run("append many times", func(t *testing.T) {
		// Проверяем, что можно много раз добавлять в конец
		pos := "a"
		for i := 0; i < 10; i++ {
			newPos, err := LexorankBetween(pos, "")
			if err != nil {
				t.Fatalf("iteration %d: unexpected error: %v", i, err)
			}
			if newPos <= pos {
				t.Fatalf("iteration %d: new position %v not greater than %v", i, newPos, pos)
			}
			pos = newPos
		}
		t.Logf("Final position after 10 appends: %s", pos)
	})
}
