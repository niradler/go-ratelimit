package strategies

import (
	"testing"
	"time"
)

func TestSlidingWindow(t *testing.T) {
	t.Run("test case 1", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(10, time.Second)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}
	})

	t.Run("test case 2", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(1, time.Second)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}

		use, err = strategy.Use(use)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("test case 3", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(2, time.Second)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}

		use, err = strategy.Use(use)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}

		use, err = strategy.Use(use)
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("test case 4", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(1, time.Millisecond)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}

		time.Sleep(2 * time.Millisecond)

		use, err = strategy.Use(use)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if use == "" {
			t.Fatalf("expected non-empty value, got %v", use)
		}
	})

	t.Run("test case 5", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(5, time.Second)
		reset, err := strategy.Reset()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if reset == "" {
			t.Fatalf("expected non-empty value, got %v", reset)
		}
	})

	t.Run("test case 6", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(5, time.Second)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		next, err := strategy.Next(use)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if next.Before(time.Now()) {
			t.Fatalf("expected next time to be in the future, got %v", next)
		}
	})

	t.Run("test case 7", func(t *testing.T) {
		strategy := NewSlidingWindowStrategy(5, time.Second)
		use, err := strategy.Use("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		capacity, err := strategy.Capacity(use)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capacity != 4 {
			t.Fatalf("expected capacity to be 4, got %v", capacity)
		}
	})
}
