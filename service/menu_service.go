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
	ErrMenuUnavailable      = errors.New("restaurant menu is not available")
)

// Details used when creating a menu item.
type CreateMenuItemRequest struct {
	RestaurantID uint   `json:"restaurant_id"`
	CategoryID   uint   `json:"category_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int64  `json:"price"`
	ImageURL     string `json:"image_url"`
}

// Details shown to the customer.
type MenuItemResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Price      int64  `json:"price"`
	ImageURL   string `json:"image_url"`
	CategoryID uint   `json:"category_id"`
	Category   string `json:"category"`
	Available  bool   `json:"available"`
}

type MenuService interface {
	CreateMenuItem(ctx context.Context, userID uint, req CreateMenuItemRequest) (*models.MenuItem, error)
	GetRestaurantMenu(ctx context.Context, restaurantID uint, categoryID uint) ([]MenuItemResponse, error)
}

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(repo repository.MenuRepository) MenuService {
	return &menuService{menuRepo: repo}
}

// Keep the existing menu creation feature.
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

	// Check the restaurant and its owner.
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
	if restaurant.Status != models.RestaurantStatusActive &&
		restaurant.Status != models.RestaurantStatusInactive {
		return nil, ErrRestaurantClosed
	}

	// Check that the category exists.
	if _, err := s.menuRepo.FindCategoryByID(ctx, req.CategoryID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	// Prepare and save the item.
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

// Get a restaurant's menu for the customer.
func (s *menuService) GetRestaurantMenu(ctx context.Context, restaurantID uint, categoryID uint) ([]MenuItemResponse, error) {
	// Check that the restaurant exists and is active.
	if restaurantID == 0 {
		return nil, ErrRestaurantIDRequired
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, restaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRestaurantNotFound
	}
	if err != nil {
		return nil, err
	}
	if restaurant.Status != models.RestaurantStatusActive {
		return nil, ErrMenuUnavailable
	}

	// Read the menu using the restaurant and optional category.
	items, err := s.menuRepo.FindMenuByRestaurant(ctx, restaurantID, categoryID)
	if err != nil {
		return nil, err
	}

	// Prepare each item for the customer.
	menu := make([]MenuItemResponse, 0)
	for _, item := range items {
		menu = append(menu, MenuItemResponse{
			ID:         item.ID,
			Name:       item.Name,
			Price:      item.Price,
			ImageURL:   item.ImageURL,
			CategoryID: item.CategoryID,
			Category:   item.Category.Name,
			Available:  item.Available,
		})
	}

	return menu, nil
}
