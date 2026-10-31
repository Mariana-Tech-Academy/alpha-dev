package service

import (
	"alpha-dev/repository"
	"gorm.io/gorm"
	"context"
	"errors"
	"fmt"

	"alpha-dev/models"
)

//grouped declaration of error values
var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderForbidden     = errors.New("you cannot manage orders for this restaurant")
	ErrInvalidOrderStatus = errors.New("status must be one of: pending, confirmed, preparing, ready_for_delivery, out_for_delivery, delivered, cancelled")
	ErrOrderChanged       = errors.New("order changed; refresh and try again")
)

type OrderTransitionError struct {
	From    models.OrderStatus
	To      models.OrderStatus
	Allowed models.OrderStatus // empty when From is a final status
}

//In Go, an error is anything that has a method called Error() that returns a string, Go's internal interface
//the receiver: this method belongs to OrderTransitionError, and inside it we call the object e
//e.Allowed is empty when From is a final status
func (e *OrderTransitionError) Error() string {
	if e.Allowed == "" {
		return fmt.Sprintf("order is %s; no further status change is allowed", e.From)
	}
	return fmt.Sprintf("cannot change order from %s to %s; the only allowed next status is %s", e.From, e.To, e.Allowed)
}

type OrderService interface {
	UpdateOrderStatus(ctx context.Context, userID uint, role models.Role, orderID uint, next models.OrderStatus) (*models.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

func (s *orderService) UpdateOrderStatus(
	ctx context.Context,
	userID uint,
	role models.Role,
	orderID uint,
	next models.OrderStatus,
) (*models.Order, error) {
	//userID == 0 || (role is neither manager nor admin) returns forbidden.
	if userID == 0 || (role != models.RoleManager && role != models.RoleAdmin) {
		return nil, ErrOrderForbidden
	}
	if !next.IsValid() {
		//!next.IsValid() rejects nonsense before touching the database, which is cheap and fast.
		return nil, ErrInvalidOrderStatus
	}

	order, err := s.orderRepo.FindByID(ctx, orderID)
	//errors.Is(err, gorm.ErrRecordNotFound) turns a database-level error into our ErrOrderNotFound
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	//Any other error is passed up unchanged.
	if err != nil {
		return nil, err
	}

	if role == models.RoleManager && order.Restaurant.ManagerID != userID {
		//manager cannot update another restaurant’s order (403)” criterion
		return nil, ErrOrderForbidden
	}

	allowed, ok := order.Status.Next()
	if !ok || next != allowed {
		//!ok (the order is final) or next != allowed (the manager tried to skip or go backwards). Both return an OrderTransitionError.
		return nil, &OrderTransitionError{From: order.Status, To: next, Allowed: allowed}
	}

	updated, err := s.orderRepo.UpdateStatus(ctx, order.ID, order.Status, next)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, ErrOrderChanged
	}

	fresh, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	return &fresh, nil
}





