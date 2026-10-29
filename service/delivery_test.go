package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"alpha-dev/models"
)

func validDeliveryOrder() *models.Order {
	return &models.Order{
		DeliveryName:         "Ahmed Hussain",
		DeliveryPhone:        "08012345678",
		DeliveryAddress:      "25 Allen Avenue",
		DeliveryCity:         "Ikeja",
		DeliveryState:        "Lagos",
		DeliveryInstructions: "Call when you arrive.",
	}
}

func TestValidateDelivery(t *testing.T) {
	tests := []struct {
		name      string
		change    func(*models.Order)
		wantField string // "" means valid
	}{
		{"valid details", func(*models.Order) {}, ""},
		{"instructions are optional", func(o *models.Order) { o.DeliveryInstructions = "" }, ""},
		{"phone with symbols", func(o *models.Order) { o.DeliveryPhone = "+234 (801) 234-5678" }, ""},

		{"missing name", func(o *models.Order) { o.DeliveryName = "" }, "delivery_name"},
		{"missing phone", func(o *models.Order) { o.DeliveryPhone = "" }, "delivery_phone"},
		{"missing address", func(o *models.Order) { o.DeliveryAddress = "" }, "delivery_address"},
		{"missing city", func(o *models.Order) { o.DeliveryCity = "" }, "delivery_city"},
		{"missing state", func(o *models.Order) { o.DeliveryState = "" }, "delivery_state"},
		{"whitespace-only address", func(o *models.Order) { o.DeliveryAddress = "   " }, "delivery_address"},

		{"phone with letters", func(o *models.Order) { o.DeliveryPhone = "0801-CALL-NOW" }, "delivery_phone"},
		{"phone too short", func(o *models.Order) { o.DeliveryPhone = "12345" }, "delivery_phone"},
		{"phone too long", func(o *models.Order) { o.DeliveryPhone = strings.Repeat("1", 31) }, "delivery_phone"},

		{"name too long", func(o *models.Order) { o.DeliveryName = strings.Repeat("a", 101) }, "delivery_name"},
		{"address too long", func(o *models.Order) { o.DeliveryAddress = strings.Repeat("a", 256) }, "delivery_address"},
		{"instructions too long", func(o *models.Order) { o.DeliveryInstructions = strings.Repeat("a", 501) }, "delivery_instructions"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := validDeliveryOrder()
			tt.change(order)
			trimDelivery(order)
			fields := validateDelivery(order)

			if tt.wantField == "" {
				if len(fields) != 0 {
					t.Fatalf("expected no errors, got %v", fields)
				}
				return
			}
			if _, ok := fields[tt.wantField]; !ok || len(fields) != 1 {
				t.Fatalf("expected only %q to fail, got %v", tt.wantField, fields)
			}
		})
	}
}

type fakeOrderRepo struct {
	gotOrder *models.Order
}

func (f *fakeOrderRepo) PlaceOrder(_ context.Context, _ uint, order *models.Order) (*models.Order, error) {
	f.gotOrder = order
	return order, nil
}

func TestPlaceOrderDelivery(t *testing.T) {
	t.Run("saves trimmed details", func(t *testing.T) {
		repo := &fakeOrderRepo{}
		order := validDeliveryOrder()
		order.DeliveryName = "  Ahmed Hussain  "

		if _, err := NewOrderService(repo).PlaceOrder(context.Background(), 7, order); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.gotOrder.DeliveryName != "Ahmed Hussain" {
			t.Errorf("name not trimmed: %q", repo.gotOrder.DeliveryName)
		}
	})

	t.Run("invalid details are not saved", func(t *testing.T) {
		repo := &fakeOrderRepo{}
		order := validDeliveryOrder()
		order.DeliveryPhone = "not a phone"

		_, err := NewOrderService(repo).PlaceOrder(context.Background(), 7, order)

		var validationError *DeliveryValidationError
		if !errors.As(err, &validationError) {
			t.Fatalf("expected DeliveryValidationError, got %v", err)
		}
		if repo.gotOrder != nil {
			t.Error("repository must not be called when validation fails")
		}
	})
}
