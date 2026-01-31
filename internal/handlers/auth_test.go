package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Atmosfr/user-service/internal/models"
	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/Atmosfr/user-service/internal/service"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserService struct {
	mock.Mock
}

func (m *mockUserService) Register(ctx context.Context, email, password, username string) (*service.LoginResponse, error) {
	args := m.Called(ctx, email, password, username)
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (*service.LoginResponse, error) {
	args := m.Called(ctx, email, password)
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name           string
		reqParams      RegisterRequest
		requestBody    string
		setupMock      func(svc *mockUserService)
		method         string
		contentType    string
		expectedStatus int
		wantErr        string
	}{
		{
			name: "valid register request",
			reqParams: RegisterRequest{
				Email:    "LhV4X@example.com",
				Password: "StrongP@ssw0rd!",
				Username: "testuser",
			},
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": "testuser"}`,
			setupMock: func(svc *mockUserService) {
				svc.On("Register", mock.Anything, "LhV4X@example.com", "StrongP@ssw0rd!", "testuser").Return(&service.LoginResponse{
					User: &models.User{
						ID:        1,
						Email:     "LhV4X@example.com",
						Username:  "testuser",
						Role:      "user",
						IsActive:  true,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Token: "mocked-jwt-token",
				}, nil)
			},
			method:         http.MethodPost,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid http method",

			requestBody:    `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": "testuser"}`,
			method:         http.MethodGet,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			wantErr:        "Method not allowed",
		},
		{
			name:           "invalid media type",
			requestBody:    `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": "testuser"}`,
			method:         http.MethodPost,
			contentType:    "text/plain",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			wantErr:        "Content-Type must be application/json",
		},
		{
			name:           "invalid request body",
			requestBody:    `{name invalid json}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid character",
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email", "password": "StrongP@ssw0rd!", "username": "testuser"}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid email format",
		},
		{
			name:        "short password",
			requestBody: `{"email": "LhV4X@example.com", "password": "short", "username": "testuser"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "password must be",
		},
		{
			name:        "invalid username",
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": "testuser!"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "username must be",
		},
		{
			name:        "empty username",
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": ""}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "username must be",
		},
		{
			name:        "empty email",
			requestBody: `{"email": "", "password": "StrongP@ssw0rd!", "username": "adawd"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid email format",
		},
		{
			name:        "empty password",
			requestBody: `{"email": "ex@ex.com", "password": "", "username": "adawd"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
			},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid credentials",
		},
		{
			name:        "service returns error",
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!", "username": "testuser"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
				svc.On("Register", mock.Anything, "LhV4X@example.com", "StrongP@ssw0rd!", "testuser").Return(&service.LoginResponse{}, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			tt.setupMock(svc)

			req, err := http.NewRequest(tt.method, "/register", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := RegisterHandler(svc)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code >= 400 {
				var errResp map[string]string
				err := json.NewDecoder(rr.Body).Decode(&errResp)
				require.NoError(t, err, "failed to decode error response")
				require.Contains(t, errResp["error"], tt.wantErr)
			}

			if rr.Code == http.StatusCreated {
				var resp service.LoginResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tt.reqParams.Email, resp.User.Email)
				require.Equal(t, tt.reqParams.Username, resp.User.Username)
				require.NotEmpty(t, resp.User.Role)
				require.NotEmpty(t, resp.Token)
				require.Empty(t, resp.User.PasswordHash)
				require.True(t, resp.User.IsActive)
				require.NotEmpty(t, resp.User.CreatedAt)
				require.NotEmpty(t, resp.User.UpdatedAt)
			}

			svc.AssertExpectations(t)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name           string
		reqParams      LoginRequest
		requestBody    string
		setupMock      func(svc *mockUserService)
		method         string
		contentType    string
		expectedStatus int
		wantErr        string
	}{
		{
			name: "valid login request",
			reqParams: LoginRequest{
				Email:    "LhV4X@example.com",
				Password: "StrongP@ssw0rd!",
			},
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
				svc.On("Login", mock.Anything, "LhV4X@example.com", "StrongP@ssw0rd!").Return(&service.LoginResponse{
					User: &models.User{
						ID:        1,
						Email:     "LhV4X@example.com",
						Username:  "testuser",
						Role:      "user",
						IsActive:  true,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Token: "mocked-jwt-token",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			wantErr:        "",
		},
		{
			name:           "invalid http method",
			requestBody:    `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!"}`,
			method:         http.MethodGet,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusMethodNotAllowed,
			wantErr:        "Method not allowed",
		},
		{
			name:           "invalid content type",
			requestBody:    `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!"}`,
			method:         http.MethodPost,
			contentType:    "text/plain",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusUnsupportedMediaType,
			wantErr:        "Content-Type must be application/json",
		},
		{
			name:           "invalid login request",
			requestBody:    `{invalid json}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "Invalid request payload",
		},
		{
			name:           "empty email",
			requestBody:    `{"email": "", "password": "StrongP@ssw0rd!"}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid email format",
		},
		{
			name:           "empty password",
			requestBody:    `{"email": "LhV4X@example.com", "password": ""}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid credentials",
		},
		{
			name:           "invalid email format",
			requestBody:    `{"email": "invalid-email", "password": "StrongP@ssw0rd!"}`,
			method:         http.MethodPost,
			contentType:    "application/json",
			setupMock:      func(svc *mockUserService) {},
			expectedStatus: http.StatusBadRequest,
			wantErr:        "invalid email format",
		},
		{
			name:        "invalid password",
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
				svc.On("Login", mock.Anything, "LhV4X@example.com", "StrongP@ssw0rd!").Return((*service.LoginResponse)(nil), repository.ErrInvalidPassword)
			},
			expectedStatus: http.StatusUnauthorized,
			wantErr:        "invalid email or password",
		},
		{
			name:        "service returns error",
			requestBody: `{"email": "LhV4X@example.com", "password": "StrongP@ssw0rd!"}`,
			method:      http.MethodPost,
			contentType: "application/json",
			setupMock: func(svc *mockUserService) {
				svc.On("Login", mock.Anything, "LhV4X@example.com", "StrongP@ssw0rd!").Return((*service.LoginResponse)(nil), errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockUserService{}
			tt.setupMock(svc)

			req, err := http.NewRequest(tt.method, "/login", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", tt.contentType)

			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler := LoginHandler(svc)
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.expectedStatus, rr.Code)

			if rr.Code >= 400 {
				var errResp map[string]string
				err := json.NewDecoder(rr.Body).Decode(&errResp)
				require.NoError(t, err, "failed to decode error response")
				require.Contains(t, errResp["error"], tt.wantErr)
			}

			if rr.Code == http.StatusOK {
				var resp service.LoginResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, tt.reqParams.Email, resp.User.Email)
				require.NotEmpty(t, resp.User.Role)
				require.NotEmpty(t, resp.Token)
				require.Empty(t, resp.User.PasswordHash)
				require.True(t, resp.User.IsActive)
				require.NotEmpty(t, resp.User.CreatedAt)
				require.NotEmpty(t, resp.User.UpdatedAt)
			}
		})
	}
}
