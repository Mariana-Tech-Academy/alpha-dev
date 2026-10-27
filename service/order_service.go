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
