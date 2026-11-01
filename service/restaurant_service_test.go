package service

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type fakeRestaurantRepo struct {
	created *models.Restaurant
	err     error
	found   models.Restaurant
	findErr error
	updates map[string]any
	list    []models.Restaurant
	listErr error
	page    int
	limit   int
	pattern string
	offset  int
}

func (f *fakeRestaurantRepo) Create(_ context.Context, restaurant *models.Restaurant) error {
	f.created = restaurant
	restaurant.ID = 7
	return f.err
}

func (f *fakeRestaurantRepo) FindByID(context.Context, uint) (models.Restaurant, error) {
	return f.found, f.findErr
}

func (f *fakeRestaurantRepo) FindActiveByID(_ context.Context, id uint) (models.Restaurant, error) {
	if f.findErr != nil {
		return models.Restaurant{}, f.findErr
	}
	if f.found.ID != id || f.found.Status != models.RestaurantStatusActive {
		return models.Restaurant{}, gorm.ErrRecordNotFound
	}
	return f.found, nil
}

func (f *fakeRestaurantRepo) UpdateAvailability(context.Context, uint, uint, models.RestaurantStatus, models.RestaurantStatus) (bool, error) {
	return false, nil
}

func (f *fakeRestaurantRepo) UpdateFields(_ context.Context, _ uint, fields map[string]any) error {
	f.updates = fields
	return nil
}

func (f *fakeRestaurantRepo) SearchActiveByNameOrLocation(_ context.Context, pattern string, limit, offset int) ([]models.Restaurant, error) {
	f.pattern = pattern
	f.limit = limit
	f.offset = offset
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.list == nil {
		return []models.Restaurant{}, nil
	}
	return f.list, nil
}

func (f *fakeRestaurantRepo) GetActiveRestaurants(_ context.Context, page, limit int) ([]models.Restaurant, error) {
	f.page = page
	f.limit = limit
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.list, nil
}

func validRestaurant() *models.Restaurant {
	return &models.Restaurant{
		ManagerID:   4,
		Name:        "  Pizza Place  ",
		Description: " Wood fired ",
		Address:     " 1 High Street ",
		City:        " London ",
		State:       " Greater London ",
		Phone:       " +44 7700 900123 ",
		Email:       "  Kitchen@Example.com ",
		OpeningTime: "09:00",
		ClosingTime: "22:00",
	}
}

func TestRegisterRestaurant(t *testing.T) {
	t.Run("trims fields and sets pending status", func(t *testing.T) {
		repo := &fakeRestaurantRepo{}
		restaurant := validRestaurant()

		err := NewRestaurantService(repo).RegisterRestaurant(context.Background(), restaurant)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if restaurant.Name != "Pizza Place" || restaurant.Email != "kitchen@example.com" {
			t.Fatalf("fields were not normalised: %+v", restaurant)
		}
		if restaurant.City != "London" || restaurant.Phone != "+44 7700 900123" {
			t.Fatalf("city or phone was not trimmed: %+v", restaurant)
		}
		if restaurant.Status != models.RestaurantStatusPending {
			t.Fatalf("status = %s", restaurant.Status)
		}
		if repo.created == nil || repo.created.ManagerID != 4 || repo.created.ID != 7 {
			t.Fatalf("restaurant was not saved: %+v", repo.created)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		restaurant := validRestaurant()
		restaurant.Name = "   "
		err := NewRestaurantService(&fakeRestaurantRepo{}).RegisterRestaurant(context.Background(), restaurant)
		var validationErr *RestaurantValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("expected validation error, got %v", err)
		}
		if _, ok := validationErr.Fields["name"]; !ok {
			t.Fatalf("expected name error, got %v", validationErr.Fields)
		}
	})

	t.Run("invalid phone", func(t *testing.T) {
		restaurant := validRestaurant()
		restaurant.Phone = "hello"
		err := NewRestaurantService(&fakeRestaurantRepo{}).RegisterRestaurant(context.Background(), restaurant)
		var validationErr *RestaurantValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("expected validation error, got %v", err)
		}
		if _, ok := validationErr.Fields["phone"]; !ok {
			t.Fatalf("expected phone error, got %v", validationErr.Fields)
		}
	})

	t.Run("closing before opening", func(t *testing.T) {
		restaurant := validRestaurant()
		restaurant.OpeningTime = "18:00"
		restaurant.ClosingTime = "09:00"
		err := NewRestaurantService(&fakeRestaurantRepo{}).RegisterRestaurant(context.Background(), restaurant)
		var validationErr *RestaurantValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("expected validation error, got %v", err)
		}
		if _, ok := validationErr.Fields["closing_time"]; !ok {
			t.Fatalf("expected closing_time error, got %v", validationErr.Fields)
		}
	})

	t.Run("missing manager", func(t *testing.T) {
		restaurant := validRestaurant()
		restaurant.ManagerID = 0
		err := NewRestaurantService(&fakeRestaurantRepo{}).RegisterRestaurant(context.Background(), restaurant)
		if !errors.Is(err, ErrMissingManager) {
			t.Fatalf("expected ErrMissingManager, got %v", err)
		}
	})
}

func strPtr(value string) *string {
	return &value
}

func ownedRestaurant() models.Restaurant {
	restaurant := models.Restaurant{
		ManagerID:   4,
		Name:        "Pizza Place",
		Address:     "1 High Street",
		Phone:       "+44 7700 900123",
		Email:       "kitchen@example.com",
		OpeningTime: "09:00",
		ClosingTime: "22:00",
		Status:      models.RestaurantStatusActive,
	}
	restaurant.ID = 3
	return restaurant
}

func TestUpdateRestaurant(t *testing.T) {
	t.Run("owner updates the name", func(t *testing.T) {
		repo := &fakeRestaurantRepo{found: ownedRestaurant()}
		updated, err := NewRestaurantService(repo).UpdateRestaurant(context.Background(), 3, 4, models.RoleManager, &models.UpdateRestaurantInput{
			Name: strPtr("  New Name  "),
		}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Fatalf("name = %q", updated.Name)
		}
		if repo.updates["name"] != "New Name" {
			t.Fatalf("saved fields = %v", repo.updates)
		}
		if _, ok := repo.updates["address"]; ok {
			t.Fatal("unchanged fields were saved")
		}
	})

	t.Run("another manager is forbidden", func(t *testing.T) {
		_, err := NewRestaurantService(&fakeRestaurantRepo{found: ownedRestaurant()}).UpdateRestaurant(context.Background(), 3, 8, models.RoleManager, &models.UpdateRestaurantInput{
			Name: strPtr("Takeover"),
		}, nil)
		if !errors.Is(err, ErrRestaurantForbidden) {
			t.Fatalf("expected ErrRestaurantForbidden, got %v", err)
		}
	})

	t.Run("admin can update", func(t *testing.T) {
		repo := &fakeRestaurantRepo{found: ownedRestaurant()}
		_, err := NewRestaurantService(repo).UpdateRestaurant(context.Background(), 3, 1, models.RoleAdmin, &models.UpdateRestaurantInput{
			Name: strPtr("Admin Edit"),
		}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.updates["name"] != "Admin Edit" {
			t.Fatalf("saved fields = %v", repo.updates)
		}
	})

	t.Run("empty name is rejected", func(t *testing.T) {
		_, err := NewRestaurantService(&fakeRestaurantRepo{found: ownedRestaurant()}).UpdateRestaurant(context.Background(), 3, 4, models.RoleManager, &models.UpdateRestaurantInput{
			Name: strPtr("   "),
		}, nil)
		var validationErr *RestaurantValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("expected validation error, got %v", err)
		}
		if _, ok := validationErr.Fields["name"]; !ok {
			t.Fatalf("fields = %v", validationErr.Fields)
		}
	})

	t.Run("gif url is rejected", func(t *testing.T) {
		_, err := NewRestaurantService(&fakeRestaurantRepo{found: ownedRestaurant()}).UpdateRestaurant(context.Background(), 3, 4, models.RoleManager, &models.UpdateRestaurantInput{
			ImageURL: strPtr("logo.gif"),
		}, nil)
		var validationErr *RestaurantValidationError
		if !errors.As(err, &validationErr) {
			t.Fatalf("expected validation error, got %v", err)
		}
		if _, ok := validationErr.Fields["image_url"]; !ok {
			t.Fatalf("fields = %v", validationErr.Fields)
		}
	})

	t.Run("file upload is rejected", func(t *testing.T) {
		_, err := NewRestaurantService(&fakeRestaurantRepo{found: ownedRestaurant()}).UpdateRestaurant(context.Background(), 3, 4, models.RoleManager, &models.UpdateRestaurantInput{}, &multipart.FileHeader{
			Filename: "logo.png",
		})
		if !errors.Is(err, ErrInvalidImage) {
			t.Fatalf("expected ErrInvalidImage, got %v", err)
		}
	})
}

func TestListActiveRestaurants(t *testing.T) {
	first := ownedRestaurant()
	second := ownedRestaurant()
	second.ID = 8
	second.Name = "Noodle Bar"

	t.Run("returns a page of active restaurants", func(t *testing.T) {
		repo := &fakeRestaurantRepo{list: []models.Restaurant{first, second}}
		profiles, err := NewRestaurantService(repo).ListActiveRestaurants(context.Background(), 0, 100)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if repo.page != 1 || repo.limit != 20 {
			t.Fatalf("page = %d limit = %d", repo.page, repo.limit)
		}
		if len(profiles) != 2 || profiles[0].Name != "Pizza Place" || profiles[1].ID != 8 {
			t.Fatalf("profiles = %+v", profiles)
		}
	})

	t.Run("empty page", func(t *testing.T) {
		profiles, err := NewRestaurantService(&fakeRestaurantRepo{}).ListActiveRestaurants(context.Background(), 2, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profiles == nil || len(profiles) != 0 {
			t.Fatalf("profiles = %#v", profiles)
		}
	})
}

func TestGetRestaurantDetails(t *testing.T) {
	available := models.MenuItem{Name: "Margherita", Description: "tomato", Price: 1250, Available: true}
	available.ID = 9
	available.Category.Name = "Pizza"
	unavailable := models.MenuItem{Name: "Sold out", Available: false}
	unavailable.ID = 10

	active := ownedRestaurant()
	active.MenuItems = []models.MenuItem{available, unavailable}

	t.Run("returns an active restaurant and its available menu", func(t *testing.T) {
		details, err := NewRestaurantService(&fakeRestaurantRepo{found: active}).GetRestaurantDetails(context.Background(), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if details.ID != 3 || details.Name != "Pizza Place" || details.Status != models.RestaurantStatusActive {
			t.Fatalf("details = %+v", details)
		}
		if len(details.MenuItems) != 1 || details.MenuItems[0].Name != "Margherita" || details.MenuItems[0].Category != "Pizza" {
			t.Fatalf("menu = %+v", details.MenuItems)
		}
	})

	t.Run("missing restaurant", func(t *testing.T) {
		_, err := NewRestaurantService(&fakeRestaurantRepo{findErr: gorm.ErrRecordNotFound}).GetRestaurantDetails(context.Background(), 3)
		if !errors.Is(err, ErrRestaurantNotFound) {
			t.Fatalf("expected ErrRestaurantNotFound, got %v", err)
		}
	})

	t.Run("inactive restaurant", func(t *testing.T) {
		inactive := active
		inactive.Status = models.RestaurantStatusInactive
		_, err := NewRestaurantService(&fakeRestaurantRepo{found: inactive}).GetRestaurantDetails(context.Background(), 3)
		if !errors.Is(err, ErrRestaurantNotFound) {
			t.Fatalf("expected ErrRestaurantNotFound, got %v", err)
		}
	})
}

func TestSearchRestaurants(t *testing.T) {
	repo := &fakeRestaurantRepo{list: []models.Restaurant{{
		Name: "Lagos Kitchen", City: "Lagos", Status: models.RestaurantStatusActive,
	}}}
	repo.list[0].ID = 4
	profiles, err := NewRestaurantService(repo).SearchRestaurants(context.Background(), "  la%gos_\\  ", 0, 100)
	if err != nil {
		t.Fatalf("search restaurants: %v", err)
	}
	if repo.pattern != `%la\%gos\_\\%` || repo.limit != 20 || repo.offset != 0 {
		t.Fatalf("pattern %q limit %d offset %d", repo.pattern, repo.limit, repo.offset)
	}
	if len(profiles) != 1 || profiles[0].Name != "Lagos Kitchen" || profiles[0].ID != 4 {
		t.Fatalf("profiles = %+v", profiles)
	}

	empty, err := NewRestaurantService(&fakeRestaurantRepo{}).SearchRestaurants(context.Background(), "   ", 1, 10)
	if err != nil {
		t.Fatalf("empty query: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("empty query returned %+v", empty)
	}
}
