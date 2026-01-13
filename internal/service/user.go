package service

import (
	"context"
	"log/slog"

	"github.com/Atmosfr/user-service/internal/models"
	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/Atmosfr/user-service/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, email, password, username string) (*models.User, error)
	Login(ctx context.Context, email, password string) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func (u *userService) Register(ctx context.Context, email, password, username string) (*models.User, error) {
	if err:= validation.ValidateRegister(email, password, username); err != nil {
		slog.Warn("registration validation failed", "email", email, "err", err)
		return nil, err
	}

	if _, err := u.repo.FindByEmail(ctx, email); err == nil {
		return nil, repository.ErrEmailAlreadyExists
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	passwordHash := string(bytes)

	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		Username:     username,
	}

	err = u.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""

	slog.Info("user registered", "email", email)
	return user, nil
}

func (u *userService) Login(ctx context.Context, email, password string) (*models.User, error) {
	if err := validation.ValidateLogin(email, password);  err != nil {
		slog.Warn("login validation failed", "email", email, "err", err)
		return nil, err
	}

	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		slog.Warn("user not found", "email", email)
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		slog.Warn("wrong password for user", "email", email)
		return nil, repository.ErrInvalidPassword
	}

	user.PasswordHash = ""
	slog.Info("login successful", "email", email)
	return user, nil
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}
