package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func RateLimitMiddleware(rate string) func(next http.Handler) http.Handler {
	store := memory.NewStore()
	rateLimit, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		panic(err)
	}

	limiter := limiter.New(store, rateLimit)

	middleware := stdlib.NewMiddleware(limiter, stdlib.WithKeyGetter(func(r *http.Request) string {
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
