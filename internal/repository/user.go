package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/Atmosfr/user-service/internal/models"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id int64) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	sql := `INSERT INTO users (email, password_hash, username) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRowContext(ctx, sql, user.Email, user.PasswordHash, user.Username).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrEmailAlreadyExists
		}
		return err
	}

	return nil
}
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {}
func (r *userRepository) FindByID(ctx context.Context, id int64) (*models.User, error)        {}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}
