package integration

import "testing"

// --- Assertion helpers ---

func requireStatus(t *testing.T, operation string, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: expected HTTP %d, got %d", operation, want, got)
	}
}

func requireNotEmpty(t *testing.T, field, value string) {
	t.Helper()
	if value == "" {
		t.Fatalf("%s must not be empty", field)
	}
}

func requireEqual(t *testing.T, field, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: expected %q, got %q", field, want, got)
	}
}

func requireNotEqual(t *testing.T, field, a, b string) {
	t.Helper()
	if a == b {
		t.Fatalf("%s: expected different values, both are %q", field, a)
	}
}
