package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Atmosfr/user-service/internal/auth"
	"github.com/Atmosfr/user-service/internal/models"
)

type contextKey string

const userKey contextKey = "user"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, `{"error": "missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, `{"error": "empty token"}`, http.StatusUnauthorized)
			return
		}

		user, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, user)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(userKey).(*models.User)
	return user, ok
}
