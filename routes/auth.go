// routes/auth.go — registers AUTH-03
package routes

import (
	"net/http"

	"github.com/Mariana-Tech-Academy/alpha-dev/handlers"
	"github.com/Mariana-Tech-Academy/alpha-dev/middleware"
)

func RegisterAuthRoutes(
	mux *http.ServeMux,
	handler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	mux.HandleFunc(
		"GET /api/auth/me",
		authMiddleware.Protect(handler.Me),
	)

	mux.HandleFunc(
		"POST /api/auth/logout",
		authMiddleware.Protect(handler.Logout),
	)
}
