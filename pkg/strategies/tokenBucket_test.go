package strategies

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTokenBucketStrategy(t *testing.T) {
	capacity := 10
	refillRate := 1
	refillPeriod := time.Second

	strategy := NewTokenBucketStrategy(capacity, refillRate, refillPeriod)

	if strategy.capacity != capacity {
		t.Errorf("expected capacity %d, got %d", capacity, strategy.capacity)
	}
	if strategy.refillRate != refillRate {
		t.Errorf("expected refillRate %d, got %d", refillRate, strategy.refillRate)
	}
	if strategy.refillPeriod != refillPeriod {
		t.Errorf("expected refillPeriod %v, got %v", refillPeriod, strategy.refillPeriod)
	}
}

func TestTokenBucketStrategy_Use(t *testing.T) {
	strategy := NewTokenBucketStrategy(10, 1, time.Second)
	initialState, err := strategy.Reset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use a token
	newState, err := strategy.Use(initialState)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bucket Bucket
	err = json.Unmarshal([]byte(newState), &bucket)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bucket.Tokens != 9 {
		t.Errorf("expected 9 tokens, got %d", bucket.Tokens)
	}
}

func TestTokenBucketStrategy_Reset(t *testing.T) {
	strategy := NewTokenBucketStrategy(10, 1, time.Second)
	state, err := strategy.Reset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bucket Bucket
	err = json.Unmarshal([]byte(state), &bucket)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bucket.Tokens != 10 {
		t.Errorf("expected 10 tokens, got %d", bucket.Tokens)
	}
}

func TestTokenBucketStrategy_Next(t *testing.T) {
	strategy := NewTokenBucketStrategy(10, 1, time.Second)
	initialState, err := strategy.Reset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Use all tokens
	state := initialState
	for i := 0; i < 10; i++ {
		state, err = strategy.Use(state)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	nextTime, err := strategy.Next(state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if nextTime.Before(time.Now()) {
		t.Errorf("expected next time to be in the future, got %v", nextTime)
	}
}

func TestTokenBucketStrategy_Capacity(t *testing.T) {
	strategy := NewTokenBucketStrategy(10, 1, time.Second)
	initialState, err := strategy.Reset()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	capacity, err := strategy.Capacity(initialState)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capacity != 10 {
		t.Errorf("expected capacity 10, got %d", capacity)
	}
}
