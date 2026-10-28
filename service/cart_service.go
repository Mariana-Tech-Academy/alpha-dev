package service

import (
	"context"
	"errors"
	"log"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrCartInvalidUser         = errors.New("invalid user")
	ErrCartInvalidMenuItem     = errors.New("invalid menu item")
	ErrCartInvalidQuantity     = errors.New("quantity must be at least 1")
	ErrCartMenuItemNotFound    = errors.New("menu item not found")
	ErrCartItemUnavailable     = errors.New("menu item is unavailable")
	ErrCartQuantityLimit       = errors.New("quantity exceeds limit")
	ErrCartDifferentRestaurant = errors.New("cart already contains items from a different restaurant. Clear cart first to add items from this restaurant")
)

type menuFinder interface {
	FindByID(ctx context.Context, id uint) (*models.MenuItem, error)
}

type CartService interface {
	AddItem(ctx context.Context, userID, menuItemID uint, quantity int) (*models.CartItem, error)
	GetCart(ctx context.Context, userID uint) (*models.Cart, error)
	ClearCart(ctx context.Context, userID uint) error
}

type cartService struct {
	cartRepo repository.CartRepository
	menuRepo menuFinder
}

func NewCartService(cartRepo repository.CartRepository, menuRepo menuFinder) CartService {
	return &cartService{cartRepo: cartRepo, menuRepo: menuRepo}
}

func (s *cartService) AddItem(ctx context.Context, userID, menuItemID uint, quantity int) (*models.CartItem, error) {
	if userID == 0 {
		return nil, ErrCartInvalidUser
	}
	if menuItemID == 0 {
		return nil, ErrCartInvalidMenuItem
	}
	if quantity < 1 {
		return nil, ErrCartInvalidQuantity
	}
	if quantity > repository.MaxCartLineQuantity {
		return nil, ErrCartQuantityLimit
	}

	menuItem, err := s.menuRepo.FindByID(ctx, menuItemID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartMenuItemNotFound
		}
		log.Printf("find menu item failed: %v", err)
		return nil, err
	}

	if !menuItem.Available {
		return nil, ErrCartItemUnavailable
	}
	if menuItem.Restaurant.Status != "" && menuItem.Restaurant.Status != models.RestaurantStatusActive {
		return nil, ErrCartItemUnavailable
	}

	item, err := s.cartRepo.AddItem(ctx, userID, menuItemID, menuItem.RestaurantID, quantity, menuItem.Price)
	if err != nil {
		if errors.Is(err, repository.ErrDifferentRestaurant) {
			return nil, ErrCartDifferentRestaurant
		}
		if errors.Is(err, repository.ErrCartLineFull) {
			return nil, ErrCartQuantityLimit
		}
		return nil, err
	}
	return item, nil
}

func (s *cartService) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	if userID == 0 {
		return nil, ErrCartInvalidUser
	}
	return s.cartRepo.GetCart(ctx, userID)
}

func (s *cartService) ClearCart(ctx context.Context, userID uint) error {
	if userID == 0 {
		return ErrCartInvalidUser
	}
	return s.cartRepo.ClearCart(ctx, userID)
}
