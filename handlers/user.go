package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"alpha-dev/service"
	"alpha-dev/utils"
)

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req service.RegisterRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid registration JSON")
		return
	}

	// Accept one JSON object only.
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid registration JSON")
		return
	}

	user, err := h.UserService.Register(r.Context(), req)
	if err != nil {
		var validationError *service.RegistrationValidationError

		switch {
		case errors.As(err, &validationError):
			utils.ValidationError(w, "check the registration fields", validationError.Fields)
		case errors.Is(err, service.ErrEmailTaken):
			utils.Error(w, http.StatusConflict, "email is already taken")
		default:
			log.Printf("register user: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to create account")
		}
		return
	}

	utils.Success(w, http.StatusCreated, "account created successfully", user.Profile())
}
