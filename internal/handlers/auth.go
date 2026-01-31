package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Atmosfr/user-service/internal/repository"
	"github.com/Atmosfr/user-service/internal/service"
	"github.com/Atmosfr/user-service/internal/validation"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(svc service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Method not allowed",
			})
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Content-Type must be application/json",
			})
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		err := validation.ValidateRegister(req.Email, req.Password, req.Username)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		loginResp, err := svc.Register(r.Context(), req.Email, req.Password, req.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		slog.Info("registration successful", "email", req.Email, "user_id", loginResp.User.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(loginResp)
	}
}

func LoginHandler(svc service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Method not allowed",
			})
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Content-Type must be application/json",
			})
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid request payload",
			})
			return
		}

		err := validation.ValidateLogin(req.Email, req.Password)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			slog.Warn("login validation failed", "email", req.Email, "err", err)
			return
		}

		loginResp, err := svc.Login(r.Context(), req.Email, req.Password)
		if err != nil {
			slog.Warn("login failed", "email", req.Email, "err", err)

			if err == repository.ErrInvalidCredentials ||
				err == repository.ErrInvalidPassword ||
				err == repository.ErrUserNotFound {
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid email or password"})
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": err.Error(),
			})
			return
		}

		slog.Info("login successful", "email", req.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(loginResp)
	}
}
