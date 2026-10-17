package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"alpha-dev/models"
	"alpha-dev/repository"
)

// RestaurantValidationError holds field-level validation errors.
type RestaurantValidationError struct {
	Fields map[string]string
}

func (e *RestaurantValidationError) Error() string {
	return "restaurant validation failed"
}

// RestaurantService defines restaurant operations.
type RestaurantService interface {
	RegisterRestaurant(
		ctx context.Context,
		restaurant *models.Restaurant,
	) error
}

// restaurantService contains the restaurant repository.
type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

// NewRestaurantService creates a restaurant service.
func NewRestaurantService(
	repo repository.RestaurantRepository,
) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

// RegisterRestaurant validates and saves a new restaurant.
func (s *restaurantService) RegisterRestaurant(
	ctx context.Context,
	restaurant *models.Restaurant,
) error {

	fields := make(map[string]string)

	// Validate required fields.
	if strings.TrimSpace(restaurant.Name) == "" {
		fields["name"] = "restaurant name is required"
	}

	if strings.TrimSpace(restaurant.Description) == "" {
		fields["description"] = "description is required"
	}

	if strings.TrimSpace(restaurant.Address) == "" {
		fields["address"] = "address is required"
	}

	if strings.TrimSpace(restaurant.Phone) == "" {
		fields["phone"] = "phone number is required"
	}

	// Validate email.
	email := strings.TrimSpace(restaurant.Email)

	if email == "" {
		fields["email"] = "email is required"
	} else {
		parsed, err := mail.ParseAddress(email)

		if err != nil || parsed.Address != email {
			fields["email"] = "invalid email address"
		}
	}

	// Validate opening and closing times.
	if strings.TrimSpace(restaurant.OpeningTime) == "" {
		fields["opening_time"] = "opening time is required"
	}

	if strings.TrimSpace(restaurant.ClosingTime) == "" {
		fields["closing_time"] = "closing time is required"
	}

	opening, openingErr := time.Parse(
		"15:04",
		restaurant.OpeningTime,
	)

	closing, closingErr := time.Parse(
		"15:04",
		restaurant.ClosingTime,
	)

	if restaurant.OpeningTime != "" && openingErr != nil {
		fields["opening_time"] = "opening time must be HH:MM"
	}

	if restaurant.ClosingTime != "" && closingErr != nil {
		fields["closing_time"] = "closing time must be HH:MM"
	}

	if openingErr == nil && closingErr == nil &&
		!closing.After(opening) {
		fields["closing_time"] = "closing time must be after opening time"
	}

	// Check the manager account ID.
	if restaurant.ManagerID == 0 {
		return errors.New("authenticated manager ID is missing")
	}

	// Return all validation errors together.
	if len(fields) > 0 {
		return &RestaurantValidationError{Fields: fields}
	}

	// Set the initial restaurant status.
	restaurant.Status = models.RestaurantStatusPending

	// Save the restaurant using the repository.
	return s.restaurantRepo.Create(ctx, restaurant)
}
