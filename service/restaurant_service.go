package service

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrMissingManager          = errors.New("authenticated manager ID is missing")
	ErrRestaurantNotFound      = errors.New("restaurant not found")
	ErrRestaurantForbidden     = errors.New("you cannot manage this restaurant")
	ErrRestaurantInvalidStatus = errors.New("status must be active or inactive")
	ErrRestaurantRestricted    = errors.New("restaurant requires admin action")
	ErrRestaurantChanged       = errors.New("restaurant changed; refresh and try again")
)

// RestaurantValidationError holds field-level validation errors.
type RestaurantValidationError struct {
	Fields map[string]string
}

func (e *RestaurantValidationError) Error() string {
	return "restaurant validation failed"
}

type RestaurantService interface {
	RegisterRestaurant(ctx context.Context, restaurant *models.Restaurant) error
	SetAvailability(
		ctx context.Context,
		userID uint,
		role models.Role,
		restaurantID uint,
		nextStatus models.RestaurantStatus,
	) error
}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

func validRestaurantPhone(phone string) bool {
	digits := 0
	for _, c := range phone {
		switch {
		case c >= '0' && c <= '9':
			digits++
		case c == ' ' || c == '+' || c == '-' || c == '(' || c == ')':
		default:
			return false
		}
	}
	return digits >= 7
}

// RegisterRestaurant validates and saves a new restaurant.
// Opening and closing times must be HH:MM on the same day; closing must be later than opening.
func (s *restaurantService) RegisterRestaurant(ctx context.Context, restaurant *models.Restaurant) error {
	if restaurant.ManagerID == 0 {
		return ErrMissingManager
	}

	restaurant.Name = strings.TrimSpace(restaurant.Name)
	restaurant.Description = strings.TrimSpace(restaurant.Description)
	restaurant.Address = strings.TrimSpace(restaurant.Address)
	restaurant.City = strings.TrimSpace(restaurant.City)
	restaurant.State = strings.TrimSpace(restaurant.State)
	restaurant.Phone = strings.TrimSpace(restaurant.Phone)
	restaurant.Email = strings.ToLower(strings.TrimSpace(restaurant.Email))
	restaurant.OpeningTime = strings.TrimSpace(restaurant.OpeningTime)
	restaurant.ClosingTime = strings.TrimSpace(restaurant.ClosingTime)

	fields := make(map[string]string)

	if restaurant.Name == "" {
		fields["name"] = "restaurant name is required"
	}
	if restaurant.Description == "" {
		fields["description"] = "description is required"
	}
	if restaurant.Address == "" {
		fields["address"] = "address is required"
	}
	if restaurant.Phone == "" {
		fields["phone"] = "phone number is required"
	} else if !validRestaurantPhone(restaurant.Phone) {
		fields["phone"] = "enter a valid phone number"
	}

	if restaurant.Email == "" {
		fields["email"] = "email is required"
	} else {
		parsed, err := mail.ParseAddress(restaurant.Email)
		if err != nil || parsed.Address != restaurant.Email {
			fields["email"] = "invalid email address"
		}
	}

	if restaurant.OpeningTime == "" {
		fields["opening_time"] = "opening time is required"
	}
	if restaurant.ClosingTime == "" {
		fields["closing_time"] = "closing time is required"
	}

	opening, openingErr := time.Parse("15:04", restaurant.OpeningTime)
	closing, closingErr := time.Parse("15:04", restaurant.ClosingTime)

	if restaurant.OpeningTime != "" && openingErr != nil {
		fields["opening_time"] = "opening time must be HH:MM"
	}
	if restaurant.ClosingTime != "" && closingErr != nil {
		fields["closing_time"] = "closing time must be HH:MM"
	}
	if openingErr == nil && closingErr == nil && !closing.After(opening) {
		fields["closing_time"] = "closing time must be after opening time"
	}

	if len(fields) > 0 {
		return &RestaurantValidationError{Fields: fields}
	}

	restaurant.Status = models.RestaurantStatusPending
	return s.restaurantRepo.Create(ctx, restaurant)
}

func (s *restaurantService) SetAvailability(
	ctx context.Context,
	userID uint,
	role models.Role,
	restaurantID uint,
	nextStatus models.RestaurantStatus,
) error {
	if userID == 0 || role != models.RoleManager {
		return ErrRestaurantForbidden
	}

	if nextStatus != models.RestaurantStatusActive && nextStatus != models.RestaurantStatusInactive {
		return ErrRestaurantInvalidStatus
	}

	restaurant, err := s.restaurantRepo.FindByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRestaurantNotFound
	}
	if err != nil {
		return err
	}

	if restaurant.ManagerID != userID {
		return ErrRestaurantForbidden
	}

	if restaurant.Status != models.RestaurantStatusActive && restaurant.Status != models.RestaurantStatusInactive {
		return ErrRestaurantRestricted
	}

	updated, err := s.restaurantRepo.UpdateAvailability(ctx, restaurantID, userID, restaurant.Status, nextStatus)
	if err != nil {
		return err
	}
	if !updated {
		return ErrRestaurantChanged
	}

	return nil
}
