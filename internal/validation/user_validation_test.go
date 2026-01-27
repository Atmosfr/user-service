package validation

import (
	"errors"
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
			expectErr: ErrPasswordTooShort,
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateRegister(tt.email, tt.password, tt.username)
			if !errors.Is(err, tt.expectErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
