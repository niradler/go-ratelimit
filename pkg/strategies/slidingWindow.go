package strategies

import (
	"encoding/json"
	"time"
)

type SlidingWindowStrategy struct {
	limit    int
	interval time.Duration
}

type Window struct {
	Timestamp time.Time `json:"timestamp"`
	Count     int       `json:"count"`
}

func NewSlidingWindowStrategy(limit int, interval time.Duration) *SlidingWindowStrategy {
	return &SlidingWindowStrategy{
		limit:    limit,
		interval: interval,
	}
}

// Use checks if a new request can be accepted and updates the window state.
func (s *SlidingWindowStrategy) Use(value string) (string, error) {
	var window Window
	if value != "" {
		err := json.Unmarshal([]byte(value), &window)
		if err != nil {
			return "", err
		}
	}

	now := time.Now()
	windowAge := now.Sub(window.Timestamp)

	// If the current interval has passed, reset the count.
	if windowAge > s.interval {
		window.Timestamp = now
		window.Count = 0
	}

	// Check if adding a new request would exceed the limit.
	if window.Count >= s.limit {
		return "", RateLimitExceededError
	}

	// Increment the count for the current request.
	window.Count++

	data, err := json.Marshal(window)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Reset resets the sliding window state.
func (s *SlidingWindowStrategy) Reset() (string, error) {
	window := Window{
		Timestamp: time.Now(),
		Count:     0,
	}
	data, err := json.Marshal(window)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *SlidingWindowStrategy) Next(value string) (time.Time, error) {
	var window Window
	err := json.Unmarshal([]byte(value), &window)
	if err != nil {
		return time.Time{}, err
	}

	if window.Count < s.limit {
		return time.Now(), nil
	}

	return window.Timestamp.Add(s.interval), nil
}

// Capacity returns the remaining capacity for the current interval.
func (s *SlidingWindowStrategy) Capacity(value string) (int, error) {
	var window Window
	err := json.Unmarshal([]byte(value), &window)
	if err != nil {
		return 0, err
	}
	return s.limit - window.Count, nil
}
