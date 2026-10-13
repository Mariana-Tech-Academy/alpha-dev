package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func AdminRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewAdminRepository(db)
	adminService := service.NewAdminService(repo)
	adminHandler := handlers.NewAdminHandler(adminService)

	// admin routes: GET /admin/users, GET /admin/restaurants, GET /admin/orders, GET /admin/statistics, ...
	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.Authenticate(tokens), middleware.RequireRole(models.RoleAdmin))

	admin.HandleFunc("/test", adminHandler.TestAdmin).Methods(http.MethodGet)
}
