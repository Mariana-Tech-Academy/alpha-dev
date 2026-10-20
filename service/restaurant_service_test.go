package service

import (
	"context"
	"errors"
	"testing"

	"alpha-dev/models"
)

type fakeRestaurantRepo struct {
	created *models.Restaurant
	err     error
}

func (f *fakeRestaurantRepo) Create(_ context.Context, restaurant *models.Restaurant) error {
	f.created = restaurant
	restaurant.ID = 7
	return f.err
}

func (f *fakeRestaurantRepo) FindByID(context.Context, uint) (models.Restaurant, error) {
	return models.Restaurant{}, nil
}

func (f *fakeRestaurantRepo) UpdateAvailability(context.Context, uint, uint, models.RestaurantStatus, models.RestaurantStatus) (bool, error) {
	return false, nil
}

func (f *fakeRestaurantRepo) Update(context.Context, *models.Restaurant) error {
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
