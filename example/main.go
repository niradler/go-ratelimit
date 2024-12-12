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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func rateLimitMiddleware(next http.Handler, rateLimiterStrategy rateLimiter.RateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "request", r)
		r = r.WithContext(ctx)
		_, err := rateLimiterStrategy.Use(ctx)
		if err != nil {

			if errors.Is(err, strategies.RateLimitExceededError) {
				http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
				return
			}

			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	strategy := strategies.NewSlidingWindowStrategy(10, time.Minute)
	keyFunc := func(ctx context.Context) (string, error) {
		r, ok := ctx.Value("request").(*http.Request)
		if !ok {
			return "", fmt.Errorf("could not get request from context")
		}

		return r.Header.Get("Authorization"), nil
	}

	slidingWindowRateLimiter, _ := rateLimiter.NewRateLimiter(rateLimiter.RateLimiter{
		Strategy: strategy,
		KeyFunc:  keyFunc,
		DB:       store.NewInMemoryStore(),
	})

	middlewareMux := rateLimitMiddleware(mux, *slidingWindowRateLimiter)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", middlewareMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
