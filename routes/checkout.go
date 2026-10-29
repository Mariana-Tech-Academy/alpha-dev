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

func CheckoutRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	cartRepo := repository.NewCartRepository(db)
	checkoutService := service.NewCheckoutService(cartRepo, service.DefaultDeliveryFee)
	checkoutHandler := handlers.NewCheckoutHandler(checkoutService)

	// customer routes: GET /checkout/summary
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	customer.HandleFunc("/checkout/summary", checkoutHandler.Summary).Methods(http.MethodGet)
}