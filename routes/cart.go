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

func CartRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	cartRepo := repository.NewCartRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	cartService := service.NewCartService(cartRepo, menuRepo)
	cartHandler := handlers.NewCartHandler(cartService)

	// Only customers can access their cart.
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	customer.HandleFunc("/cart", cartHandler.GetCart).Methods(http.MethodGet)
	customer.HandleFunc("/cart/items", cartHandler.AddItem).Methods(http.MethodPost)
	customer.HandleFunc("/cart/items/{id}", cartHandler.UpdateQuantity).Methods(http.MethodPut)
}
