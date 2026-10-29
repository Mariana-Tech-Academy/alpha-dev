package service

import (
	"context"
	"errors"
	"strings"

	"alpha-dev/models"
	"alpha-dev/repository"
)

var (
	ErrEmptyCart                  = errors.New("your cart is empty")
	ErrCheckoutRestaurantInactive = errors.New("restaurant is not active")
	ErrCheckoutItemUnavailable    = errors.New("one or more menu items are unavailable")
)

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

type OrderService interface {
	PlaceOrder(ctx context.Context, userID uint, req PlaceOrderRequest) (PlacedOrder, error)
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
