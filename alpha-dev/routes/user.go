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
	public.HandleFunc("/register", userHandler.Register).Methods(http.MethodPost)
	public.HandleFunc("/login", userHandler.Login).Methods(http.MethodPost)

	// protected routes: POST /auth/logout, GET /auth/me, PUT /users/me, PUT /users/me/password
	authed := protected(router, tokens)
	authed.HandleFunc("/auth/me", userHandler.Me).Methods(http.MethodGet)
	authed.HandleFunc("/auth/logout", userHandler.Logout).Methods(http.MethodPost)
	authed.HandleFunc("/users/me", userHandler.UpdateUserProfile).Methods(http.MethodPut)
	authed.HandleFunc("/users/me/password", userHandler.ChangePassword).Methods(http.MethodPut)
}
