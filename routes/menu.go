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

func MenuRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(repo)
	menuHandler := handlers.NewMenuHandler(menuService)

	// public routes: GET /restaurants/{id}/menu, GET /menu/{id}
	router.HandleFunc("/menu/{id}", menuHandler.GetMenuItemDetails).Methods(http.MethodGet)

	// manager routes: POST /menu, PUT /menu/{id}, DELETE /menu/{id}
	// The service must also check the item belongs to the calling manager's restaurant.
	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager))

	manager.HandleFunc("/menu", menuHandler.CreateMenuItem).Methods(http.MethodPost)

	manager.HandleFunc("/menu/{id}/availability", menuHandler.UpdateMenuItemAvailability).Methods(http.MethodPatch)
}
