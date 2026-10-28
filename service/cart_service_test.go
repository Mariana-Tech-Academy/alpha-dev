package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

type fakeCartRepo struct {
	restaurantID *uint
	items        map[uint]models.CartItem
}

func (f *fakeCartRepo) AddItem(_ context.Context, _ uint, menuItemID uint, restaurantID uint, quantity int, price int64) (*models.CartItem, error) {
	if f.items == nil {
		f.items = map[uint]models.CartItem{}
	}
	if f.restaurantID != nil && *f.restaurantID != restaurantID {
		return nil, repository.ErrDifferentRestaurant
	}
	if f.restaurantID == nil {
		id := restaurantID
		f.restaurantID = &id
	}

	if existing, ok := f.items[menuItemID]; ok {
		next := existing.Quantity + quantity
		if next > repository.MaxCartLineQuantity {
			return nil, repository.ErrCartLineFull
		}
		existing.Quantity = next
		f.items[menuItemID] = existing
		return &existing, nil
	}

	if quantity > repository.MaxCartLineQuantity {
		return nil, repository.ErrCartLineFull
	}
	item := models.CartItem{
		CartID:     1,
		MenuItemID: menuItemID,
		Quantity:   quantity,
		Price:      price,
	}
	item.ID = uint(len(f.items) + 1)
	f.items[menuItemID] = item
	return &item, nil
}

func TestAddCartItem(t *testing.T) {
	active := models.Restaurant{Status: models.RestaurantStatusActive}
	item := models.MenuItem{RestaurantID: 3, Price: 1250, Available: true}
	item.ID = 9

	t.Run("adds a new line at the menu price", func(t *testing.T) {
		carts := &fakeCartRepo{}
		added, err := NewCartService(carts, &fakeMenuRepo{restaurant: active, item: item}).AddItem(context.Background(), 4, 9, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if added.Quantity != 2 || added.Price != 1250 || added.MenuItemID != 9 {
			t.Fatalf("item = %+v", added)
		}
	})

	t.Run("second add increases quantity and keeps the price", func(t *testing.T) {
		carts := &fakeCartRepo{}
		menus := &fakeMenuRepo{restaurant: active, item: item}
		svc := NewCartService(carts, menus)
		if _, err := svc.AddItem(context.Background(), 4, 9, 2); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		menus.item.Price = 2000
		added, err := svc.AddItem(context.Background(), 4, 9, 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if added.Quantity != 5 || added.Price != 1250 {
			t.Fatalf("item = %+v", added)
		}
	})

	t.Run("quantity below 1", func(t *testing.T) {
		_, err := NewCartService(&fakeCartRepo{}, &fakeMenuRepo{restaurant: active, item: item}).AddItem(context.Background(), 4, 9, 0)
		if !errors.Is(err, ErrInvalidQuantity) {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("missing item", func(t *testing.T) {
		_, err := NewCartService(&fakeCartRepo{}, &fakeMenuRepo{findItemErr: gorm.ErrRecordNotFound}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrCartMenuItemNotFound) {
			t.Fatalf("expected ErrCartMenuItemNotFound, got %v", err)
		}
	})

	t.Run("unavailable item", func(t *testing.T) {
		hidden := item
		hidden.Available = false
		_, err := NewCartService(&fakeCartRepo{}, &fakeMenuRepo{restaurant: active, item: hidden}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrItemUnavailable) {
			t.Fatalf("expected ErrItemUnavailable, got %v", err)
		}
	})

	t.Run("restaurant is not active", func(t *testing.T) {
		pending := active
		pending.Status = models.RestaurantStatusPending
		_, err := NewCartService(&fakeCartRepo{}, &fakeMenuRepo{restaurant: pending, item: item}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrItemUnavailable) {
			t.Fatalf("expected ErrItemUnavailable, got %v", err)
		}
	})

	t.Run("cart already holds another restaurant", func(t *testing.T) {
		carts := &fakeCartRepo{}
		menus := &fakeMenuRepo{restaurant: active, item: item}
		svc := NewCartService(carts, menus)
		if _, err := svc.AddItem(context.Background(), 4, 9, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		other := item
		other.ID = 10
		other.RestaurantID = 8
		menus.item = other
		_, err := svc.AddItem(context.Background(), 4, 10, 1)
		if !errors.Is(err, ErrCartDifferentRestaurant) {
			t.Fatalf("expected ErrCartDifferentRestaurant, got %v", err)
		}
	})

	t.Run("line quantity cannot exceed 99", func(t *testing.T) {
		carts := &fakeCartRepo{}
		svc := NewCartService(carts, &fakeMenuRepo{restaurant: active, item: item})
		if _, err := svc.AddItem(context.Background(), 4, 9, 90); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.AddItem(context.Background(), 4, 9, 10)
		if !errors.Is(err, ErrCartQuantityLimit) {
			t.Fatalf("expected ErrCartQuantityLimit, got %v", err)
		}
	})
}
