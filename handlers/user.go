package handlers

import (
	"errors"
	"net/http"
	"strings"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/service"
	"alpha-dev/utils"
)

// UserProfile excludes passwords and related records.
type UserProfile struct {
	ID        uint              `json:"id"`
	FirstName string            `json:"first_name"`
	LastName  string            `json:"last_name"`
	Email     string            `json:"email"`
	Phone     string            `json:"phone"`
	Role      models.Role       `json:"role"`
	Status    models.UserStatus `json:"status"`
}

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{UserService: userService}
}

func (h *UserHandler) activeUser(w http.ResponseWriter, r *http.Request) (models.User, bool) {
	claims, err := auth.ClaimsFromContext(r.Context())
	if err != nil || claims == nil || claims.UserID == 0 {
		utils.Error(w, http.StatusUnauthorized, "unauthenticated")
		return models.User{}, false
	}

	user, err := h.UserService.GetActiveUser(r.Context(), claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			utils.Error(w, http.StatusUnauthorized, "user not found")
		case errors.Is(err, service.ErrInactiveUser):
			utils.Error(w, http.StatusForbidden, "account is not active")
		default:
			utils.Error(w, http.StatusInternalServerError, "unable to retrieve user")
		}
		return models.User{}, false
	}

	return user, true
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.activeUser(w, r)
	if !ok {
		return
	}

	utils.Success(w, http.StatusOK, "current user", UserProfile{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Phone:     user.Phone,
		Role:      user.Role,
		Status:    user.Status,
	})
}

func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.activeUser(w, r); !ok {
		return
	}

	tokenString, ok := strings.CutPrefix(
		r.Header.Get("Authorization"),
		"Bearer ",
	)
	if !ok || tokenString == "" {
		utils.Error(
			w,
			http.StatusUnauthorized,
			"missing or malformed authorization header",
		)
		return
	}

	if err := h.UserService.Logout(r.Context(), tokenString); err != nil {
		utils.Error(
			w,
			http.StatusInternalServerError,
			"unable to log out",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
