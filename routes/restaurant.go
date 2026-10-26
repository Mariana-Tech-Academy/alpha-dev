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

func RestaurantRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewRestaurantRepository(db)
	restaurantService := service.NewRestaurantService(repo)
	restaurantHandler := handlers.NewRestaurantHandler(restaurantService)

	// public routes: GET /restaurants, GET /restaurants/{id}
	router.HandleFunc("/restaurants/{id}", restaurantHandler.GetRestaurantDetails).Methods(http.MethodGet)
	// manager routes: POST /restaurants, PUT /restaurants/{id}, PATCH /restaurants/{id}/status
	// The service must also check the restaurant belongs to the calling manager.
	manager := protected(router, tokens)
	manager.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	_, _ = restaurantHandler, manager
}
