package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/niradler/go-ratelimit/pkg/rateLimiter"
	"github.com/niradler/go-ratelimit/pkg/store"
	"github.com/niradler/go-ratelimit/pkg/strategies"
)

// example shows how to use the rate limiter package

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func rateLimitMiddleware(next http.Handler, rateLimiters ...rateLimiter.RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		ctx := context.WithValue(r.Context(), "request", r)
		r = r.WithContext(ctx)
		var limiterType = []string{"ddos", "header sliding window", "header token bucket"}
		for i, rl := range rateLimiters {
			capacity, _ := rl.Capacity(ctx)
			nextAvailable, _ := rl.Next(ctx)
			log.Println("Rate limit ", limiterType[i], "capacity:", capacity, ", wait for:", time.Until(nextAvailable))
			_, err := rl.Use(ctx)
			if err != nil {
				if errors.Is(err, strategies.RateLimitExceededError) {
					http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
					return
				}

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func getHeader(key string, prefix string) rateLimiter.KeyGenerator {
	return func(ctx context.Context) (string, error) {
		r, ok := ctx.Value("request").(*http.Request)
		if !ok {
			return "", fmt.Errorf("could not get request from context")
		}

		return prefix + r.Header.Get(key), nil
	}
}

func getRemoteIp(ctx context.Context) (string, error) {
	r, ok := ctx.Value("request").(*http.Request)
	if !ok {
		return "", fmt.Errorf("could not get request from context")
	}

	return r.RemoteAddr, nil
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)

	ddosRateLimiter := &rateLimiter.RateLimiter{
		Strategy: strategies.NewSlidingWindowStrategy(10, time.Second*5),
		KeyFunc:  getRemoteIp,
		DB:       store.NewInMemoryStore(),
	}

	tokenRateLimiter, _ := rateLimiter.NewRateLimiter(rateLimiter.RateLimiterConfig{
		Strategy: strategies.NewSlidingWindowStrategy(30, time.Minute),
		KeyFunc:  getHeader("Authorization", "auth - sliding window:"),
		DB: store.NewRedisStore(store.NewRedisStoreOptions{
			Addr: "localhost:6379",
		}),
	})

	tokenBucketRateLimiter, _ := rateLimiter.NewRateLimiter(rateLimiter.RateLimiterConfig{
		Strategy: strategies.NewTokenBucketStrategy(10, 10, time.Second*5),
		KeyFunc:  getHeader("Authorization", "auth - token bucket:"),
		DB: store.NewRedisStore(store.NewRedisStoreOptions{
			Addr: "localhost:6379",
		}),
	})

	middleware := rateLimitMiddleware(mux, *ddosRateLimiter, *tokenRateLimiter, *tokenBucketRateLimiter)

	middlewareMux := http.NewServeMux()
	middlewareMux.Handle("/", middleware)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", middlewareMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
