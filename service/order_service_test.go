package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"
)

type fakeOrderRepo struct {
	err   error
	saved *models.Order
}

func (f *fakeOrderRepo) PlaceOrder(_ context.Context, _ uint, order *models.Order) (*models.Order, error) {
	copied := *order
	f.saved = &copied
	if f.err != nil {
		return nil, f.err
	}
	order.OrderNumber = "ORD-20261029-001"
	order.Status = models.OrderStatusPending
	order.DeliveryFee = 0
	order.Subtotal = 1250
	order.Total = 1250
	order.Items = []models.OrderItem{{
		Name: "Margherita", UnitPrice: 1250, Quantity: 1, LineTotal: 1250,
	}}
	return order, nil
}

func (f *fakeOrderRepo) FindOrderByID(context.Context, uint) (*models.Order, error) {
	return nil, errors.New("not used in place order tests")
}

func (f *fakeOrderRepo) CancelOrder(context.Context, uint) error {
	return errors.New("not used in place order tests")
}

var _ repository.OrderRepository = (*fakeOrderRepo)(nil)

func validDelivery() PlaceOrderRequest {
	return PlaceOrderRequest{
		DeliveryName:    "  Ada Lovelace  ",
		DeliveryPhone:   "08012345678",
		DeliveryAddress: "12 Broad Street",
		DeliveryCity:    "Lagos",
		DeliveryState:   "Lagos",
	}
}

func TestPlaceOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("trims the delivery name and ignores a client total", func(t *testing.T) {
		repo := &fakeOrderRepo{}
		placed, err := NewOrderService(repo).PlaceOrder(ctx, 4, validDelivery())
		if err != nil {
			t.Fatalf("place order: %v", err)
		}
		if repo.saved.DeliveryName != "Ada Lovelace" || repo.saved.DeliveryFee != 0 {
			t.Fatalf("saved = %+v", repo.saved)
		}
		if placed.OrderNumber != "ORD-20261029-001" || placed.Total != 1250 || placed.Items[0].UnitPrice != 1250 {
			t.Fatalf("placed = %+v", placed)
		}
	})

	t.Run("blank name", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryName = "   "
		_, err := NewOrderService(&fakeOrderRepo{}).PlaceOrder(ctx, 4, req)
		var validationErr *OrderValidationError
		if !errors.As(err, &validationErr) || validationErr.Fields["delivery_name"] != "name is required" {
			t.Fatalf("blank name: %v", err)
		}
	})

	t.Run("invalid phone", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryPhone = "123"
		_, err := NewOrderService(&fakeOrderRepo{}).PlaceOrder(ctx, 4, req)
		var validationErr *OrderValidationError
		if !errors.As(err, &validationErr) || validationErr.Fields["delivery_phone"] != "Enter a valid phone number" {
			t.Fatalf("phone: %v", err)
		}
	})

	t.Run("phone with symbols is accepted", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryPhone = "+234 (801) 234-5678"
		repo := &fakeOrderRepo{}
		if _, err := NewOrderService(repo).PlaceOrder(ctx, 4, req); err != nil {
			t.Fatalf("formatted phone: %v", err)
		}
		if repo.saved.DeliveryPhone != "+234 (801) 234-5678" {
			t.Fatalf("phone = %q", repo.saved.DeliveryPhone)
		}
	})

	t.Run("phone with letters", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryPhone = "0801-CALL-NOW"
		_, err := NewOrderService(&fakeOrderRepo{}).PlaceOrder(ctx, 4, req)
		var validationErr *OrderValidationError
		if !errors.As(err, &validationErr) || validationErr.Fields["delivery_phone"] == "" {
			t.Fatalf("letter phone: %v", err)
		}
	})

	t.Run("instructions too long", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryInstructions = strings.Repeat("a", 201)
		_, err := NewOrderService(&fakeOrderRepo{}).PlaceOrder(ctx, 4, req)
		var validationErr *OrderValidationError
		if !errors.As(err, &validationErr) || validationErr.Fields["delivery_instructions"] == "" {
			t.Fatalf("long instructions: %v", err)
		}
	})

	t.Run("empty cart", func(t *testing.T) {
		_, err := NewOrderService(&fakeOrderRepo{err: repository.ErrEmptyCart}).PlaceOrder(ctx, 4, validDelivery())
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("empty cart: %v", err)
		}
	})

	t.Run("inactive restaurant", func(t *testing.T) {
		_, err := NewOrderService(&fakeOrderRepo{err: repository.ErrInactiveRestaurant}).PlaceOrder(ctx, 4, validDelivery())
		if !errors.Is(err, ErrCheckoutRestaurantInactive) {
			t.Fatalf("inactive restaurant: %v", err)
		}
	})

	t.Run("unavailable item", func(t *testing.T) {
		_, err := NewOrderService(&fakeOrderRepo{err: repository.ErrUnavailableItem}).PlaceOrder(ctx, 4, validDelivery())
		if !errors.Is(err, ErrCheckoutItemUnavailable) {
			t.Fatalf("unavailable item: %v", err)
		}
	})

	t.Run("name too long", func(t *testing.T) {
		req := validDelivery()
		req.DeliveryName = strings.Repeat("a", 101)
		_, err := NewOrderService(&fakeOrderRepo{}).PlaceOrder(ctx, 4, req)
		var validationErr *OrderValidationError
		if !errors.As(err, &validationErr) || validationErr.Fields["delivery_name"] == "" {
			t.Fatalf("long name: %v", err)
		}
	})
}
