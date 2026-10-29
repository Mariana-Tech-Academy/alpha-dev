package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

func (f *fakeCartRepo) FindCartByUserID(context.Context, uint) (*models.Cart, error) {
	return nil, gorm.ErrRecordNotFound
}

type fakeCheckoutCartRepo struct {
	cart *models.Cart
	err  error
}

func (f *fakeCheckoutCartRepo) FindCartByUserID(context.Context, uint) (*models.Cart, error) {
	return f.cart, f.err
}

func (f *fakeCheckoutCartRepo) AddItem(context.Context, uint, uint, uint, int, int64) (*models.CartItem, error) {
	return nil, errors.New("not implemented")
}

var _ repository.CartRepository = (*fakeCheckoutCartRepo)(nil)

func cartLine(menuItemID uint, name string, qty int, price int64, restaurantID uint) models.CartItem {
	item := models.CartItem{
		MenuItemID: menuItemID,
		MenuItem:   models.MenuItem{Name: name, Available: true, RestaurantID: restaurantID},
		Quantity:   qty,
		Price:      price,
	}
	item.MenuItem.ID = menuItemID
	return item
}

func TestCheckoutSummary(t *testing.T) {
	restaurantID := uint(3)
	restaurant := &models.Restaurant{Name: "Mama Put", Status: models.RestaurantStatusActive}
	restaurant.ID = restaurantID

	t.Run("totals the cart line prices and the delivery fee", func(t *testing.T) {
		cart := &models.Cart{
			RestaurantID: &restaurantID,
			Restaurant:   restaurant,
			Items: []models.CartItem{
				cartLine(1, "Jollof Rice", 2, 450000, restaurantID),
				cartLine(2, "Fried Chicken", 1, 550000, restaurantID),
			},
		}
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
		if len(got.Items) != 2 || got.Items[0].LineTotal != 900000 || got.Items[0].UnitPrice != 450000 || got.Items[0].Name != "Jollof Rice" {
			t.Fatalf("items = %+v", got.Items)
		}
	})

	t.Run("no cart is empty", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{err: gorm.ErrRecordNotFound}, 0).Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("cart with no items is empty", func(t *testing.T) {
		cart := &models.Cart{RestaurantID: &restaurantID, Restaurant: restaurant}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{}, 0).Summary(context.Background(), 0)
		if !errors.Is(err, ErrInvalidUser) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("inactive restaurant", func(t *testing.T) {
		pending := *restaurant
		pending.Status = models.RestaurantStatusPending
		cart := &models.Cart{
			RestaurantID: &restaurantID,
			Restaurant:   &pending,
			Items:        []models.CartItem{cartLine(1, "Jollof Rice", 1, 450000, restaurantID)},
		}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).Summary(context.Background(), 4)
		if !errors.Is(err, ErrCheckoutRestaurantInactive) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("unavailable item", func(t *testing.T) {
		line := cartLine(1, "Jollof Rice", 1, 450000, restaurantID)
		line.MenuItem.Available = false
		cart := &models.Cart{
			RestaurantID: &restaurantID,
			Restaurant:   restaurant,
			Items:        []models.CartItem{line},
		}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).Summary(context.Background(), 4)
		if !errors.Is(err, ErrCheckoutItemUnavailable) {
			t.Fatalf("err = %v", err)
		}
	})
}
