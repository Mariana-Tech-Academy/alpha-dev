package service

import (
	"context"
	"errors"
	"strings"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrNameRequired         = errors.New("name is required")
	ErrInvalidPrice         = errors.New("price must be greater than zero")
	ErrRestaurantIDRequired = errors.New("restaurant_id is required")
	ErrNotYourRestaurant    = errors.New("you do not manage this restaurant")
	ErrRestaurantClosed     = errors.New("restaurant cannot accept menu items")
	ErrCategoryIDRequired   = errors.New("category_id is required")
	ErrCategoryNotFound     = errors.New("category not found")
	ErrMenuItemNotFound     = errors.New("menu item not found")
)

// CreateMenuItemRequest is the body for POST /menu.
// Price is an integer in minor units, so 1250 means 12.50.
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
	UpdateMenuItemAvailability(
		ctx context.Context,
		userID uint,
		ItemID uint,
		available bool,
	) error
}

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: repo}
}

func (s *menuService) CreateMenuItem(ctx context.Context, userID uint, req CreateMenuItemRequest) (*models.MenuItem, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrNameRequired
	}
	if req.Price <= 0 {
		return nil, ErrInvalidPrice
	}
	if req.RestaurantID == 0 {
		return nil, ErrRestaurantIDRequired
	}
	if req.CategoryID == 0 {
		return nil, ErrCategoryIDRequired
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, req.RestaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}
	if restaurant.ManagerID != userID {
		return nil, ErrNotYourRestaurant
	}
	if restaurant.Status != models.RestaurantStatusActive && restaurant.Status != models.RestaurantStatusInactive {
		return nil, ErrRestaurantClosed
	}

	if _, err := s.menuRepo.FindCategoryByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	item := models.MenuItem{
		RestaurantID: req.RestaurantID,
		CategoryID:   req.CategoryID,
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		Price:        req.Price,
		ImageURL:     strings.TrimSpace(req.ImageURL),
		Available:    true,
	}
	if err := s.menuRepo.CreateMenuItem(ctx, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// UpdateMenuItemAvailability allows a restaurant manager
// to mark their menu item as available or unavailable.
func (s *menuService) UpdateMenuItemAvailability(
	ctx context.Context,
	userID uint,
	itemID uint,
	available bool,
) error {

	// Find the menu item
	item, err := s.menuRepo.FindMenuItemByID(ctx, itemID)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrMenuItemNotFound
	}

	if err != nil {
		return err
	}

	// Find the restaurant that own the menu item.
	restaurant, err := s.menuRepo.FindRestaurantByID(
		ctx,
		item.RestaurantID,
	)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrRestaurantNotFound
	}

	if err != nil {
		return err
	}

	// Check that the logged-in user manages this restaurant.
	if restaurant.ManagerID != userID {
		return ErrNotYourRestaurant
	}

	// Update the availablilty in PostgreSQL
	return s.menuRepo.UpdateMenuItemAvailability(
		ctx,
		itemID,
		available,
	)
}