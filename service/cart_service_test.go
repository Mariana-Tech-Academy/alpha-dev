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
	it := models.CartItem{CartID: 1, MenuItemID: menuItemID, Quantity: quantity, Price: price}
	it.ID = uint(len(f.items) + 1)
	f.items[menuItemID] = it
	return &it, nil
}
func (f *fakeCartRepo) GetCart(_ context.Context, _ uint) (*models.Cart, error) {
	return &models.Cart{}, nil
}
func (f *fakeCartRepo) ClearCart(_ context.Context, _ uint) error     { return nil }
func (f *fakeCartRepo) RemoveItem(_ context.Context, _, _ uint) error { return nil }

// ACTION: Renamed to avoid collision with menu_service_test.go fakeMenuRepo
type fakeCartMenuRepo struct {
	restaurant  models.Restaurant
	item        models.MenuItem
	findItemErr error
}

func (f *fakeCartMenuRepo) FindByID(_ context.Context, _ uint) (*models.MenuItem, error) {
	if f.findItemErr != nil {
		return nil, f.findItemErr
	}
	m := f.item
	m.Restaurant = f.restaurant
	return &m, nil
}

func TestAddCartItem(t *testing.T) {
	active := models.Restaurant{Status: models.RestaurantStatusActive}
	item := models.MenuItem{RestaurantID: 3, Price: 1250, Available: true}
	item.ID = 9

	t.Run("adds a new line at the menu price", func(t *testing.T) {
		carts := &fakeCartRepo{}
		added, err := NewCartService(carts, &fakeCartMenuRepo{restaurant: active, item: item}).AddItem(context.Background(), 4, 9, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if added.Quantity != 2 || added.Price != 1250 {
			t.Fatalf("item = %+v", added)
		}
	})
	t.Run("second add increases quantity and keeps the price", func(t *testing.T) {
		carts := &fakeCartRepo{}
		menus := &fakeCartMenuRepo{restaurant: active, item: item}
		svc := NewCartService(carts, menus)
		if _, err := svc.AddItem(context.Background(), 4, 9, 2); err != nil {
			t.Fatalf("%v", err)
		}
		menus.item.Price = 2000
		added, err := svc.AddItem(context.Background(), 4, 9, 3)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if added.Quantity != 5 || added.Price != 1250 {
			t.Fatalf("item = %+v", added)
		}
	})
	t.Run("quantity below 1", func(t *testing.T) {
		_, err := NewCartService(&fakeCartRepo{}, &fakeCartMenuRepo{restaurant: active, item: item}).AddItem(context.Background(), 4, 9, 0)
		if !errors.Is(err, ErrCartInvalidQuantity) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("missing item", func(t *testing.T) {
		_, err := NewCartService(&fakeCartRepo{}, &fakeCartMenuRepo{findItemErr: gorm.ErrRecordNotFound}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrCartMenuItemNotFound) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("unavailable item", func(t *testing.T) {
		hidden := item
		hidden.Available = false
		_, err := NewCartService(&fakeCartRepo{}, &fakeCartMenuRepo{restaurant: active, item: hidden}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrCartItemUnavailable) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("restaurant is not active", func(t *testing.T) {
		pending := active
		pending.Status = models.RestaurantStatusPending
		_, err := NewCartService(&fakeCartRepo{}, &fakeCartMenuRepo{restaurant: pending, item: item}).AddItem(context.Background(), 4, 9, 1)
		if !errors.Is(err, ErrCartItemUnavailable) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("cart already holds another restaurant", func(t *testing.T) {
		carts := &fakeCartRepo{}
		menus := &fakeCartMenuRepo{restaurant: active, item: item}
		svc := NewCartService(carts, menus)
		svc.AddItem(context.Background(), 4, 9, 1)
		other := item
		other.ID = 10
		other.RestaurantID = 8
		menus.item = other
		_, err := svc.AddItem(context.Background(), 4, 10, 1)
		if !errors.Is(err, ErrCartDifferentRestaurant) {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("line quantity cannot exceed 99", func(t *testing.T) {
		carts := &fakeCartRepo{}
		svc := NewCartService(carts, &fakeCartMenuRepo{restaurant: active, item: item})
		svc.AddItem(context.Background(), 4, 9, 90)
		_, err := svc.AddItem(context.Background(), 4, 9, 10)
		if !errors.Is(err, ErrCartQuantityLimit) {
			t.Fatalf("got %v", err)
		}
	})
}
