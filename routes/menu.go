package routes

import (
	"net/http"

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
	// Connect the repository, service and handler.
	repo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(repo)
	menuHandler := handlers.NewMenuHandler(menuService)

	// View a restaurant's menu.
	router.HandleFunc(
		"/restaurants/{id}/menu",
		menuHandler.GetRestaurantMenu,
	).Methods(http.MethodGet)

	// Allow managers to create menu items.
	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager))
	manager.HandleFunc(
		"/menu",
		menuHandler.CreateMenuItem,
	).Methods(http.MethodPost)
}
