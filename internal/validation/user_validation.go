package validation

import (
	"errors"
	"regexp"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
	trans    ut.Translator
)

func init() {
	validate = validator.New()
	validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		return regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(s)
	})

	enTrans := en.New()
	uni := ut.New(enTrans, enTrans)
	trans, _ = uni.GetTranslator("en")

	validate.RegisterTranslation("username", trans, func(ut ut.Translator) error {
		return ut.Add("username", "{0} can only contain letters, numbers, underscores or hyphens", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Field())
		return t + " can only contain letters, numbers, underscores or hyphens"
	})

	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is required", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Field())
		return t + " is required"
	})

	validate.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must be at least {param}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Field())
		return t + " must be at least " + fe.Param() + " characters"
	})

	validate.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "{0} must be at most {param}", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Field())
		return t + " must be at most " + fe.Param() + " characters"
	})

	validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} must be a valid email address", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Field())
		return t + " must be a valid email address"
	})
}

type RegisterRequest struct {
	Email    string `validate:"required,email" json:"email"`
	Password string `validate:"required,min=8" json:"password"`
	Username string `validate:"required,username,min=3,max=30" json:"username"`
}

type LoginRequest struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8"`
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
					return ErrInvalidCredentials
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
			case "username":
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
					return ErrInvalidCredentials
				case "Email":
					return ErrInvalidEmail
				default:
					return ErrInvalidCredentials
				}
			case "email":
				return ErrInvalidEmail
			case "min":
				if e.Field() == "Password" {
					return ErrPasswordTooShort
				}
			}
		}
	}

	return err
}
