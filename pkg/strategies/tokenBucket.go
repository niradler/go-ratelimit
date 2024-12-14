package strategies

import (
	"encoding/json"
	"time"
)

type TokenBucketStrategy struct {
	capacity     int
	refillRate   int
	refillPeriod time.Duration
}

type Bucket struct {
	Tokens    int       `json:"tokens"`
	Timestamp time.Time `json:"timestamp"`
}

func NewTokenBucketStrategy(capacity, refillRate int, refillPeriod time.Duration) *TokenBucketStrategy {
	return &TokenBucketStrategy{
		capacity:     capacity,
		refillRate:   refillRate,
		refillPeriod: refillPeriod,
	}
}

// Use checks if a new request can be accepted and updates the bucket state.
func (s *TokenBucketStrategy) Use(value string) (string, error) {
	var bucket Bucket
	if value != "" {
		err := json.Unmarshal([]byte(value), &bucket)
		if err != nil {
			return "", err
		}
	}

	now := time.Now()
	elapsed := now.Sub(bucket.Timestamp)

	// Refill tokens based on the elapsed time.
	refillTokens := int(elapsed/s.refillPeriod) * s.refillRate
	bucket.Tokens = min(s.capacity, bucket.Tokens+refillTokens)
	bucket.Timestamp = now

	// Check if there are enough tokens for the request.
	if bucket.Tokens <= 0 {
		return "", RateLimitExceededError
	}

	// Decrement the tokens for the current request.
	bucket.Tokens--

	data, err := json.Marshal(bucket)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Reset resets the token bucket state.
func (s *TokenBucketStrategy) Reset() (string, error) {
	bucket := Bucket{
		Tokens:    s.capacity,
		Timestamp: time.Now(),
	}
	data, err := json.Marshal(bucket)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Next returns the time when the next token will be available.
func (s *TokenBucketStrategy) Next(value string) (time.Time, error) {
	var bucket Bucket
	err := json.Unmarshal([]byte(value), &bucket)
	if err != nil {
		return time.Time{}, err
	}

	if bucket.Tokens > 0 {
		return time.Now(), nil
	}

	return bucket.Timestamp.Add(s.refillPeriod), nil
}

// Capacity returns the remaining tokens in the bucket.
func (s *TokenBucketStrategy) Capacity(value string) (int, error) {
	var bucket Bucket
	err := json.Unmarshal([]byte(value), &bucket)
	if err != nil {
		return 0, err
	}
	return bucket.Tokens, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
