package strategies

import (
	"errors"
	"time"
)

var RateLimitExceededError = errors.New("rate limit exceeded")

type Strategy interface {
	Use(value string) (string, error)
	Reset() (string, error)
	Next(value string) (time.Time, error)
	Capacity(value string) (int, error)
}
