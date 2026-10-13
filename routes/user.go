package routes

import (
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

	// protected routes: POST /auth/logout, GET /auth/me, PUT /users/me, PUT /users/me/password
	authed := protected(router, tokens)

	_, _, _ = public, authed, userHandler
}
