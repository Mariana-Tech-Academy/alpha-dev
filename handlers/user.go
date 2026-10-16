package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"alpha-dev/auth"
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req service.LoginRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid login JSON")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid login JSON")
		return
	}

	result, err := h.UserService.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			utils.Error(w, http.StatusUnauthorized, "invalid email or password")
		case errors.Is(err, service.ErrInactiveUser):
			utils.Error(w, http.StatusForbidden, "account is not active")
		default:
			log.Printf("login: %v", err)
			utils.Error(w, http.StatusInternalServerError, "unable to log in")
		}
		return
	}

	utils.Success(w, http.StatusOK, "login successful", result)
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	user, err := h.UserService.GetActiveUser(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			utils.Error(w, http.StatusUnauthorized, "user not found")
		case errors.Is(err, service.ErrInactiveUser):
			utils.Error(w, http.StatusForbidden, "account is not active")
		default:
			log.Printf("get active user %d: %v", claims.UserID, err)
			utils.Error(w, http.StatusInternalServerError, "unable to retrieve user")
		}
		return
	}

	utils.Success(w, http.StatusOK, "current user", user.Profile())
}

// Logout revokes the current token. Suspended users can still log out.
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.UserService.Logout(r.Context(), claims); err != nil {
		log.Printf("logout user %d: %v", claims.UserID, err)
		utils.Error(w, http.StatusInternalServerError, "unable to log out")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 16*1024)

	var req service.ChangePasswordRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "invalid password change JSON")
		return
	}

	if err := decoder.Decode(new(any)); err != io.EOF {
		utils.Error(w, http.StatusBadRequest, "invalid password change JSON")
		return
	}

	if err := h.UserService.ChangePassword(r.Context(), claims, req); err != nil {
		switch {
		case errors.Is(err, service.ErrIncorrectPassword):
			utils.Error(w, http.StatusUnauthorized, "incorrect current password")
		case errors.Is(err, service.ErrPasswordMismatch),
			errors.Is(err, service.ErrWeakPassword),
			errors.Is(err, service.ErrPasswordTooLong),
			errors.Is(err, service.ErrPasswordRequired),
			errors.Is(err, service.ErrConfirmRequired),
			errors.Is(err, service.ErrSamePassword):
			utils.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			utils.Error(w, http.StatusUnauthorized, "user not found")
		default:
			log.Printf("change password for user %d: %v", claims.UserID, err)
			utils.Error(w, http.StatusInternalServerError, "unable to change password")
		}
		return
	}

	utils.Success(w, http.StatusOK, "password updated successfully", nil)
}
