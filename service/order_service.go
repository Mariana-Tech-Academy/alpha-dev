package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrEmptyCart                  = errors.New("your cart is empty")
	ErrCheckoutRestaurantInactive = errors.New("restaurant is not active")
	ErrCheckoutItemUnavailable    = errors.New("one or more menu items are unavailable")
	ErrOrderNotFound              = errors.New("order not found")
	ErrOrderForbidden             = errors.New("you cannot manage orders for this restaurant")
	ErrInvalidOrderStatus         = errors.New("status must be one of: pending, confirmed, preparing, ready_for_delivery, out_for_delivery, delivered, cancelled")
	ErrOrderChanged               = errors.New("order changed; refresh and try again")
	ErrNoRestaurantManaged        = errors.New("no restaurant assigned to this manager")
)

type OrderTransitionError struct {
	From    models.OrderStatus
	To      models.OrderStatus
	Allowed models.OrderStatus
}

func (e *OrderTransitionError) Error() string {
	if e.Allowed == "" {
		return fmt.Sprintf("order is %s; no further status change is allowed", e.From)
	}
	return fmt.Sprintf("cannot change order from %s to %s; the only allowed next status is %s", e.From, e.To, e.Allowed)
}

const (
	maxDeliveryNameLength         = 100
	maxDeliveryPhoneLength        = 30
	maxDeliveryAddressLength      = 200
	maxDeliveryInstructionsLength = 200
)

type OrderValidationError struct {
	Fields map[string]string
}

func (err *OrderValidationError) Error() string {
	return "invalid order"
}

type PlaceOrderRequest struct {
	DeliveryName         string `json:"delivery_name"`
	DeliveryPhone        string `json:"delivery_phone"`
	DeliveryAddress      string `json:"delivery_address"`
	DeliveryCity         string `json:"delivery_city"`
	DeliveryState        string `json:"delivery_state"`
	DeliveryInstructions string `json:"delivery_instructions"`
}

type PlacedOrderItem struct {
	Name      string `json:"name"`
	UnitPrice int64  `json:"unit_price"`
	Quantity  int    `json:"quantity"`
	LineTotal int64  `json:"line_total"`
}

type PlacedOrder struct {
	OrderNumber string             `json:"order_number"`
	Status      models.OrderStatus `json:"status"`
	Subtotal    int64              `json:"subtotal"`
	DeliveryFee int64              `json:"delivery_fee"`
	Total       int64              `json:"total"`
	Items       []PlacedOrderItem  `json:"items"`
}

type OrderHistoryItem struct {
	Items                []PlacedOrderItem  `json:"items"`
	Subtotal             int64              `json:"subtotal"`
	DeliveryFee          int64              `json:"delivery_fee"`
	DeliveryName         string             `json:"delivery_name,omitempty"`
	DeliveryPhone        string             `json:"delivery_phone,omitempty"`
	DeliveryAddress      string             `json:"delivery_address,omitempty"`
	DeliveryCity         string             `json:"delivery_city,omitempty"`
	DeliveryState        string             `json:"delivery_state,omitempty"`
	DeliveryInstructions string             `json:"delivery_instructions,omitempty"`
	ID                   uint               `json:"id"`
	OrderNumber          string             `json:"order_number"`
	RestaurantID         uint               `json:"restaurant_id"`
	RestaurantName       string             `json:"restaurant_name"`
	Total                int64              `json:"total"`
	Status               models.OrderStatus `json:"status"`
	CreatedAt            time.Time          `json:"created_at"`
}

type OrderHistory struct {
	Orders []OrderHistoryItem `json:"orders"`
	Page   int                `json:"page"`
	Limit  int                `json:"limit"`
	Total  int64              `json:"total"`
}

const maxOrderPageSize = 20

type UpdatedOrder struct {
	ID          uint               `json:"id"`
	OrderNumber string             `json:"order_number"`
	Status      models.OrderStatus `json:"status"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type OrderService interface {
	AdvanceDemoOrder(ctx context.Context, userID uint, role models.Role, orderID uint, expected models.OrderStatus) (UpdatedOrder, error)
	PlaceOrder(ctx context.Context, userID uint, req PlaceOrderRequest) (PlacedOrder, error)
	GetOrderHistory(ctx context.Context, customerID uint, page, limit int) (OrderHistory, error)
	UpdateOrderStatus(ctx context.Context, userID uint, role models.Role, orderID uint, next models.OrderStatus) (UpdatedOrder, error)
	GetManagerOrders(ctx context.Context, managerID uint, statusStr string) ([]models.ManagerOrderResponse, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) OrderService {
	return &orderService{orderRepo: repo}
}

func (s *orderService) PlaceOrder(ctx context.Context, userID uint, req PlaceOrderRequest) (PlacedOrder, error) {
	if userID == 0 {
		return PlacedOrder{}, ErrInvalidUser
	}

	name := strings.TrimSpace(req.DeliveryName)
	phone := strings.TrimSpace(req.DeliveryPhone)
	address := strings.TrimSpace(req.DeliveryAddress)
	city := strings.TrimSpace(req.DeliveryCity)
	state := strings.TrimSpace(req.DeliveryState)
	instructions := strings.TrimSpace(req.DeliveryInstructions)

	fields := map[string]string{}
	if name == "" {
		fields["delivery_name"] = "name is required"
	} else if len([]rune(name)) > maxDeliveryNameLength {
		fields["delivery_name"] = "name must not exceed 100 characters"
	}
	if phone == "" {
		fields["delivery_phone"] = "Phone number is required"
	} else if len([]rune(phone)) > maxDeliveryPhoneLength {
		fields["delivery_phone"] = "Phone number must not exceed 30 characters"
	} else if !validPhone(phone) {
		fields["delivery_phone"] = "Enter a valid phone number"
	}
	if address == "" {
		fields["delivery_address"] = "address is required"
	} else if len([]rune(address)) > maxDeliveryAddressLength {
		fields["delivery_address"] = "address must not exceed 200 characters"
	}
	if city == "" {
		fields["delivery_city"] = "city is required"
	} else if len([]rune(city)) > maxDeliveryNameLength {
		fields["delivery_city"] = "city must not exceed 100 characters"
	}
	if state == "" {
		fields["delivery_state"] = "state is required"
	} else if len([]rune(state)) > maxDeliveryNameLength {
		fields["delivery_state"] = "state must not exceed 100 characters"
	}
	if len([]rune(instructions)) > maxDeliveryInstructionsLength {
		fields["delivery_instructions"] = "instructions must not exceed 200 characters"
	}
	if len(fields) > 0 {
		return PlacedOrder{}, &OrderValidationError{Fields: fields}
	}

	order := &models.Order{
		DeliveryName:         name,
		DeliveryPhone:        phone,
		DeliveryAddress:      address,
		DeliveryCity:         city,
		DeliveryState:        state,
		DeliveryInstructions: instructions,
	}
	placed, err := s.orderRepo.PlaceOrder(ctx, userID, order)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmptyCart):
			return PlacedOrder{}, ErrEmptyCart
		case errors.Is(err, repository.ErrInactiveRestaurant):
			return PlacedOrder{}, ErrCheckoutRestaurantInactive
		case errors.Is(err, repository.ErrUnavailableItem):
			return PlacedOrder{}, ErrCheckoutItemUnavailable
		default:
			return PlacedOrder{}, err
		}
	}

	items := make([]PlacedOrderItem, 0, len(placed.Items))
	for _, item := range placed.Items {
		items = append(items, PlacedOrderItem{
			Name:      item.Name,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			LineTotal: item.LineTotal,
		})
	}
	return PlacedOrder{
		OrderNumber: placed.OrderNumber,
		Status:      placed.Status,
		Subtotal:    placed.Subtotal,
		DeliveryFee: placed.DeliveryFee,
		Total:       placed.Total,
		Items:       items,
	}, nil
}

func (s *orderService) GetOrderHistory(ctx context.Context, customerID uint, page, limit int) (OrderHistory, error) {
	if customerID == 0 {
		return OrderHistory{}, ErrInvalidUser
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > maxOrderPageSize {
		limit = maxOrderPageSize
	}

	orders, total, err := s.orderRepo.FindCustomerOrders(ctx, customerID, limit, (page-1)*limit)
	if err != nil {
		return OrderHistory{}, err
	}

	items := make([]OrderHistoryItem, 0, len(orders))
	for _, order := range orders {
		snapshots := make([]PlacedOrderItem, 0, len(order.Items))
		for _, line := range order.Items {
			snapshots = append(snapshots, PlacedOrderItem{Name: line.Name, UnitPrice: line.UnitPrice, Quantity: line.Quantity, LineTotal: line.LineTotal})
		}
		items = append(items, OrderHistoryItem{
			Items:                snapshots,
			Subtotal:             order.Subtotal,
			DeliveryFee:          order.DeliveryFee,
			DeliveryName:         order.DeliveryName,
			DeliveryPhone:        order.DeliveryPhone,
			DeliveryAddress:      order.DeliveryAddress,
			DeliveryCity:         order.DeliveryCity,
			DeliveryState:        order.DeliveryState,
			DeliveryInstructions: order.DeliveryInstructions,
			ID:                   order.ID,
			OrderNumber:          order.OrderNumber,
			RestaurantID:         order.RestaurantID,
			RestaurantName:       order.Restaurant.Name,
			Total:                order.Total,
			Status:               order.Status,
			CreatedAt:            order.CreatedAt,
		})
	}
	return OrderHistory{
		Orders: items,
		Page:   page,
		Limit:  limit,
		Total:  total,
	}, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, userID uint, role models.Role, orderID uint, next models.OrderStatus) (UpdatedOrder, error) {
	return s.updateStatus(ctx, userID, role, orderID, next, nil)
}

// A separate, gated demo operation uses native persisted values for four stages.
func (s *orderService) AdvanceDemoOrder(ctx context.Context, userID uint, role models.Role, orderID uint, expected models.OrderStatus) (UpdatedOrder, error) {
	next := map[models.OrderStatus]models.OrderStatus{
		models.OrderStatusPending:          models.OrderStatusPreparing,
		models.OrderStatusConfirmed:        models.OrderStatusPreparing,
		models.OrderStatusPreparing:        models.OrderStatusOutForDelivery,
		models.OrderStatusReadyForDelivery: models.OrderStatusOutForDelivery,
		models.OrderStatusOutForDelivery:   models.OrderStatusDelivered,
	}[expected]
	return s.updateStatus(ctx, userID, role, orderID, next, &expected)
}

func (s *orderService) updateStatus(ctx context.Context, userID uint, role models.Role, orderID uint, next models.OrderStatus, expected *models.OrderStatus) (UpdatedOrder, error) {

	if userID == 0 {
		return UpdatedOrder{}, ErrInvalidUser
	}
	if role != models.RoleManager && role != models.RoleAdmin {
		return UpdatedOrder{}, ErrOrderForbidden
	}
	if expected == nil && !next.IsValid() {
		return UpdatedOrder{}, ErrInvalidOrderStatus
	}

	order, err := s.orderRepo.FindByID(ctx, orderID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return UpdatedOrder{}, ErrOrderNotFound
	}
	if err != nil {
		return UpdatedOrder{}, err
	}
	if role == models.RoleManager && order.Restaurant.ManagerID != userID {
		return UpdatedOrder{}, ErrOrderForbidden
	}

	allowed, ok := order.Status.Next()
	if expected != nil {
		if !expected.IsValid() {
			return UpdatedOrder{}, ErrInvalidOrderStatus
		}
		if order.Status != *expected {
			return UpdatedOrder{}, ErrOrderChanged
		}
		allowed = next
		ok = next != ""
	}
	if !ok || next != allowed {
		return UpdatedOrder{}, &OrderTransitionError{From: order.Status, To: next, Allowed: allowed}
	}

	updated, err := s.orderRepo.UpdateStatus(ctx, order.ID, order.Status, next)
	if err != nil {
		return UpdatedOrder{}, err
	}
	if !updated {
		return UpdatedOrder{}, ErrOrderChanged
	}

	fresh, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		return UpdatedOrder{}, err
	}
	return UpdatedOrder{
		ID:          fresh.ID,
		OrderNumber: fresh.OrderNumber,
		Status:      fresh.Status,
		UpdatedAt:   fresh.UpdatedAt,
	}, nil
}

func (s *orderService) GetManagerOrders(ctx context.Context, managerID uint, statusStr string) ([]models.ManagerOrderResponse, error) {
	if managerID == 0 {
		return nil, ErrInvalidUser
	}

	var status models.OrderStatus
	if strings.TrimSpace(statusStr) != "" {
		status = models.OrderStatus(strings.TrimSpace(statusStr))
		if !status.IsValid() {
			return nil, ErrInvalidOrderStatus
		}
	}

	restaurantID, err := s.orderRepo.GetRestaurantIDByManager(ctx, managerID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNoRestaurantManaged
	}
	if err != nil {
		return nil, err
	}

	orders, err := s.orderRepo.GetOrdersByRestaurant(ctx, restaurantID, status)
	if err != nil {
		return nil, err
	}

	response := make([]models.ManagerOrderResponse, 0, len(orders))
	for _, order := range orders {
		items := make([]models.ManagerOrderItemResponse, 0, len(order.Items))
		for _, item := range order.Items {
			items = append(items, models.ManagerOrderItemResponse{
				MenuItemID: item.MenuItemID,
				Name:       item.Name,
				UnitPrice:  item.UnitPrice,
				Quantity:   item.Quantity,
				LineTotal:  item.LineTotal,
			})
		}
		response = append(response, models.ManagerOrderResponse{
			ID:                   order.ID,
			OrderNumber:          order.OrderNumber,
			Status:               order.Status,
			Subtotal:             order.Subtotal,
			DeliveryFee:          order.DeliveryFee,
			Total:                order.Total,
			DeliveryName:         order.DeliveryName,
			DeliveryPhone:        order.DeliveryPhone,
			DeliveryAddress:      order.DeliveryAddress,
			DeliveryCity:         order.DeliveryCity,
			DeliveryState:        order.DeliveryState,
			DeliveryInstructions: order.DeliveryInstructions,
			Items:                items,
			CreatedAt:            order.CreatedAt,
		})
	}
	return response, nil
}
