package routes

import (
	"net/http"

	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/repository"
	"alpha-dev/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func UserRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo, tokens)
	userHandler := handlers.NewUserHandler(userService)

	// public routes: POST /auth/register, POST /auth/login
	public := router.PathPrefix("/auth").Subrouter()
	_ = public

	// protected routes
	authed := protected(router, tokens)
	authed.HandleFunc("/auth/me", userHandler.Me).Methods(http.MethodGet)
	authed.HandleFunc("/auth/logout", userHandler.Logout).Methods(http.MethodPost)

	// PUT /users/me and PUT /users/me/password are not implemented yet.
}
