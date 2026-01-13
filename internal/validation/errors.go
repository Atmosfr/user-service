package validation

import "errors"

var (
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidUsername  = errors.New("username must be 3-30 characters, alphanumeric")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
