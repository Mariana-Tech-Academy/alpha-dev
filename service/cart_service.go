package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	ErrCartItemNotFound        = errors.New("cart item not found")
)

type CartService interface {
	AddItem(ctx context.Context, userID uint, menuItemID uint, quantity int) (*models.CartItem, error)
	GetCartView(ctx context.Context, userID uint) (*models.CartViewResponse, error)
	UpdateQuantity(ctx context.Context, userID uint, itemID uint, quantity int) (*models.CartViewResponse, error)
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

func emptyCartView() *models.CartViewResponse {
	return &models.CartViewResponse{Items: []models.CartItemResponse{}}
}

func (s *cartService) GetCartView(ctx context.Context, userID uint) (*models.CartViewResponse, error) {
	if userID == 0 {
		return nil, ErrInvalidUser
	}

	cart, err := s.cartRepo.FindCartByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return emptyCartView(), nil
	}
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return emptyCartView(), nil
	}
	if len(cart.Items) == 0 {
		return &models.CartViewResponse{
			CartID:       cart.ID,
			RestaurantID: cart.RestaurantID,
			Items:        []models.CartItemResponse{},
		}, nil
	}

	restaurantOpen := cart.Restaurant != nil &&
		cart.RestaurantID != nil &&
		cart.Restaurant.Status == models.RestaurantStatusActive

	items := make([]models.CartItemResponse, 0, len(cart.Items))
	var subtotal int64

	for _, item := range cart.Items {
		live := item.MenuItem.Price
		locked := item.Price
		available := restaurantOpen &&
			item.MenuItem.ID != 0 &&
			item.MenuItem.Available &&
			item.MenuItem.RestaurantID == *cart.RestaurantID

		row := models.CartItemResponse{
			ID:          item.ID,
			MenuItemID:  item.MenuItemID,
			Name:        item.MenuItem.Name,
			Price:       live,
			LockedPrice: locked,
			Quantity:    item.Quantity,
			IsAvailable: available,
		}

		var notices []string
		if !available {
			notices = append(notices, "This item is currently unavailable.")
		}
		if item.MenuItem.ID != 0 && live != locked {
			row.PriceChanged = true
			notices = append(notices, fmt.Sprintf("Price changed from %d to %d.", locked, live))
		}
		row.Notice = strings.Join(notices, " ")

		if available {
			row.LineTotal = locked * int64(item.Quantity)
			subtotal += row.LineTotal
		}
		items = append(items, row)
	}

	return &models.CartViewResponse{
		CartID:       cart.ID,
		RestaurantID: cart.RestaurantID,
		Items:        items,
		Subtotal:     subtotal,
	}, nil
}

// Accept quantities from 1 to 99. Invalid quantities leave the item unchanged.
func (s *cartService) UpdateQuantity(
	ctx context.Context,
	userID uint,
	itemID uint,
	quantity int,
) (*models.CartViewResponse, error) {
	if userID == 0 {
		return nil, ErrInvalidUser
	}
	if itemID == 0 {
		return nil, ErrCartItemNotFound
	}
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}
	if quantity > repository.MaxCartLineQuantity {
		return nil, ErrCartQuantityLimit
	}

	// Update the customer's item.
	err := s.cartRepo.UpdateQuantity(ctx, userID, itemID, quantity)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCartItemNotFound
	}
	if err != nil {
		return nil, err
	}

	// Reuse the existing calculation for the new totals.
	return s.GetCartView(ctx, userID)
}
