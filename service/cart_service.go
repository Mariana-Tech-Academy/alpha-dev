package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidQuantity         = errors.New("quantity must be at least 1")
	ErrCartMenuItemNotFound    = errors.New("menu item not found")
	ErrItemUnavailable         = errors.New("menu item is unavailable")
	ErrInvalidMenuItem         = errors.New("invalid menu item ID")
	ErrInvalidUser             = errors.New("invalid user ID")
	ErrCartDifferentRestaurant = errors.New("cart contains items from another restaurant")
	ErrCartQuantityLimit       = errors.New("quantity cannot exceed 99")
)

type CartService interface {
	AddItem(ctx context.Context, userID uint, menuItemID uint, quantity int) (*models.CartItem, error)
}

type cartService struct {
	cartRepo repository.CartRepository
	menuRepo repository.MenuRepository
}

func NewCartService(cartRepo repository.CartRepository, menuRepo repository.MenuRepository) CartService {
	return &cartService{cartRepo: cartRepo, menuRepo: menuRepo}
}

func (s *cartService) AddItem(ctx context.Context, userID uint, menuItemID uint, quantity int) (*models.CartItem, error) {
	if userID == 0 {
		return nil, ErrInvalidUser
	}
	if menuItemID == 0 {
		return nil, ErrInvalidMenuItem
	}
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}
	if quantity > repository.MaxCartLineQuantity {
		return nil, ErrCartQuantityLimit
	}

	menuItem, err := s.menuRepo.FindMenuItemByID(ctx, menuItemID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartMenuItemNotFound
	}
	if err != nil {
		return nil, err
	}
	if !menuItem.Available {
		return nil, ErrItemUnavailable
	}

	restaurant, err := s.menuRepo.FindRestaurantByID(ctx, menuItem.RestaurantID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrItemUnavailable
	}
	if err != nil {
		return nil, err
	}
	if restaurant.Status != models.RestaurantStatusActive {
		return nil, ErrItemUnavailable
	}

	item, err := s.cartRepo.AddItem(ctx, userID, menuItemID, menuItem.RestaurantID, quantity, menuItem.Price)
	if errors.Is(err, repository.ErrDifferentRestaurant) {
		return nil, ErrCartDifferentRestaurant
	}
	if errors.Is(err, repository.ErrCartLineFull) {
		return nil, ErrCartQuantityLimit
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}
