package validation

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateRegister(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		username  string
		expectErr error
	}{
		{
			name:      "valid input",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "validuser",
			expectErr: nil,
		},
		{
			name:      "invalid email",
			email:     "invalid-email",
			password:  "StrongP@ssw0rd!",
			username:  "validuser",
			expectErr: ErrInvalidEmail,
		},
		{
			name:      "empty email",
			email:     "",
			password:  "StrongP@ssw0rd!",
			username:  "validuser",
			expectErr: ErrInvalidEmail,
		},
		{
			name:      "empty password",
			email:     "LhV4X@example.com",
			password:  "",
			username:  "validuser",
			expectErr: ErrInvalidCredentials,
		},
		{
			name:      "short password",
			email:     "LhV4X@example.com",
			password:  "short",
			username:  "validuser",
			expectErr: ErrPasswordTooShort,
		},
		{
			name:      "empty username",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "",
			expectErr: ErrInvalidUsername,
		},
		{
			name:      "too long username",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "verylongusernamewithmorethanthirtycharacters",
			expectErr: ErrInvalidUsername,
		},
		{
			name:      "too short username",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "us",
			expectErr: ErrInvalidUsername,
		},
		{
			name:      "username with special characters",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "user!name",
			expectErr: ErrInvalidUsername,
		},
		{
			name:      "valid input with tag",
			email:     "test+tag@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "valid_user_123",
			expectErr: nil,
		},
		{
			name:      "username with underscore and digits",
			email:     "test@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "user_123",
			expectErr: nil,
		},
		{
			name:      "username with space",
			email:     "test@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "user name",
			expectErr: ErrInvalidUsername,
		},
		{
			name:      "username with hyphen",
			email:     "test@example.com",
			password:  "StrongP@ssw0rd!",
			username:  "user-name",
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRegister(tt.email, tt.password, tt.username)
			if tt.expectErr != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tt.expectErr)
				} else if !errors.Is(err, tt.expectErr) {
					t.Errorf("expected error %v, got %v", tt.expectErr, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateLogin(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		password  string
		expectErr error
	}{
		{
			name:      "valid input",
			email:     "LhV4X@example.com",
			password:  "StrongP@ssw0rd!",
			expectErr: nil,
		},
		{
			name:      "invalid email",
			email:     "invalid-email",
			password:  "StrongP@ssw0rd!",
			expectErr: ErrInvalidEmail,
		},
		{
			name:      "empty email",
			email:     "",
			password:  "StrongP@ssw0rd!",
			expectErr: ErrInvalidEmail,
		},
		{
			name:      "empty password",
			email:     "LhV4X@example.com",
			password:  "",
			expectErr: ErrInvalidCredentials,
		},
		{
			name:      "short password",
			email:     "LhV4X@example.com",
			password:  "short",
			expectErr: ErrPasswordTooShort,
		},
		{
			name:      "very long password",
			email:     "test@example.com",
			password:  strings.Repeat("a", 100),
			expectErr: nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateLogin(tt.email, tt.password)
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
