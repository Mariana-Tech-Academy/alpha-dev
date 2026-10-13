package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"alpha-dev/repository"
	"alpha-dev/service"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func registrationJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		registrationJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"success": false,
			"message": "Use POST to register",
		})
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)
	defer r.Body.Close()

	var req service.RegisterRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		registrationJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid registration JSON",
		})
		return
	}

	// Accept one JSON object only.
	if err := decoder.Decode(new(any)); err != io.EOF {
		registrationJSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid registration JSON",
		})
		return
	}

	user, err := h.UserService.Register(r.Context(), req)
	if err != nil {
		var validationError *service.RegistrationValidationError

		switch {
		case errors.As(err, &validationError):
			registrationJSON(w, http.StatusUnprocessableEntity, map[string]any{
				"success": false,
				"message": "Check the registration fields",
				"errors":  validationError.Fields,
			})

		case errors.Is(err, repository.ErrEmailTaken):
			registrationJSON(w, http.StatusConflict, map[string]any{
				"success": false,
				"message": "Email is already taken",
			})

		default:
			registrationJSON(w, http.StatusInternalServerError, map[string]any{
				"success": false,
				"message": "Unable to create account",
			})
		}
		return
	}

	registrationJSON(w, http.StatusCreated, map[string]any{
		"success": true,
		"message": "Account created successfully",
		"data": map[string]any{
			"first_name":   user.FirstName,
			"last_name":    user.LastName,
			"email":        user.Email,
			"phone_number": user.Phone,
			"role":         user.Role,
			"status":       user.Status,
		},
	})
}
