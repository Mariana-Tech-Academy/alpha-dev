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

func (f *fakeCartRepo) UpdateQuantity(context.Context, uint, uint, int) error {
	return errors.New("not used by add-item tests")
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

// Change an item only when the test cart belongs to this customer.
func (f *fakeCheckoutCartRepo) UpdateQuantity(
	_ context.Context,
	userID uint,
	itemID uint,
	quantity int,
) error {
	if f.err != nil {
		return f.err
	}
	if f.cart == nil || f.cart.UserID != userID {
		return gorm.ErrRecordNotFound
	}

	for i := range f.cart.Items {
		if f.cart.Items[i].ID == itemID {
			f.cart.Items[i].Quantity = quantity
			return nil
		}
	}

	return gorm.ErrRecordNotFound
}

var _ repository.CartRepository = (*fakeCheckoutCartRepo)(nil)

func cartLine(menuItemID uint, name string, qty int, price int64, restaurantID uint) models.CartItem {
	item := models.CartItem{
		MenuItemID: menuItemID,
		MenuItem: models.MenuItem{
			Name:         name,
			Available:    true,
			RestaurantID: restaurantID,
		},
		Quantity: qty,
		Price:    price,
	}
	item.MenuItem.ID = menuItemID
	return item
}

func TestCheckoutSummary(t *testing.T) {
	restaurantID := uint(3)
	restaurant := &models.Restaurant{
		Name:   "Mama Put",
		Status: models.RestaurantStatusActive,
	}
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
		got, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 200000).
			Summary(context.Background(), 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Subtotal != 1450000 || got.DeliveryFee != 200000 || got.Total != 1650000 {
			t.Fatalf("summary = %+v", got)
		}
		if got.Restaurant.ID != 3 || got.Restaurant.Name != "Mama Put" {
			t.Fatalf("restaurant = %+v", got.Restaurant)
		}
		if len(got.Items) != 2 ||
			got.Items[0].LineTotal != 900000 ||
			got.Items[0].UnitPrice != 450000 ||
			got.Items[0].Name != "Jollof Rice" {
			t.Fatalf("items = %+v", got.Items)
		}
	})

	t.Run("no cart is empty", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{err: gorm.ErrRecordNotFound}, 0).
			Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("cart with no items is empty", func(t *testing.T) {
		cart := &models.Cart{
			RestaurantID: &restaurantID,
			Restaurant:   restaurant,
		}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).
			Summary(context.Background(), 4)
		if !errors.Is(err, ErrEmptyCart) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("invalid user", func(t *testing.T) {
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{}, 0).
			Summary(context.Background(), 0)
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
			Items: []models.CartItem{
				cartLine(1, "Jollof Rice", 1, 450000, restaurantID),
			},
		}
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).
			Summary(context.Background(), 4)
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
		_, err := NewCheckoutService(&fakeCheckoutCartRepo{cart: cart}, 0).
			Summary(context.Background(), 4)
		if !errors.Is(err, ErrCheckoutItemUnavailable) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestUpdateCartQuantity(t *testing.T) {
	restaurantID := uint(3)
	item := cartLine(9, "Pizza", 2, 1000, restaurantID)
	item.ID = 7

	cart := &models.Cart{
		UserID:       4,
		RestaurantID: &restaurantID,
		Restaurant: &models.Restaurant{
			Status: models.RestaurantStatusActive,
		},
		Items: []models.CartItem{item},
	}
	svc := NewCartService(&fakeCheckoutCartRepo{cart: cart}, nil)

	// Increasing or decreasing quantity updates both totals.
	for _, quantity := range []int{5, 1} {
		view, err := svc.UpdateQuantity(context.Background(), 4, 7, quantity)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		row := view.Items[0]
		expected := int64(quantity) * 1000
		if row.ID != 7 || row.Quantity != quantity ||
			row.LockedPrice != 1000 ||
			row.LineTotal != expected || view.Subtotal != expected {
			t.Fatalf("unexpected cart: %+v", view)
		}
	}

	// Invalid quantities must not change the saved quantity.
	for _, quantity := range []int{0, -1, 100} {
		_, err := svc.UpdateQuantity(context.Background(), 4, 7, quantity)
		if !errors.Is(err, ErrInvalidQuantity) &&
			!errors.Is(err, ErrCartQuantityLimit) {
			t.Fatalf("expected quantity error, got %v", err)
		}
		if cart.Items[0].Quantity != 1 {
			t.Fatal("invalid request changed the quantity")
		}
	}

	// Another customer cannot change this cart.
	_, err := svc.UpdateQuantity(context.Background(), 8, 7, 5)
	if !errors.Is(err, ErrCartItemNotFound) {
		t.Fatalf("expected item not found, got %v", err)
	}
	if cart.Items[0].Quantity != 1 {
		t.Fatal("another customer's quantity changed")
	}

	// An unknown cart item must return not found.
	_, err = svc.UpdateQuantity(context.Background(), 4, 99, 5)
	if !errors.Is(err, ErrCartItemNotFound) {
		t.Fatalf("expected item not found, got %v", err)
	}
}
