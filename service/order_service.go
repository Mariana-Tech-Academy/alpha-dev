package service

import (
	"context"
	"errors"
	"time"

	"alpha-dev/models"
	"alpha-dev/repository"
)

var (
	ErrOrderHistoryUnauthenticated = errors.New("unauthenticated")
	ErrOrderHistoryPagination      = errors.New(
		"page must be positive and limit must be between 1 and 100",
	)
)

// Details shown for each order in the history.
type OrderHistoryItem struct {
	ID             uint               `json:"id"`
	OrderNumber    string             `json:"order_number"`
	RestaurantID   uint               `json:"restaurant_id"`
	RestaurantName string             `json:"restaurant_name"`
	Total          int64              `json:"total"`
	Status         models.OrderStatus `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
}

// Include page information with the orders.
type OrderHistoryResponse struct {
	Orders []OrderHistoryItem `json:"orders"`
	Page   int                `json:"page"`
	Limit  int                `json:"limit"`
	Total  int64              `json:"total"`
}

type OrderService interface {
	PlaceOrder(
		ctx context.Context,
		userID uint,
		order *models.Order,
	) (*models.Order, error)

	GetOrderHistory(
		ctx context.Context,
		customerID uint,
		page int,
		limit int,
	) (*OrderHistoryResponse, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

// Keep the existing checkout feature.
func (s *orderService) PlaceOrder(
	ctx context.Context,
	userID uint,
	order *models.Order,
) (*models.Order, error) {
	return s.orderRepo.PlaceOrder(ctx, userID, order)
}

// Get one page of the logged-in customer's order history.
func (s *orderService) GetOrderHistory(
	ctx context.Context,
	customerID uint,
	page int,
	limit int,
) (*OrderHistoryResponse, error) {
	if customerID == 0 {
		return nil, ErrOrderHistoryUnauthenticated
	}

	// Reject invalid pagination and numbers too large to calculate safely.
	maxInt := int(^uint(0) >> 1)
	if page < 1 || limit < 1 || limit > 100 {
		return nil, ErrOrderHistoryPagination
	}
	if page-1 > maxInt/limit {
		return nil, ErrOrderHistoryPagination
	}

	offset := (page - 1) * limit

	orders, total, err := s.orderRepo.FindCustomerOrders(
		ctx, customerID, limit, offset,
	)
	if err != nil {
		return nil, err
	}

	// Return only the details needed by the order-history list.
	items := make([]OrderHistoryItem, 0, len(orders))

	for _, order := range orders {
		items = append(items, OrderHistoryItem{
			ID:             order.ID,
			OrderNumber:    order.OrderNumber,
			RestaurantID:   order.RestaurantID,
			RestaurantName: order.Restaurant.Name,
			Total:          order.Total,
			Status:         order.Status,
			CreatedAt:      order.CreatedAt,
		})
	}

	return &OrderHistoryResponse{
		Orders: items,
		Page:   page,
		Limit:  limit,
		Total:  total,
	}, nil
}
