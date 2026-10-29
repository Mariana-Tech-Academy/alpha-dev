package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

// fakeCartRepo (cart_service_test.go) must satisfy the full CartRepository interface.
func (f *fakeCartRepo) FindCartByUserID(context.Context, uint) (*models.Cart, error) {
	return nil, gorm.ErrRecordNotFound
}

//pretend store room
//holds the cart and error we want to hand back
type fakeCheckoutCartRepo struct {
	repository.CartRepository
	cart *models.Cart
	err  error
}

func (f *fakeCheckoutCartRepo) FindCartByUserID(context.Context, uint) (*models.Cart, error) {
	return f.cart, f.err
}

func cartLine(menuItemID uint, name string, qty int, price int64) models.CartItem {
	return models.CartItem{
		MenuItemID: menuItemID,
		MenuItem:   models.MenuItem{Name: name},
		Quantity:   qty,
		Price:      price,
	}
}

func TestCheckoutSummary(t *testing.T) {
	restaurant := &models.Restaurant{Name: "Mama Put"}
	restaurant.ID = 3

	t.Run("totals subtotal, delivery fee and total", func(t *testing.T) {
		// N14,500 subtotal + N2,000 delivery = N16,500 (values in kobo).
		cart := &models.Cart{Restaurant: restaurant, Items: []models.CartItem{
			cartLine(1, "Jollof Rice", 2, 450000),
			cartLine(2, "Fried Chicken", 1, 550000),
		}}
		got, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 200000).Summary(context.Background(), 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Subtotal != 1450000 || got.DeliveryFee != 200000 || got.Total != 1650000 {
			t.Fatalf("summary = %+v", got)
		}
		if got.Restaurant.ID != 3 || got.Restaurant.Name != "Mama Put" {
			t.Fatalf("restaurant = %+v", got.Restaurant)
		}
		if len(got.Items) != 2 || got.Items[0].LineTotal != 900000 || got.Items[0].Quantity != 2 || got.Items[0].Name != "Jollof Rice" {
			t.Fatalf("items = %+v", got.Items)
		}
	})

	t.Run("no cart is treated as empty", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{err: gorm.ErrRecordNotFound}, 200000).Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v, want ErrEmptyCart", err)
		}
	})

	t.Run("cart with no items is blocked", func(t *testing.T) {
		cart := &models.Cart{Restaurant: restaurant}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 200000).Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v, want ErrEmptyCart", err)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{}, 200000).Summary(context.Background(), 0)
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("err = %v, want ErrInvalidUser", err)
		}
	})

	t.Run("repository failure is returned", func(t *testing.T) {
		boom := errors.New("db down")
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{err: boom}, 200000).Summary(context.Background(), 4)
		if !errors.Is(err, boom) {
			t.Fatalf("err = %v, want db error", err)
		}
	})
}