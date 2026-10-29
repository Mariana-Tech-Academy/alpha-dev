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
	ErrInvalidPriceRange    = errors.New("min_price cannot be greater than max_price")
	ErrInvalidPriceBound    = errors.New("min_price and max_price must be zero or greater")
	ErrInvalidCategoryID    = errors.New("category_id must be a positive number")
	ErrInvalidRestaurantID  = errors.New("restaurant_id must be a positive number")
	ErrMenuUnavailable      = errors.New("restaurant menu is not available")
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

// MenuItemResponse is one row on a restaurant's menu.
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
	UpdateMenuItemAvailability(ctx context.Context, userID uint, itemID uint, available bool) (*models.MenuItem, error)
	GetMenuItemDetails(ctx context.Context, id uint) (models.PublicMenuItem, error)
	UpdateMenuItem(ctx context.Context, userID uint, id uint, req UpdateMenuItemRequest) (*models.MenuItem, error)
	DeleteMenuItem(ctx context.Context, userID uint, id uint) error
	SearchMenuItems(ctx context.Context, search MenuSearch) ([]models.MenuSearchResult, error)
}

// MenuSearch is the query for GET /menu.
type MenuSearch struct {
	Query        string
	CategoryID   *uint
	RestaurantID *uint
	MinPrice     *int64
	MaxPrice     *int64
	Available    *bool
	Page         int
	Limit        int
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

func (s *menuService) UpdateMenuItemAvailability(ctx context.Context, userID uint, itemID uint, available bool) (*models.MenuItem, error) {
	item, err := s.menuRepo.FindMenuItemByID(ctx, itemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, item.RestaurantID)
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

	if err := s.menuRepo.UpdateMenuItemAvailability(ctx, itemID, available); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMenuItemNotFound
		}
		return nil, err
	}

	item.Available = available
	return &item, nil
}

func (s *menuService) GetMenuItemDetails(ctx context.Context, id uint) (models.PublicMenuItem, error) {
	item, err := s.menuRepo.FindMenuItemByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.PublicMenuItem{}, ErrMenuItemNotFound
	}
	if err != nil {
		return models.PublicMenuItem{}, err
	}
	if !item.Available || item.Restaurant.Status != models.RestaurantStatusActive {
		return models.PublicMenuItem{}, ErrMenuItemNotFound
	}

	return models.PublicMenuItem{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		ImageURL:    item.ImageURL,
		Category:    item.Category.Name,
	}, nil
}

const maxMenuPageSize = 20

func (s *menuService) SearchMenuItems(ctx context.Context, search MenuSearch) ([]models.MenuSearchResult, error) {
	search.Query = strings.TrimSpace(search.Query)
	if search.CategoryID != nil && *search.CategoryID == 0 {
		return nil, ErrInvalidCategoryID
	}
	if search.RestaurantID != nil && *search.RestaurantID == 0 {
		return nil, ErrInvalidRestaurantID
	}
	if (search.MinPrice != nil && *search.MinPrice < 0) || (search.MaxPrice != nil && *search.MaxPrice < 0) {
		return nil, ErrInvalidPriceBound
	}
	if search.MinPrice != nil && search.MaxPrice != nil && *search.MinPrice > *search.MaxPrice {
		return nil, ErrInvalidPriceRange
	}
	if search.Page < 1 {
		search.Page = 1
	}
	if search.Limit < 1 {
		search.Limit = 10
	}
	if search.Limit > maxMenuPageSize {
		search.Limit = maxMenuPageSize
	}
	available := true
	if search.Available != nil {
		available = *search.Available
	}

	items, err := s.menuRepo.SearchMenuItems(ctx, repository.MenuSearchQuery{
		Query:        escapeLike(search.Query),
		CategoryID:   search.CategoryID,
		RestaurantID: search.RestaurantID,
		MinPrice:     search.MinPrice,
		MaxPrice:     search.MaxPrice,
		Available:    available,
		Limit:        search.Limit,
		Offset:       (search.Page - 1) * search.Limit,
	})
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []models.MenuSearchResult{}
	}
	return items, nil
}

func (s *menuService) GetRestaurantMenu(ctx context.Context, restaurantID uint, categoryID uint) ([]MenuItemResponse, error) {
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

	items, err := s.menuRepo.FindMenuByRestaurant(ctx, restaurantID, categoryID)
	if err != nil {
		return nil, err
	}

	menu := make([]MenuItemResponse, 0, len(items))
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

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
