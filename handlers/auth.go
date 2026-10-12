// handlers/auth.go - endpoint responses

package handlers

import (
	"net/http"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/Mariana-Tech-Academy/alpha-dev/utils"
)

// Profile deliberately excludes passwords and related records.
type Profile struct {
	ID          uint   `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	Status      string `json:"status"`
}

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) Me(
	w http.ResponseWriter,
	r *http.Request,
	user models.User,
) {
	utils.WriteJSON(w, http.StatusOK, Profile{
		ID:          user.ID,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		PhoneNumber: user.PhoneNumber,
		Role:        string(user.Role),
		Status:      string(user.Status),
	})
}

func (h *AuthHandler) Logout(
	w http.ResponseWriter,
	r *http.Request,
	user models.User,
) {
	// The frontend must delete its token.
	// This endpoint does not revoke the JWT on the server.
	w.WriteHeader(http.StatusNoContent)
}
