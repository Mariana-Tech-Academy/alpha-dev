package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
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

// RegisterRoutes adds the AUTH-03 endpoints.
// Login must issue HS256 tokens with the user ID as a string in "sub"
// and an expiry timestamp in "exp", using the same JWT_SECRET.
func RegisterRoutes(mux *http.ServeMux, db *gorm.DB, secret []byte) {
	protect := func(next func(http.ResponseWriter, *http.Request, models.User)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")

			parts := strings.Fields(r.Header.Get("Authorization"))
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				unauthorized(w)
				return
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(
				parts[1],
				claims,
				func(token *jwt.Token) (interface{}, error) {
					return secret, nil
				},
				jwt.WithValidMethods([]string{"HS256"}),
				jwt.WithExpirationRequired(),
			)
			if err != nil || !token.Valid {
				unauthorized(w)
				return
			}

			id, err := strconv.ParseUint(claims.Subject, 10, strconv.IntSize)
			if err != nil || id == 0 {
				unauthorized(w)
				return
			}

			var user models.User
			err = db.WithContext(r.Context()).
				Select(
					"id", "first_name", "last_name", "email",
					"phone_number", "role", "status",
				).
				First(&user, uint(id)).Error

			if errors.Is(err, gorm.ErrRecordNotFound) {
				unauthorized(w)
				return
			}
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Unable to retrieve user",
				})
				return
			}

			if string(user.Status) != "active" {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Account is not active",
				})
				return
			}

			next(w, r, user)
		}
	}

	mux.HandleFunc("GET /api/auth/me", protect(
		func(w http.ResponseWriter, r *http.Request, user models.User) {
			writeJSON(w, http.StatusOK, Profile{
				ID:          user.ID,
				FirstName:   user.FirstName,
				LastName:    user.LastName,
				Email:       user.Email,
				PhoneNumber: user.PhoneNumber,
				Role:        string(user.Role),
				Status:      string(user.Status),
			})
		},
	))

	mux.HandleFunc("POST /api/auth/logout", protect(
		func(w http.ResponseWriter, r *http.Request, user models.User) {
			// Client-side logout: the frontend must delete its token.
			// This endpoint does not revoke the JWT on the server.
			w.WriteHeader(http.StatusNoContent)
		},
	))
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	writeJSON(w, http.StatusUnauthorized, map[string]string{
		"error": "Missing, invalid or expired token",
	})
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
