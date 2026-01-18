package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Atmosfr/user-service/internal/service"
	"github.com/Atmosfr/user-service/internal/validation"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func RegisterHandler(svc service.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := validation.ValidateRegister(req.Email, req.Password, req.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		loginResp, err := svc.Register(r.Context(), req.Email, req.Password, req.Username)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(loginResp)
	}
}
