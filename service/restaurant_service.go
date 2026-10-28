package service

import (
	"alpha-dev/models"
	"alpha-dev/repository"
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"
)

type RestaurantService interface {
	GetRestaurantDetails(ctx context.Context, id uint) (models.Restaurant, error)
	SearchRestaurants(ctx context.Context, query string) ([]models.Restaurant, error)
}

type restaurantService struct {
	restaurantRepo repository.RestaurantRepository
}

func NewRestaurantService(repo repository.RestaurantRepository) RestaurantService {
	return &restaurantService{restaurantRepo: repo}
}

// GetRestaurantDetails retrieves the details of an active restaurant by its ID, including its menu items.
func (s *restaurantService) GetRestaurantDetails(ctx context.Context, id uint) (models.Restaurant, error) {
    restaurant, err := s.restaurantRepo.FindActiveByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return models.Restaurant{}, errors.New("restaurant not found or not active")
        }
        return models.Restaurant{}, err
    }
    return restaurant, nil
}

func (s *restaurantService) SearchRestaurants(ctx context.Context, query string) ([]models.Restaurant, error) {
	//Trim whitespace from the user's input
	trimmedQuery := strings.TrimSpace(query)

	// Return an empty slice immediately if query is empty, instead of executing an unnecessary database query.
	if trimmedQuery == "" {
		return []models.Restaurant{}, nil
	}

	// Prepare the search pattern for the repository
	pattern := "%" + trimmedQuery + "%"

	return s.restaurantRepo.SearchActiveByNameOrLocation(ctx, pattern)
}