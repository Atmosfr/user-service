package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Atmosfr/user-service/internal/handlers"
	"github.com/Atmosfr/user-service/internal/middleware"
	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/Atmosfr/user-service/internal/service"
	"github.com/pressly/goose/v3"
)

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	return goose.Up(db, "migrations")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "user not found in context", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"email":      user.Email,
		"username":   user.Username,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"is_active":  user.IsActive,
	})
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dsn := "host=db user=postgres password=postgres dbname=user_service_db port=5432 sslmode=disable"

	db, err := repository.NewDB(ctx, dsn)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := runMigrations(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	slog.Info("database migrations applied successfully")

	// router
	mux := http.NewServeMux()
	repo := repository.NewUserRepository(db)
	authMiddleware := middleware.NewAuthMiddleware(repo)
	svc := service.NewUserService(repo)

	mux.Handle("POST /register", handlers.RegisterHandler(svc))
	mux.Handle("POST /login", handlers.LoginHandler(svc))
	mux.Handle("GET /health", http.HandlerFunc(healthHandler))
	mux.Handle("GET /me", authMiddleware(http.HandlerFunc(meHandler)))

	// server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("shutting down server...")
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown failed", "error", err)
		} else {
			slog.Info("server shutdown completed")
		}
	}()

	slog.Info("starting server on :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
