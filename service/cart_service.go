package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrInvalidQuantity = errors.New("quantity must be at least 1")

	ErrCartMenuItemNotFound = errors.New("menu item not found")
	ErrItemUnavailable      = errors.New("menu item is unavailable")
	ErrInvalidMenuItem      = errors.New("invalid menu item ID")
	ErrInvalidUser          = errors.New("invalid user ID")
)

type CartService interface {
	AddItem(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		quantity int,
	) (*models.CartItem, error)
}

type cartService struct {
	cartRepo repository.CartRepository
	db       *gorm.DB
}

func NewCartService(
	repo repository.CartRepository,
	db *gorm.DB,
) CartService {
	return &cartService{
		cartRepo: repo,
		db:       db,
	}
}

func (s *cartService) AddItem(
	ctx context.Context,
	userID uint,
	menuItemID uint,
	quantity int,
) (*models.CartItem, error) {

	// Validate the customer ID.
	if userID == 0 {
		return nil, ErrInvalidUser
	}

	// Validate the quantity.
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	// Validate the menu item ID.
	if menuItemID == 0 {
		return nil, ErrInvalidMenuItem
	}

	// Retrieve the menu item from PostgreSQL.
	var menuItem models.MenuItem

	err := s.db.WithContext(ctx).
		First(&menuItem, menuItemID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartMenuItemNotFound
	}

	if err != nil {
		return nil, err
	}

	// Check whether the menu item is available.
	if !menuItem.Available {
		return nil, ErrItemUnavailable
	}

	// Add the item to the customer's cart.
 item, err := s.cartRepo.AddItem(
    ctx,
    userID,
    menuItemID,
    menuItem.RestaurantID,
    quantity,
 )

 if err != nil {
    return nil, err
}

// Include the current menu item details in the response.
item.MenuItem = menuItem

return item, nil
}