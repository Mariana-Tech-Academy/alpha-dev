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

func OrderRoutes(router *mux.Router, db *gorm.DB, tokens *auth.JWTManager) {
	// Connect the order repository, service and handler.
	repo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(repo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// Customers can place orders and view their own history.
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))

	customer.HandleFunc("/orders", orderHandler.PlaceOrder).Methods(http.MethodPost)
	customer.HandleFunc("/orders", orderHandler.GetOrderHistory).Methods(http.MethodGet)

	// Keep the existing route groups for future team features.
	authed := protected(router, tokens)

	staff := protected(router, tokens)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))
	staff.HandleFunc("/orders/{id}/status", orderHandler.UpdateOrderStatus).Methods(http.MethodPatch)
	staff.HandleFunc("/manager/orders", orderHandler.GetManagerOrders).Methods(http.MethodGet)

	_ = authed
}
