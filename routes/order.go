// EPIC 6 - CHK-03: ORDER ROUTES - FILE 4:

// PURPOSE:
// Registers the checkout API endpoint and controls who can access it.

// IMPLEMENTATION:
// 1. Creates the order repository using the database connection.
// 2. Creates the order service and connects it to the repository.
// 3. Creates the order handler and connects it to the service.
// 4. Registers the POST /orders checkout endpoint.
// 5. Uses authentication middleware to protect the endpoint.
// 6. Restricts checkout access to customers.

// HOW IT WORKS:
// An authenticated customer sends a POST request to /orders.
// The middleware checks authentication and the customer's role.
// If authorised, the request is passed to OrderHandler.PlaceOrder.

// I registered the checkout endpoint using Gorilla Mux.
// I connected the repository, service and handler through dependency
// injection. I also protected the endpoint so only authenticated
// customers can place orders.

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
	repo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(repo)
	orderHandler := handlers.NewOrderHandler(orderService)

	// customer routes: POST /orders (checkout)
	customer := protected(router, tokens)
	customer.Use(middleware.RequireRole(models.RoleCustomer))
	customer.HandleFunc("/orders", orderHandler.PlaceOrder).Methods(http.MethodPost)

	// authenticated routes: GET /orders, GET /orders/{id}
	// The service scopes results by role: own orders, own restaurant's orders, or all for admin.
	authed := protected(router, tokens)

	// manager/admin routes: PATCH /orders/{id}/status
	staff := protected(router, tokens)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))

	_, _, _ = customer, authed, staff
}
