package service

import (
	"alpha-dev/models"
	"context"
	"encoding/json"
	"gorm.io/gorm"
	"testing"
)

func TestBasketAndCheckoutShareDeliveryPolicy(t *testing.T) {
	id := uint(3)
	restaurant := &models.Restaurant{Model: gorm.Model{ID: id}, Status: models.RestaurantStatusActive}
	cart := &models.Cart{RestaurantID: &id, Restaurant: restaurant, Items: []models.CartItem{cartLine(1, "Burger", 2, 1250, id), cartLine(2, "Side", 3, 450, id)}}
	repo := &fakeCheckoutCartRepo{cart: cart}
	basket, err := NewCartService(repo, nil).GetCartView(context.Background(), 4)
	if err != nil {
		t.Fatal(err)
	}
	summary, err := NewCheckoutService(repo).Summary(context.Background(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if basket.Subtotal != 3850 || basket.DeliveryFee != 299 || basket.Total != 4149 || summary.Subtotal != basket.Subtotal || summary.DeliveryFee != basket.DeliveryFee || summary.Total != basket.Total {
		t.Fatalf("basket %+v checkout %+v", basket, summary)
	}
	raw, err := json.Marshal(basket)
	if err != nil {
		t.Fatal(err)
	}
	var response map[string]any
	if err = json.Unmarshal(raw, &response); err != nil {
		t.Fatal(err)
	}
	if response["delivery_fee"] != float64(299) || response["total"] != float64(4149) {
		t.Fatal("cart API pricing fields missing")
	}
	cart.Items = nil
	empty, err := NewCartService(repo, nil).GetCartView(context.Background(), 4)
	if err != nil || empty.Subtotal != 0 || empty.DeliveryFee != 0 || empty.Total != 0 {
		t.Fatalf("empty cart %+v %v", empty, err)
	}
	absent, err := NewCartService(&fakeCheckoutCartRepo{err: gorm.ErrRecordNotFound}, nil).GetCartView(context.Background(), 4)
	if err != nil || absent.Subtotal != 0 || absent.DeliveryFee != 0 || absent.Total != 0 {
		t.Fatal("missing cart has a fee")
	}
}
func TestHistoryRetainsStoredZeroAndNewDeliveryFees(t *testing.T) {
	repo := &orderHistoryRepoStub{total: 2, orders: []models.Order{
		{Model: gorm.Model{ID: 1}, Subtotal: 2500, DeliveryFee: 0, Total: 2500},
		{Model: gorm.Model{ID: 2}, Subtotal: 2500, DeliveryFee: 299, Total: 2799},
	}}
	history, err := NewOrderService(repo).GetOrderHistory(context.Background(), 7, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(history.Orders) != 2 || history.Orders[0].DeliveryFee != 0 || history.Orders[0].Total != 2500 || history.Orders[1].DeliveryFee != 299 || history.Orders[1].Total != 2799 {
		t.Fatal("history repriced saved orders")
	}
}
