package middleware

import (
	"net/http"
	"slices"
	"strings"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/utils"
)

// Authenticate rejects requests without a valid bearer token and stores the claims in the request context.
func Authenticate(tokens *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || tokenString == "" {
				utils.Error(w, http.StatusUnauthorized, "missing or malformed authorization header")
				return
			}

			claims, err := tokens.Verify(tokenString)
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}

// RequireRole must be used after Authenticate.
func RequireRole(roles ...models.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := auth.ClaimsFromContext(r.Context())
			if err != nil {
				utils.Error(w, http.StatusUnauthorized, "unauthenticated")
				return
			}
			if !slices.Contains(roles, claims.Role) {
				utils.Error(w, http.StatusForbidden, "you do not have permission to perform this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
