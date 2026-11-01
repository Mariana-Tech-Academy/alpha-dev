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

	registerDemoOrderRoute(router, tokens, orderHandler)

	staff := protected(router, tokens)
	staff.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))
	staff.HandleFunc("/orders/{id}/status", orderHandler.UpdateOrderStatus).Methods(http.MethodPatch)
	staff.HandleFunc("/manager/orders", orderHandler.GetManagerOrders).Methods(http.MethodGet)

}

func registerDemoOrderRoute(router *mux.Router, tokens *auth.JWTManager, handler *handlers.OrderHandler) {
	// Gate before JWT authentication; customer tokens are never sufficient.
	demo := router.NewRoute().Subrouter()
	demo.Use(middleware.LocalDemoOnly)
	demo.Use(middleware.Authenticate(tokens))
	demo.Use(middleware.RequireRole(models.RoleManager, models.RoleAdmin))
	demo.HandleFunc("/demo/orders/{id}/advance", handler.AdvanceDemoOrder).Methods(http.MethodPost)
}
