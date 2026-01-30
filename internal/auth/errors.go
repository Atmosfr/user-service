package auth

import (
	"errors"
)

var (
	ErrEmptyJwtSecret = errors.New("jwt secret is empty")
	ErrInvalidUserId  = errors.New("invalid user ID")
	ErrUserIsNil      = errors.New("user is nil")
	ErrInvalidToken   = errors.New("invalid token")
)
