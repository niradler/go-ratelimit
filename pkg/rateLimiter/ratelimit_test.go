package rateLimiter

import (
	"context"
	"testing"
	"time"

	"github.com/niradler/go-ratelimit/pkg/store"
)

// MockStrategy is a mock implementation of the strategies.Strategy interface.
type MockStrategy struct {
	CapacityFunc func(value string) (int, error)
	NextFunc     func(value string) (time.Time, error)
	ResetFunc    func() (string, error)
	UseFunc      func(value string) (string, error)
}

func (m *MockStrategy) Capacity(value string) (int, error) {
	if m.CapacityFunc != nil {
		return m.CapacityFunc(value)
	}
	return 0, nil
}

func (m *MockStrategy) Next(value string) (time.Time, error) {
	if m.NextFunc != nil {
		return m.NextFunc(value)
	}
	return time.Time{}, nil
}

func (m *MockStrategy) Reset() (string, error) {
	if m.ResetFunc != nil {
		return m.ResetFunc()
	}
	return "", nil
}

func (m *MockStrategy) Use(value string) (string, error) {
	if m.UseFunc != nil {
		return m.UseFunc(value)
	}
	return "", nil
}

// NewMockDB is a mock implementation of a database.
func NewMockDB() *store.InMemoryStore {
	return store.NewInMemoryStore()
}

func TestNewRateLimiter(t *testing.T) {
	tests := []struct {
		name    string
		config  RateLimiterConfig
		wantErr bool
	}{
		{
			name: "valid configuration",
			config: RateLimiterConfig{
				DB:       NewMockDB(),
				Strategy: &MockStrategy{},
				KeyFunc:  func(ctx context.Context) (string, error) { return "test_key", nil },
			},
			wantErr: false,
		},
		{
			name: "nil strategy",
			config: RateLimiterConfig{
				DB:       NewMockDB(),
				Strategy: nil,
				KeyFunc:  func(ctx context.Context) (string, error) { return "test_key", nil },
			},
			wantErr: true,
		},
		{
			name: "nil database",
			config: RateLimiterConfig{
				DB:       nil,
				Strategy: &MockStrategy{},
				KeyFunc:  func(ctx context.Context) (string, error) { return "test_key", nil },
			},
			wantErr: true,
		},
		{
			name: "nil key function",
			config: RateLimiterConfig{
				DB:       NewMockDB(),
				Strategy: &MockStrategy{},
				KeyFunc:  nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRateLimiter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRateLimiter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReset(t *testing.T) {
	db := NewMockDB()
	strategy := &MockStrategy{
		ResetFunc: func() (string, error) { return "reset_value", nil },
	}
	keyFunc := func(ctx context.Context) (string, error) { return "test_key", nil }

	rl, err := NewRateLimiter(RateLimiterConfig{DB: db, Strategy: strategy, KeyFunc: keyFunc})
	if err != nil {
		t.Fatalf("Failed to create RateLimiter: %v", err)
	}

	ctx := context.Background()
	err = rl.Reset(ctx)
	if err != nil {
		t.Fatalf("Reset() error = %v", err)
	}

	value, err := db.Get(hashKey("test_key"))
	if err != nil {
		t.Fatalf("Failed to get value from DB: %v", err)
	}
	if value != "reset_value" {
		t.Errorf("expected reset_value, got %v", value)
	}
}

func TestUse(t *testing.T) {
	db := NewMockDB()
	strategy := &MockStrategy{
		UseFunc: func(value string) (string, error) { return "used", nil },
	}
	keyFunc := func(ctx context.Context) (string, error) { return "test_key", nil }

	rl, err := NewRateLimiter(RateLimiterConfig{DB: db, Strategy: strategy, KeyFunc: keyFunc})
	if err != nil {
		t.Fatalf("Failed to create RateLimiter: %v", err)
	}

	ctx := context.Background()
	success, err := rl.Use(ctx)
	if err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	if !success {
		t.Errorf("expected success, got %v", success)
	}

	value, err := db.Get(hashKey("test_key"))
	if err != nil {
		t.Fatalf("Failed to get value from DB: %v", err)
	}
	if value != "used" {
		t.Errorf("expected used, got %v", value)
	}
}
