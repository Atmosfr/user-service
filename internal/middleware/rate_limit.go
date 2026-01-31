package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	limiter "github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	redisstore "github.com/ulule/limiter/v3/drivers/store/redis"
)

func RateLimitMiddleware(rate string) func(next http.Handler) http.Handler {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		_, err := fmt.Sscanf(redisDBStr, "%d", &redisDB)
		if err != nil {
			slog.Warn("invalid REDIS_DB, using default 0", "err", err)
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to Redis, falling back to in-memory store", "err", err)
		store := memory.NewStore()
		return createLimiterMiddleware(store, rate)
	}

	store, err := redisstore.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "ratelimit:",
		MaxRetry: 3,
	})
	if err != nil {
		slog.Error("failed to create Redis store", "err", err)
		store = memory.NewStore()
	}

	return createLimiterMiddleware(store, rate)
}

func createLimiterMiddleware(store limiter.Store, rate string) func(next http.Handler) http.Handler {
	rateLimit, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		slog.Error("invalid rate format, using default", "rate", rate, "err", err)
		rateLimit = limiter.Rate{
			Period: time.Second,
			Limit:  10,
		}
	}

	lim := limiter.New(store, rateLimit)

	middleware := stdlib.NewMiddleware(lim, stdlib.WithKeyGetter(func(r *http.Request) string {
		ip := r.RemoteAddr
		if colon := strings.LastIndex(ip, ":"); colon != -1 {
			ip = ip[:colon]
		}
		slog.Info("Rate limit key", "original", r.RemoteAddr, "clean_ip", ip)
		return ip
	}))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware.Handler(next).ServeHTTP(w, r)

			if w.Header().Get("X-RateLimit-Remaining") == "0" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error": "Too many requests. Try again later."}`))
				return
			}
		})
	}
}
