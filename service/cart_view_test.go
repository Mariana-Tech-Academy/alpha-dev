package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"

	"gorm.io/gorm"
)

func TestGetCartView(t *testing.T) {
	restaurantID := uint(3)
	restaurant := &models.Restaurant{Status: models.RestaurantStatusActive}
	restaurant.ID = restaurantID

	burger := models.MenuItem{Name: "Chicken Burger", Price: 1000, Available: true, RestaurantID: restaurantID}
	burger.ID = 1
	pizza := models.MenuItem{Name: "Pizza", Price: 450, Available: false, RestaurantID: restaurantID}
	pizza.ID = 2

	cart := &models.Cart{
		UserID:       4,
		RestaurantID: &restaurantID,
		Restaurant:   restaurant,
		Items: []models.CartItem{
			{MenuItemID: 1, MenuItem: burger, Quantity: 2, Price: 850},
			{MenuItemID: 2, MenuItem: pizza, Quantity: 1, Price: 450},
		},
	}
	cart.ID = 8

	view, err := NewCartService(&fakeCheckoutCartRepo{cart: cart}, &fakeMenuRepo{}).GetCartView(context.Background(), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.CartID != 8 || view.RestaurantID == nil || *view.RestaurantID != 3 || view.Subtotal != 1700 {
		t.Fatalf("cart = %+v", view)
	}
	if len(view.Items) != 2 {
		t.Fatalf("items = %+v", view.Items)
	}
	burgerRow := view.Items[0]
	if !burgerRow.IsAvailable || !burgerRow.PriceChanged || burgerRow.LockedPrice != 850 || burgerRow.Price != 1000 || burgerRow.LineTotal != 1700 || burgerRow.Notice != "Price changed from 850 to 1000." {
		t.Fatalf("burger = %+v", burgerRow)
	}
	pizzaRow := view.Items[1]
	if pizzaRow.IsAvailable || pizzaRow.LineTotal != 0 || pizzaRow.Notice != "This item is currently unavailable." {
		t.Fatalf("pizza = %+v", pizzaRow)
	}
}

func TestGetCartViewEmptyAndInvalid(t *testing.T) {
	t.Run("missing cart", func(t *testing.T) {
		view, err := NewCartService(&fakeCheckoutCartRepo{err: gorm.ErrRecordNotFound}, &fakeMenuRepo{}).GetCartView(context.Background(), 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.Items == nil || len(view.Items) != 0 || view.Subtotal != 0 {
			t.Fatalf("cart = %+v", view)
		}
	})

	t.Run("missing customer", func(t *testing.T) {
		_, err := NewCartService(&fakeCheckoutCartRepo{}, &fakeMenuRepo{}).GetCartView(context.Background(), 0)
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("expected ErrInvalidUser, got %v", err)
		}
	})

	t.Run("inactive restaurant", func(t *testing.T) {
		restaurantID := uint(3)
		restaurant := &models.Restaurant{Status: models.RestaurantStatusPending}
		restaurant.ID = restaurantID
		menu := models.MenuItem{Name: "Burger", Price: 850, Available: true, RestaurantID: restaurantID}
		menu.ID = 1
		cart := &models.Cart{
			RestaurantID: &restaurantID,
			Restaurant:   restaurant,
			Items:        []models.CartItem{{MenuItemID: 1, MenuItem: menu, Quantity: 1, Price: 850}},
		}
		view, err := NewCartService(&fakeCheckoutCartRepo{cart: cart}, &fakeMenuRepo{}).GetCartView(context.Background(), 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if view.Subtotal != 0 || view.Items[0].IsAvailable || view.Items[0].Notice != "This item is currently unavailable." {
			t.Fatalf("cart = %+v", view)
		}
	})

	t.Run("database error", func(t *testing.T) {
		databaseErr := errors.New("database unavailable")
		_, err := NewCartService(&fakeCheckoutCartRepo{err: databaseErr}, &fakeMenuRepo{}).GetCartView(context.Background(), 4)
		if !errors.Is(err, databaseErr) {
			t.Fatalf("expected database error, got %v", err)
		}
	})
}
