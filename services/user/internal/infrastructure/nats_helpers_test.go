package infrastructure

import (
	"testing"
	"time"
)

func TestBackoffDelay(t *testing.T) {
	t.Run("attempt 0 returns approximately 1s", func(t *testing.T) {
		delay := backoffDelay(0)
		base := 1 * time.Second
		lower := time.Duration(float64(base) * (1 - jitterFactor))
		upper := time.Duration(float64(base) * (1 + jitterFactor))
		if delay < lower || delay > upper {
			t.Fatalf("expected delay in [%v, %v], got %v", lower, upper, delay)
		}
	})

	t.Run("attempt 3 returns approximately 8s", func(t *testing.T) {
		delay := backoffDelay(3)
		base := 8 * time.Second // 2^3 = 8
		lower := time.Duration(float64(base) * (1 - jitterFactor))
		upper := time.Duration(float64(base) * (1 + jitterFactor))
		if delay < lower || delay > upper {
			t.Fatalf("expected delay in [%v, %v], got %v", lower, upper, delay)
		}
	})

	t.Run("attempt exceeding maxBackoff is capped at 30s", func(t *testing.T) {
		// 2^5 = 32s which exceeds maxBackoff (30s), so delay should be capped
		delay := backoffDelay(5)
		lower := time.Duration(float64(maxBackoff) * (1 - jitterFactor))
		upper := time.Duration(float64(maxBackoff) * (1 + jitterFactor))
		if delay < lower || delay > upper {
			t.Fatalf("expected delay in [%v, %v] (capped at maxBackoff), got %v", lower, upper, delay)
		}
	})

	t.Run("jitter is within plus or minus 20 percent", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			delay := backoffDelay(2)
			base := 4 * time.Second // 2^2 = 4
			lower := time.Duration(float64(base) * (1 - jitterFactor))
			upper := time.Duration(float64(base) * (1 + jitterFactor))
			if delay < lower || delay > upper {
				t.Fatalf("iteration %d: expected delay in [%v, %v], got %v", i, lower, upper, delay)
			}
		}
	})

	t.Run("jitter produces variation", func(t *testing.T) {
		seen := make(map[time.Duration]bool)
		for i := 0; i < 100; i++ {
			delay := backoffDelay(2)
			seen[delay] = true
		}
		if len(seen) < 2 {
			t.Fatal("expected jitter to produce varying delays, but all were identical")
		}
	})
}
