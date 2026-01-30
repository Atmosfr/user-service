package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Atmosfr/user-service/internal/auth"
	"github.com/Atmosfr/user-service/internal/models"
	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/Atmosfr/user-service/internal/validation"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	err := args.Error(0)
	return err
}

func (m *mockUserRepo) FindByID(ctx context.Context, id int64) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestUserService_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		email      string
		password   string
		username   string
		setupMock  func(repo *mockUserRepo)
		wantErr    error
		wantUserID int64
	}{
		{
			name:     "successful registration",
			email:    "new@example.com",
			password: "StrongPass!12",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				repo.On("FindByEmail", mock.Anything, "new@example.com").Return((*models.User)(nil), repository.ErrUserNotFound)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(nil).Run(func(args mock.Arguments) {
					user := args.Get(1).(*models.User)
					user.ID = 1
				})
			},
			wantErr:    nil,
			wantUserID: 1,
		},
		{
			name:     "email already exists",
			email:    "existing@example.com",
			password: "StrongPass!12",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				repo.On("FindByEmail", mock.Anything, "existing@example.com").Return(&models.User{ID: 999, Email: "existing@example.com"}, nil)
			},
			wantErr:    repository.ErrEmailAlreadyExists,
			wantUserID: 0,
		},
		{
			name:     "invalid email format",
			email:    "invalid-email",
			password: "StrongPass!12",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				// No repository calls expected
			},
			wantErr:    validation.ErrInvalidEmail,
			wantUserID: 0,
		},
		{
			name:     "short password",
			email:    "new@example.com",
			password: "short",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				// No repository calls expected
			},
			wantErr:    validation.ErrPasswordTooShort,
			wantUserID: 0,
		},
		{
			name:     "empty email",
			email:    "",
			password: "StrongPass!12",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				// No repository calls expected
			},
			wantErr:    validation.ErrInvalidEmail,
			wantUserID: 0,
		},
		{
			name:     "invalid username",
			email:    "new@example.com",
			password: "StrongPass!12",
			username: "user!",
			setupMock: func(repo *mockUserRepo) {
				// No repository calls expected
			},
			wantErr:    validation.ErrInvalidUsername,
			wantUserID: 0,
		},
		{
			name:     "empty username",
			email:    "new@example.com",
			password: "StrongPass!12",
			username: "",
			setupMock: func(repo *mockUserRepo) {
				// No repository calls expected
			},
			wantErr:    validation.ErrInvalidUsername,
			wantUserID: 0,
		},
		{
			name:     "database error on create",
			email:    "new@example.com",
			password: "StrongPass!12",
			username: "newuser",
			setupMock: func(repo *mockUserRepo) {
				repo.On("FindByEmail", mock.Anything, "new@example.com").Return((*models.User)(nil), repository.ErrUserNotFound)
				repo.On("Create", mock.Anything, mock.AnythingOfType("*models.User")).Return(errors.New("database error")).Times(1)
			},
			wantErr:    errors.New("database error"),
			wantUserID: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockUserRepo)
			svc := NewUserService(repo)
			auth.JwtSecret = []byte("secret")

			tt.setupMock(repo)

			logResponse, err := svc.Register(ctx, tt.email, tt.password, tt.username)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr.Error())
				return
			}

			require.NoError(t, err)
			require.NotNil(t, logResponse)
			require.NotEmpty(t, logResponse.Token)

			user := logResponse.User
			require.Equal(t, tt.wantUserID, user.ID)
			require.Empty(t, user.PasswordHash)
			require.Equal(t, tt.email, user.Email)
			require.Equal(t, tt.username, user.Username)

			repo.AssertNumberOfCalls(t, "Create", 1)
			repo.AssertExpectations(t)
		})
	}
}
