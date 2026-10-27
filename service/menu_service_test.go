package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// Use a fake repository so these tests do not need PostgreSQL.
type fakeMenuRepo struct {
	restaurant models.Restaurant
	findRest   error
	findCat    error
	created    *models.MenuItem
}

func (f *fakeMenuRepo) FindRestaurantByID(context.Context, uint) (models.Restaurant, error) {
	return f.restaurant, f.findRest
}

func (f *fakeMenuRepo) FindCategoryByID(context.Context, uint) (models.Category, error) {
	return models.Category{Name: "Mains"}, f.findCat
}

func (f *fakeMenuRepo) CreateMenuItem(_ context.Context, item *models.MenuItem) error {
	item.ID = 9
	f.created = item
	return nil
}

// Support menu reading in the fake repository.
func (f *fakeMenuRepo) FindMenuByRestaurant(context.Context, uint, uint) ([]models.MenuItem, error) {
	return []models.MenuItem{}, nil
}

// Start each creation test with valid details.
func validMenuRequest() CreateMenuItemRequest {
	return CreateMenuItemRequest{
		RestaurantID: 3,
		CategoryID:   2,
		Name:         "  Margherita  ",
		Description:  " tomato ",
		Price:        1250,
	}
}

// Keep the team's existing menu creation tests.
func TestCreateMenuItem(t *testing.T) {
	active := models.Restaurant{
		ManagerID: 4,
		Status:    models.RestaurantStatusActive,
	}

	t.Run("creates an available item", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active}
		item, err := NewMenuService(repo).CreateMenuItem(context.Background(), 4, validMenuRequest())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Name != "Margherita" || item.Price != 1250 || !item.Available {
			t.Fatalf("unexpected item: %+v", item)
		}
		if repo.created == nil || repo.created.RestaurantID != 3 {
			t.Fatal("item was not saved")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		req := validMenuRequest()
		req.Name = "  "
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active}).CreateMenuItem(context.Background(), 4, req)
		if !errors.Is(err, ErrNameRequired) {
			t.Fatalf("expected ErrNameRequired, got %v", err)
		}
	})

	t.Run("price must be positive", func(t *testing.T) {
		req := validMenuRequest()
		req.Price = 0
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active}).CreateMenuItem(context.Background(), 4, req)
		if !errors.Is(err, ErrInvalidPrice) {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("missing restaurant id", func(t *testing.T) {
		req := validMenuRequest()
		req.RestaurantID = 0
		_, err := NewMenuService(&fakeMenuRepo{}).CreateMenuItem(context.Background(), 4, req)
		if !errors.Is(err, ErrRestaurantIDRequired) {
			t.Fatalf("expected ErrRestaurantIDRequired, got %v", err)
		}
	})

	t.Run("another manager", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active}).CreateMenuItem(context.Background(), 8, validMenuRequest())
		if !errors.Is(err, ErrNotYourRestaurant) {
			t.Fatalf("expected ErrNotYourRestaurant, got %v", err)
		}
	})

	t.Run("pending restaurant", func(t *testing.T) {
		pending := active
		pending.Status = models.RestaurantStatusPending
		_, err := NewMenuService(&fakeMenuRepo{restaurant: pending}).CreateMenuItem(context.Background(), 4, validMenuRequest())
		if !errors.Is(err, ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
	})

	t.Run("missing category", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{
			restaurant: active,
			findCat:    gorm.ErrRecordNotFound,
		}).CreateMenuItem(context.Background(), 4, validMenuRequest())
		if !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got %v", err)
		}
	})
}
