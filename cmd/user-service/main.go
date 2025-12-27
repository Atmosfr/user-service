package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"

	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/pressly/goose/v3"
)

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	return goose.Up(db, "migrations")
}

func main() {
	ctx := context.Background()

	dsn := "host=localhost user=postgres password=postgres dbname=user_service_db port=5432 sslmode=disable"

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

	select {}
}
