package store

import (
	"testing"
)

func TestReset(t *testing.T) {
	db := NewInMemoryStore()
	db.Set("my-key", "my-value")
	value, _ := db.Get("my-key")
	if value != "my-value" {
		t.Fatalf("expected value to be 'my-value', got %v", value)
	}
	db.Delete("my-key")
	value, _ = db.Get("my-key")
	if value != "" {
		t.Fatalf("expected value to be '', got %v", value)
	}
}
