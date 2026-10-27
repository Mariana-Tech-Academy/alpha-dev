package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

type fakeMenuRepo struct {
	restaurant models.Restaurant
	findRest   error
	findCat    error
	created    *models.MenuItem

	item             models.MenuItem
	findItemErr      error
	updateErr        error
	updatedID        uint
	updatedAvailable bool
	didUpdate        bool
	saved            map[string]any
	savedID          uint
	deletedID        uint
	didDelete        bool
	searchQuery      repository.MenuSearchQuery
	searchErr        error
	searchResults    []models.MenuSearchResult
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

func (f *fakeMenuRepo) FindMenuItemByID(context.Context, uint) (models.MenuItem, error) {
	return f.item, f.findItemErr
}

func (f *fakeMenuRepo) UpdateMenuItemAvailability(_ context.Context, id uint, available bool) error {
	f.updatedID = id
	f.updatedAvailable = available
	f.didUpdate = true
	if f.updateErr != nil {
		return f.updateErr
	}
	f.item.Available = available
	return nil
}

func (f *fakeMenuRepo) UpdateMenuItem(_ context.Context, id uint, fields map[string]any) error {
	f.savedID = id
	f.saved = fields
	return nil
}

func (f *fakeMenuRepo) DeleteMenuItem(_ context.Context, id uint) error {
	f.deletedID = id
	f.didDelete = true
	return nil
}

func (f *fakeMenuRepo) SearchMenuItems(_ context.Context, q repository.MenuSearchQuery) ([]models.MenuSearchResult, error) {
	f.searchQuery = q
	if f.searchErr != nil {
		return nil, f.searchErr
	}
	if f.searchResults == nil {
		return []models.MenuSearchResult{}, nil
	}
	return f.searchResults, nil
}

func validMenuRequest() CreateMenuItemRequest {
	return CreateMenuItemRequest{
		RestaurantID: 3,
		CategoryID:   2,
		Name:         "  Margherita  ",
		Description:  " tomato ",
		Price:        1250,
	}
}

func TestCreateMenuItem(t *testing.T) {
	active := models.Restaurant{ManagerID: 4, Status: models.RestaurantStatusActive}

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
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active, findCat: gorm.ErrRecordNotFound}).CreateMenuItem(context.Background(), 4, validMenuRequest())
		if !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got %v", err)
		}
	})
}

func TestUpdateMenuItemAvailability(t *testing.T) {
	active := models.Restaurant{ManagerID: 4, Status: models.RestaurantStatusActive}
	item := models.MenuItem{RestaurantID: 3, Name: "Margherita", Available: true}
	item.ID = 5

	t.Run("owner marks item unavailable", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		updated, err := NewMenuService(repo).UpdateMenuItemAvailability(context.Background(), 4, 5, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Available || !repo.didUpdate || repo.updatedID != 5 || repo.updatedAvailable {
			t.Fatalf("item available=%v updated id=%d available=%v", updated.Available, repo.updatedID, repo.updatedAvailable)
		}
	})

	t.Run("inactive restaurant", func(t *testing.T) {
		inactive := active
		inactive.Status = models.RestaurantStatusInactive
		repo := &fakeMenuRepo{restaurant: inactive, item: item}
		updated, err := NewMenuService(repo).UpdateMenuItemAvailability(context.Background(), 4, 5, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !updated.Available || repo.updatedID != 5 {
			t.Fatalf("updated = %+v", updated)
		}
	})

	t.Run("another manager", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		_, err := NewMenuService(repo).UpdateMenuItemAvailability(context.Background(), 8, 5, false)
		if !errors.Is(err, ErrNotYourRestaurant) {
			t.Fatalf("expected ErrNotYourRestaurant, got %v", err)
		}
		if repo.didUpdate {
			t.Fatal("availability was updated")
		}
	})

	t.Run("missing item", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{findItemErr: gorm.ErrRecordNotFound}).UpdateMenuItemAvailability(context.Background(), 4, 5, false)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})

	t.Run("missing restaurant", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{item: item, findRest: gorm.ErrRecordNotFound}).UpdateMenuItemAvailability(context.Background(), 4, 5, false)
		if !errors.Is(err, ErrRestaurantNotFound) {
			t.Fatalf("expected ErrRestaurantNotFound, got %v", err)
		}
	})

	t.Run("pending restaurant", func(t *testing.T) {
		pending := active
		pending.Status = models.RestaurantStatusPending
		repo := &fakeMenuRepo{restaurant: pending, item: item}
		_, err := NewMenuService(repo).UpdateMenuItemAvailability(context.Background(), 4, 5, false)
		if !errors.Is(err, ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
		if repo.didUpdate {
			t.Fatal("availability was updated")
		}
	})

	t.Run("suspended restaurant", func(t *testing.T) {
		suspended := active
		suspended.Status = models.RestaurantStatusSuspended
		_, err := NewMenuService(&fakeMenuRepo{restaurant: suspended, item: item}).UpdateMenuItemAvailability(context.Background(), 4, 5, true)
		if !errors.Is(err, ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
	})

	t.Run("item removed before update", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active, item: item, updateErr: gorm.ErrRecordNotFound}).UpdateMenuItemAvailability(context.Background(), 4, 5, false)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})
}

func TestGetMenuItemDetails(t *testing.T) {
	item := models.MenuItem{
		Name:        "Margherita",
		Description: "tomato",
		Price:       1250,
		ImageURL:    "pizza.jpg",
		Available:   true,
		Restaurant:  models.Restaurant{Status: models.RestaurantStatusActive},
		Category:    models.Category{Name: "Pizza"},
	}
	item.ID = 9

	t.Run("visible item", func(t *testing.T) {
		got, err := NewMenuService(&fakeMenuRepo{item: item}).GetMenuItemDetails(context.Background(), 9)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != 9 || got.Name != "Margherita" || got.Price != 1250 || got.Category != "Pizza" || got.ImageURL != "pizza.jpg" {
			t.Fatalf("item = %+v", got)
		}
	})

	t.Run("missing item", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{findItemErr: gorm.ErrRecordNotFound}).GetMenuItemDetails(context.Background(), 9)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})

	t.Run("unavailable item", func(t *testing.T) {
		hidden := item
		hidden.Available = false
		_, err := NewMenuService(&fakeMenuRepo{item: hidden}).GetMenuItemDetails(context.Background(), 9)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})

	t.Run("restaurant is not active", func(t *testing.T) {
		hidden := item
		hidden.Restaurant.Status = models.RestaurantStatusPending
		_, err := NewMenuService(&fakeMenuRepo{item: hidden}).GetMenuItemDetails(context.Background(), 9)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})
}

func TestUpdateMenuItem(t *testing.T) {
	active := models.Restaurant{ManagerID: 4, Status: models.RestaurantStatusActive}
	item := models.MenuItem{RestaurantID: 3, Name: "Old", Available: true}
	item.ID = 9
	req := UpdateMenuItemRequest{
		Name:        "  New Name  ",
		Description: "  cheese ",
		Price:       1500,
		ImageURL:    "  pizza.jpg ",
		CategoryID:  2,
	}

	t.Run("owner updates the item", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		updated, err := NewMenuService(repo).UpdateMenuItem(context.Background(), 4, 9, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" || updated.Description != "cheese" || updated.ImageURL != "pizza.jpg" || !updated.Available {
			t.Fatalf("item = %+v", updated)
		}
		if repo.savedID != 9 || repo.saved["name"] != "New Name" || repo.saved["price"] != int64(1500) || repo.saved["category_id"] != uint(2) {
			t.Fatalf("saved = %#v", repo.saved)
		}
		if _, ok := repo.saved["available"]; ok {
			t.Fatal("availability was saved")
		}
	})

	t.Run("another manager", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		_, err := NewMenuService(repo).UpdateMenuItem(context.Background(), 8, 9, req)
		if !errors.Is(err, ErrNotYourRestaurant) {
			t.Fatalf("expected ErrNotYourRestaurant, got %v", err)
		}
		if repo.saved != nil {
			t.Fatal("item was saved")
		}
	})

	t.Run("missing item", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{findItemErr: gorm.ErrRecordNotFound}).UpdateMenuItem(context.Background(), 4, 9, req)
		if !errors.Is(err, ErrMenuItemNotFound) {
			t.Fatalf("expected ErrMenuItemNotFound, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		bad := req
		bad.Name = "  "
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active, item: item}).UpdateMenuItem(context.Background(), 4, 9, bad)
		if !errors.Is(err, ErrNameRequired) {
			t.Fatalf("expected ErrNameRequired, got %v", err)
		}
	})

	t.Run("price must be positive", func(t *testing.T) {
		bad := req
		bad.Price = 0
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active, item: item}).UpdateMenuItem(context.Background(), 4, 9, bad)
		if !errors.Is(err, ErrInvalidPrice) {
			t.Fatalf("expected ErrInvalidPrice, got %v", err)
		}
	})

	t.Run("missing category", func(t *testing.T) {
		_, err := NewMenuService(&fakeMenuRepo{restaurant: active, item: item, findCat: gorm.ErrRecordNotFound}).UpdateMenuItem(context.Background(), 4, 9, req)
		if !errors.Is(err, ErrCategoryNotFound) {
			t.Fatalf("expected ErrCategoryNotFound, got %v", err)
		}
	})

	t.Run("pending restaurant", func(t *testing.T) {
		pending := active
		pending.Status = models.RestaurantStatusPending
		repo := &fakeMenuRepo{restaurant: pending, item: item}
		_, err := NewMenuService(repo).UpdateMenuItem(context.Background(), 4, 9, req)
		if !errors.Is(err, ErrRestaurantClosed) {
			t.Fatalf("expected ErrRestaurantClosed, got %v", err)
		}
		if repo.saved != nil {
			t.Fatal("item was saved")
		}
	})
}

func TestDeleteMenuItem(t *testing.T) {
	active := models.Restaurant{ManagerID: 4, Status: models.RestaurantStatusActive}
	item := models.MenuItem{RestaurantID: 3}
	item.ID = 9

	t.Run("owner deletes the item", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		err := NewMenuService(repo).DeleteMenuItem(context.Background(), 4, 9)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !repo.didDelete || repo.deletedID != 9 {
			t.Fatalf("deleted id = %d", repo.deletedID)
		}
	})

	t.Run("another manager", func(t *testing.T) {
		repo := &fakeMenuRepo{restaurant: active, item: item}
		err := NewMenuService(repo).DeleteMenuItem(context.Background(), 8, 9)
		if !errors.Is(err, ErrNotYourRestaurant) {
			t.Fatalf("expected ErrNotYourRestaurant, got %v", err)
		}
		if repo.didDelete {
			t.Fatal("item was deleted")
		}
	})
}

func TestSearchMenuItems(t *testing.T) {
	t.Run("trims the query and searches available items", func(t *testing.T) {
		repo := &fakeMenuRepo{}
		items, err := NewMenuService(repo).SearchMenuItems(context.Background(), MenuSearch{
			Query: "  pi%zza_\\  ",
			Page:  0,
			Limit: 100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if items == nil {
			t.Fatal("expected an empty slice")
		}
		if repo.searchQuery.Query != `pi\%zza\_\\` || !repo.searchQuery.Available {
			t.Fatalf("query = %#v", repo.searchQuery)
		}
		if repo.searchQuery.Limit != 20 || repo.searchQuery.Offset != 0 {
			t.Fatalf("limit = %d offset = %d", repo.searchQuery.Limit, repo.searchQuery.Offset)
		}
	})

	t.Run("negative price", func(t *testing.T) {
		price := int64(-1)
		_, err := NewMenuService(&fakeMenuRepo{}).SearchMenuItems(context.Background(), MenuSearch{MinPrice: &price})
		if !errors.Is(err, ErrInvalidPriceBound) {
			t.Fatalf("expected ErrInvalidPriceBound, got %v", err)
		}
	})

	t.Run("min price above max price", func(t *testing.T) {
		minPrice := int64(500)
		maxPrice := int64(100)
		_, err := NewMenuService(&fakeMenuRepo{}).SearchMenuItems(context.Background(), MenuSearch{
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		})
		if !errors.Is(err, ErrInvalidPriceRange) {
			t.Fatalf("expected ErrInvalidPriceRange, got %v", err)
		}
	})

	t.Run("category id zero", func(t *testing.T) {
		id := uint(0)
		_, err := NewMenuService(&fakeMenuRepo{}).SearchMenuItems(context.Background(), MenuSearch{CategoryID: &id})
		if !errors.Is(err, ErrInvalidCategoryID) {
			t.Fatalf("expected ErrInvalidCategoryID, got %v", err)
		}
	})
}
