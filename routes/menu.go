package routes

import (
	"alpha-dev/auth"
	"alpha-dev/handlers"
	"alpha-dev/middleware"
	"alpha-dev/models"
	"alpha-dev/repository"
	"alpha-dev/service"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func MenuRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(repo)
	menuHandler := handlers.NewMenuHandler(menuService)

	// public routes: GET /restaurants/{id}/menu, GET /menu/{id}

	// manager routes: POST /menu, PUT /menu/{id}, DELETE /menu/{id}
	// The service must also check the item belongs to the calling manager's restaurant.
	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	_, _ = menuHandler, manager
}
