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

func CartRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	repo := repository.NewCartRepository(db)
	cartService := service.NewCartService(repo)
	cartHandler := handlers.NewCartHandler(cartService)

	// customer routes: GET /cart, POST /cart/items, PUT /cart/items/{id}, DELETE /cart/items/{id}, DELETE /cart
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	_, _ = cartHandler, customer
}
