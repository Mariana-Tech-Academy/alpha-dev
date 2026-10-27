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
