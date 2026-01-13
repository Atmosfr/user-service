package validation

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
	Username string `validate:"required,alphanum,min=3,max=30"`
}

type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

func ValidateRegister(email, password, username string) error {
	req := &RegisterRequest{
		Email:    email,
		Password: password,
		Username: username,
	}
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		for _, e := range errs {
			switch e.Tag() {
			case "required":
				switch e.Field() {
				case "Password":
					return ErrPasswordTooShort
				case "Email":
					return ErrInvalidEmail
				case "Username":
					return ErrInvalidUsername

				default:
					return ErrInvalidCredentials
				}
			case "email":
				return ErrInvalidEmail
			case "min":
				if e.Field() == "Password" {
					return ErrPasswordTooShort
				}
				if e.Field() == "Username" {
					return ErrInvalidUsername
				}
			case "max":
				return ErrInvalidUsername
			case "alphanum":
				return ErrInvalidUsername
			}
		}
	}

	return err
}

func ValidateLogin(email, password string) error {
	req := &LoginRequest{
		Email:    email,
		Password: password,
	}
	err := validate.Struct(req)
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors

	if errors.As(err, &errs) {
		for _, e := range errs {
			switch e.Tag() {
			case "required":
				switch e.Field() {
				case "Password":
					return ErrPasswordTooShort
				case "Email":
					return ErrInvalidEmail
				default:
					return ErrInvalidCredentials
				}
			case "email":
				return ErrInvalidEmail
			}
		}
	}

	return err
}
