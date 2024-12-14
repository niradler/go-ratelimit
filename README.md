# Rate Limit Module

This is an open-source solution to manage rate limiting in Go applications. It provides a flexible and easy-to-use interface to implement rate limiting strategies.

## Features

- Rate limit strategies like sliding window and token bucket.
- In-Memory/Redis Store for Rate Limits.

## Installation

To install the package, use the following command:

```sh
go get github.com/niradler/go-ratelimit
```

## Usage

Here is an example of how to use the rate limiter package:

```go
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
```

## Extending the Package

### Adding a New Strategy

To add a new strategy, implement the `Strategy` interface defined in `pkg/strategies/strategies.go`:

```go
type Strategy interface {
    Use(value string) (string, error)
    Reset() (string, error)
    Next(value string) (time.Time, error)
    Capacity(value string) (int, error)
}
```

### Adding a New Store

To add a new store, implement the `DB` interface defined in `pkg/store/store.go`:

```go
type DB interface {
    Init() error
    Get(key string) (string, error)
    Set(key string, value string) error
    Delete(key string) error
}
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

We welcome contributions! Please follow these steps to contribute:

1. Fork the repository.
2. Create a new branch (`git checkout -b feature-branch`).
3. Make your changes.
4. Commit your changes (`git commit -am 'Add new feature'`).
5. Push to the branch (`git push origin feature-branch`).
6. Create a new Pull Request.

For major changes, please open an issue first to discuss what you would like to change.

Thank you for your contributions!