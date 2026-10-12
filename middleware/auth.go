// middleware/auth.go - protects the endpoints

package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/Mariana-Tech-Academy/alpha-dev/service"
	"github.com/Mariana-Tech-Academy/alpha-dev/utils"
	"github.com/golang-jwt/jwt/v5"
)

type AuthenticatedHandler func(
	http.ResponseWriter,
	*http.Request,
	models.User,
)

type AuthMiddleware struct {
	authService *service.AuthService
	secret      []byte
}

func NewAuthMiddleware(
	authService *service.AuthService,
	secret []byte,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		secret:      secret,
	}
}

func (m *AuthMiddleware) Protect(
	next AuthenticatedHandler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			utils.Unauthorized(w)
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (interface{}, error) {
				return m.secret, nil
			},
			jwt.WithValidMethods([]string{"HS256"}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || token == nil || !token.Valid {
			utils.Unauthorized(w)
			return
		}

		id, err := strconv.ParseUint(claims.Subject, 10, strconv.IntSize)
		if err != nil || id == 0 {
			utils.Unauthorized(w)
			return
		}

		user, err := m.authService.GetActiveUser(r.Context(), uint(id))

		if errors.Is(err, service.ErrUserNotFound) {
			utils.Unauthorized(w)
			return
		}
		if errors.Is(err, service.ErrInactiveUser) {
			utils.WriteJSON(w, http.StatusForbidden, map[string]string{
				"error": "Account is not active",
			})
			return
		}
		if err != nil {
			utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Unable to retrieve user",
			})
			return
		}

		next(w, r, user)
	}
}
