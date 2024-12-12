package store

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisStore(addr string, password string, db int) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisStore{
		client: rdb,
		ctx:    context.Background(),
	}
}

// Init initializes the Redis store.
func (s *RedisStore) Init() error {
	return s.client.Ping(s.ctx).Err()
}

// Get retrieves a value from the Redis store.
func (s *RedisStore) Get(key string) (string, error) {
	val, err := s.client.Get(s.ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("key not found")
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Set sets a value in the Redis store.
func (s *RedisStore) Set(key string, value string) error {
	return s.client.Set(s.ctx, key, value, 0).Err()
}

// Delete deletes a value from the Redis store.
func (s *RedisStore) Delete(key string) error {
	return s.client.Del(s.ctx, key).Err()
}
