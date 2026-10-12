// middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Mariana-Tech-Academy/alpha-dev/auth"
)

// contextKey is a private type so our keys can't clash with other packages.
//UserID and role are in auth packages so this prevents mix ups
type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

// Authenticate requires a valid "Authorization: Bearer <token>" header.
// On success it stores the user's ID and role in the request context.
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		// Header must look like: Bearer <token>
		tokenString, found := strings.CutPrefix(header, "Bearer ")
		if !found || tokenString == "" {
			http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Use the auth package to verify the token.
		userID, role, err := auth.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Save the verified values for the handlers that run next.
		//lets you attach info to a request so code further down line can read it
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		ctx = context.WithValue(ctx, roleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole allows only users whose role matches one of the given roles.
// Use it after Authenticate.
//can allow several roles too
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := GetRole(r)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}

// GetUserID returns the authenticated user's ID from the request.
func GetUserID(r *http.Request) (uint, bool) {
	id, ok := r.Context().Value(userIDKey).(uint)
	return id, ok
}

// GetRole returns the authenticated user's role from the request.
func GetRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(roleKey).(string)
	return role, ok
}