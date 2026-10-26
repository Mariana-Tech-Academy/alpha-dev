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

func CartRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	cartRepo := repository.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo, db)
	cartHandler := handlers.NewCartHandler(cartService)

	// customer routes: GET /cart, POST /cart/items, PUT /cart/items/{id}, DELETE /cart/items/{id}, DELETE /cart
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	customer.HandleFunc(
		"/cart/items",
		cartHandler.AddItem,
	).Methods(http.MethodPost)
}
