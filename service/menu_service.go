package service

import (
	"alpha-dev/repository"
	"context"
	"errors"
	"strings"

	"alpha-dev/models"

	"gorm.io/gorm"
)

var (
	ErrNameRequired       = errors.New("name is required")
	ErrInvalidPrice       = errors.New("price must be greater than zero")
	ErrRestaurantNotFound = errors.New("restaurant not found")
	ErrNotYourRestaurant  = errors.New("you do not manage this restaurant")
	ErrInvalidInput      =  errors.New("invalid input")
)
//describes the data client sends when creating menu item
type CreateMenuItemRequest struct {
	RestaurantID uint   `json:"restaurant_id"`
	CategoryID   uint   `json:"category_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	ImageURL     string `json:"image_url"`
}

type MenuService interface {
	CreateMenuItem(ctx context.Context, userID uint, req CreateMenuItemRequest) (*models.MenuItem, error)
}

func (s *menuService) CreateMenuItem(ctx context.Context, userID uint, req CreateMenuItemRequest) (*models.MenuItem, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, ErrNameRequired
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, req.RestaurantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRestaurantNotFound
		}
		return nil, err
	}
	if restaurant.ManagerID != userID {
		return nil, ErrNotYourRestaurant
	}

	item := models.MenuItem{
		RestaurantID: req.RestaurantID,
		CategoryID:   req.CategoryID,
		Name:         strings.TrimSpace(req.Name),
		Description:  req.Description,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		Available:    true,
	}
	if err := s.menuRepo.CreateMenuItem(ctx, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

//holds repository in a slot called menuRepo
type menuService struct {
	menuRepo repository.MenuRepository
}

//builds service and hands it to repository
func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: repo}
}
