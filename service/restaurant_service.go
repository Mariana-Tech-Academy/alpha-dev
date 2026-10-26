package service

import (
	"alpha-dev/repository"
	"alpha-dev/models"
	"context"	
	"errors"
	"gorm.io/gorm"
)

type RestaurantService interface {
	GetRestaurantDetails(ctx context.Context, id uint) (models.Restaurant, error)
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
