package auth

import (
	"testing"
	"time"

	"github.com/Atmosfr/user-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name:    "generate token for valid user",
			user:    &models.User{ID: 1, Role: "user"},
			wantErr: false,
		},
		{
			name:    "generate token for admin user",
			user:    &models.User{ID: 1, Role: "admin"},
			wantErr: false,
		},
		{
			name:    "empty role is allowed",
			user:    &models.User{ID: 1, Role: ""},
			wantErr: false,
		},
		{
			name:    "nil user is not allowed",
			user:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token, err := GenerateToken(tt.user)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
			}

			parsedToken, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			})
			if err != nil {
				t.Errorf("failed to parse generated token: %v", err)
				return
			}

			claims := parsedToken.Claims.(*Claims)

			if float64(claims.UserID) != float64(tt.user.ID) {
				t.Errorf("expected user_id %v, got %v", tt.user.ID, claims.UserID)
			}

			if claims.Role != tt.user.Role {
				t.Errorf("expected user_role %v, got %v", tt.user.Role, claims.Role)
			}

			if claims.ExpiresAt.Unix() <= time.Now().Unix() {
				t.Errorf("token expired immediately")
			}

			if claims.IssuedAt.Unix() > time.Now().Unix() {
				t.Errorf("token issued in the future")
			}
		})
	}
}
