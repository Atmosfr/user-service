package auth

import (
	"testing"
	"time"

	"github.com/Atmosfr/user-service/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateToken(t *testing.T) {
	JwtSecret = []byte("secret")

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
		{
			name:    "invalid user ID",
			user:    &models.User{ID: 0, Role: "user"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token, err := GenerateToken(tt.user, time.Hour*24, JwtSecret)

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
				return JwtSecret, nil
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

func TestValidateToken(t *testing.T) {
	JwtSecret = []byte("secret")
	validUser := &models.User{ID: 1, Role: "user"}
	validToken, _ := GenerateToken(validUser, time.Hour*24, JwtSecret)
	expiredToken, _ := GenerateToken(validUser, -time.Hour, JwtSecret)

	tests := []struct {
		name     string
		token    string
		wantUser *models.User
		wantErr  bool
	}{
		{
			name:     "valid token",
			token:    validToken,
			wantUser: validUser,
			wantErr:  false,
		},
		{
			name:     "invalid signature",
			token:    validToken + "tampered",
			wantUser: nil,
			wantErr:  true,
		},
		{
			name:     "expired token",
			token:    expiredToken,
			wantUser: nil,
			wantErr:  true,
		},
		{
			name: "zero duration token",
			token: generateTamperedToken(t, validUser, func(claims *Claims) {
				claims.ExpiresAt = jwt.NewNumericDate(time.Now())
			}, jwt.SigningMethodHS256),
			wantUser: nil,
			wantErr:  true,
		},
		{
			name:     "token without user_id",
			token:    generateTokenWithoutUserID(t, validUser),
			wantUser: nil,
			wantErr:  true,
		},
		{
			name:     "invalid token",
			token:    validToken + "invalid",
			wantUser: nil,
			wantErr:  true,
		},
		{
			name: "token with future issued_at",
			token: generateTamperedToken(t, validUser, func(c *Claims) {
				c.IssuedAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
			}, jwt.SigningMethodHS256),
			wantUser: nil,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			user, err := ValidateToken(tt.token, JwtSecret)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user.ID != tt.wantUser.ID {
				t.Errorf("expected user ID %v, got %v", tt.wantUser.ID, user.ID)
			}

			if user.Role != tt.wantUser.Role {
				t.Errorf("expected user role %v, got %v", tt.wantUser.Role, user.Role)
			}
		})
	}
}

func generateTamperedToken(t *testing.T, user *models.User, modify func(*Claims), signingMethod jwt.SigningMethod) string {
	t.Helper()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	modify(&claims)
	token := jwt.NewWithClaims(signingMethod, claims)
	tokenString, err := token.SignedString(JwtSecret)
	if err != nil {
		t.Fatalf("failed to sign tampered token: %v", err)
	}
	return tokenString
}

func generateTokenWithoutUserID(t *testing.T, user *models.User) string {
	t.Helper()
	claims := Claims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(JwtSecret)
	if err != nil {
		t.Fatalf("failed to sign token without user_id: %v", err)
	}
	return s
}
