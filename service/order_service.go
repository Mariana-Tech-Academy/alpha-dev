// EPIC 6 - CHK-03: ORDER SERVICE- FILE 2:

// PURPOSE:
// Connects the order handler to the repository layer.

// IMPLEMENTATION:
// 1. Defines the OrderService interface with the PlaceOrder method.
// 2. Stores the order repository as a dependency.
// 3. Uses NewOrderService to initialise the service.
// 4. Passes the customer ID and order information to the repository.
// 5. Returns the created order or any error to the handler.

// HOW IT WORKS:
// The handler calls OrderService.PlaceOrder.
// The service forwards the request to OrderRepository.PlaceOrder.
// The repository performs the database operations and returns the result.

// I implemented the order service to connect the handler and repository.
// It passes the authenticated customer's ID and delivery information
// to the repository, then returns the result to the handler.

package service

import (
    "context"

    "alpha-dev/models"
    "alpha-dev/repository"
)

type OrderService interface {
    PlaceOrder(
        ctx context.Context,
        userID uint,
        order *models.Order,
    ) (*models.Order, error)
}

type orderService struct {
    orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
    return &orderService{orderRepo: repo}
}
func (s *orderService) PlaceOrder(
    ctx context.Context,
    userID uint,
    order *models.Order,
) (*models.Order, error) {

    return s.orderRepo.PlaceOrder(ctx, userID, order)
}
