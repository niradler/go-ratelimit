package rateLimiter

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"time"

	"github.com/niradler/go-ratelimit/pkg/store"
	"github.com/niradler/go-ratelimit/pkg/strategies"
)

type KeyGenerator func(ctx context.Context) (string, error)

type RateLimiter struct {
	DB       store.DB
	Strategy strategies.Strategy
	KeyFunc  KeyGenerator
}

type RateLimiterConfig struct {
	DB       store.DB
	Strategy strategies.Strategy
	KeyFunc  KeyGenerator
}

func NewRateLimiter(config RateLimiterConfig) (*RateLimiter, error) {
	if config.Strategy == nil {
		return nil, errors.New("strategy cannot be nil")
	}
	if config.KeyFunc == nil {
		return nil, errors.New("key function cannot be nil")
	}
	if config.DB == nil {
		return nil, errors.New("DB cannot be nil")
	}

	return &RateLimiter{
		Strategy: config.Strategy,
		KeyFunc:  config.KeyFunc,
		DB:       config.DB,
	}, nil
}

func (rl *RateLimiter) Reset(ctx context.Context) error {
	rawKey, err := rl.KeyFunc(ctx)
	if err != nil {
		return err
	}
	hashedKey := hashKey(rawKey)
	newValue, err := rl.Strategy.Reset()
	if err != nil {
		return err
	}

	return rl.DB.Set(hashedKey, newValue)
}

func (rl *RateLimiter) Use(ctx context.Context) (bool, error) {
	rawKey, err := rl.KeyFunc(ctx)
	if err != nil {
		return false, err
	}
	hashedKey := hashKey(rawKey)
	value, err := rl.DB.Get(hashedKey)
	if err != nil {
		if errors.Is(err, store.KeyNotFoundError) {
			value = ""
		} else {
			return false, err
		}
	}

	newValue, err := rl.Strategy.Use(value)
	if err != nil {
		return false, err
	}

	err = rl.DB.Set(hashedKey, newValue)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (rl *RateLimiter) Capacity(ctx context.Context) (int, error) {
	rawKey, err := rl.KeyFunc(ctx)
	if err != nil {
		return 0, err
	}
	hashedKey := hashKey(rawKey)
	value, err := rl.DB.Get(hashedKey)
	if err != nil {
		return 0, err
	}

	return rl.Strategy.Capacity(value)
}

func (rl *RateLimiter) Next(ctx context.Context) (time.Time, error) {
	rawKey, err := rl.KeyFunc(ctx)
	if err != nil {
		return time.Time{}, err
	}
	hashedKey := hashKey(rawKey)
	value, err := rl.DB.Get(hashedKey)
	if err != nil {
		return time.Time{}, err
	}

	return rl.Strategy.Next(value)
}

func hashKey(key string) string {
	hash := md5.Sum([]byte(key))
	return hex.EncodeToString(hash[:])
}
