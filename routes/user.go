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

func UserRoutes(
	router *mux.Router,
	db *gorm.DB,
	tokens *auth.JWTManager,
) {
	repo := repository.NewUserRepository(db)
	userService := service.NewUserService(repo, tokens)
	userHandler := handlers.NewUserHandler(userService)

	// Public registration: POST /api/auth/register.
	public := router.PathPrefix("/auth").Subrouter()
	public.HandleFunc("/register", userHandler.Register).
		Methods(http.MethodPost)

	// Keep the professor's structure for the other authentication tasks.
	authed := protected(router, tokens)
	_ = authed
}
