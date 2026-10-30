package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrOrderNotFound         = errors.New("order not found")
	ErrNotYourOrder          = errors.New("you cannot cancel this order")
	ErrOrderAlreadyCancelled = errors.New("order is already cancelled")
	ErrOrderNotCancellable   = errors.New("order can no longer be cancelled")
)

type CancelledOrder struct {
	ID          uint               `json:"id"`
	OrderNumber string             `json:"order_number"`
	Status      models.OrderStatus `json:"status"`
}

// CancelOrder cancels a pending or confirmed order.
func (s *orderService) CancelOrder(ctx context.Context, userID uint, role models.Role, orderID uint) (CancelledOrder, error) {
	if userID == 0 {
		return CancelledOrder{}, ErrInvalidUser
	}

	// 1. Find the order.
	order, err := s.orderRepo.FindOrderByID(ctx, orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CancelledOrder{}, ErrOrderNotFound
	}
	if err != nil {
		return CancelledOrder{}, err
	}

	// 2. Check the user may cancel it.
	if !canCancel(order, userID, role) {
		return CancelledOrder{}, ErrNotYourOrder
	}

	// 3. Check the status still allows it.
	if order.Status == models.OrderStatusCancelled {
		return CancelledOrder{}, ErrOrderAlreadyCancelled
	}
	if !order.Status.CanBeCancelled() {
		return CancelledOrder{}, ErrOrderNotCancellable
	}

	// 4. Cancel it.
	err = s.orderRepo.CancelOrder(ctx, order.ID)
	if errors.Is(err, repository.ErrOrderNotCancellable) {
		// The status changed between step 1 and step 4.
		return CancelledOrder{}, ErrOrderNotCancellable
	}
	if err != nil {
		return CancelledOrder{}, err
	}

	return CancelledOrder{ID: order.ID, OrderNumber: order.OrderNumber, Status: models.OrderStatusCancelled}, nil
}

// canCancel: admins can cancel any order, customers their own orders,
// and managers the orders of the restaurant they manage.
func canCancel(order *models.Order, userID uint, role models.Role) bool {
	switch role {
	case models.RoleAdmin:
		return true
	case models.RoleCustomer:
		return order.CustomerID == userID
	case models.RoleManager:
		return order.Restaurant.ManagerID == userID
	default:
		return false
	}
}
