
package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"


)

var ErrOrderForbidden = errors.New("customer is not allowed to access this order")


type OrderService interface {
	GetOrderForCustomer(
		ctx context.Context,
		orderID uint,
		customerID uint,
	) (models.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

func (s *orderService) GetOrderForCustomer(
	ctx context.Context,
	orderID uint,
	customerID uint,
) (models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return models.Order{}, err
	}

	if order.CustomerID != customerID {
		return models.Order{}, ErrOrderForbidden
	}

	return order, nil
}

