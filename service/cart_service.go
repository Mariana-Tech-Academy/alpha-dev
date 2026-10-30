package service

import (
	"context"
	"errors"
	"fmt"
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
	GetCartView(ctx context.Context, userID uint) (*models.CartViewResponse, error)
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

func (s *cartService) GetCartView(ctx context.Context, userID uint) (*models.CartViewResponse, error) {
	if userID == 0 {
		return nil, errors.New("invalid user ID")
	}

	// Call the repository to fetch the cart for the user
	cart, err := s.cartRepo.FindCartByUserID(ctx, userID)
	if err != nil {
		// If the cart is not found, return an empty cart view response instead of an error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.CartViewResponse{
				CartID:       0,
				RestaurantID: nil,
				Items:        []models.CartItemResponse{},
				Subtotal:     0.0,
			}, nil
		}
		// For other errors, return the error
		return nil, err
	}

	// If the user doesn't have an active cart yet, return an empty response cleanly
	if cart == nil {
		return &models.CartViewResponse{
			Items:    []models.CartItemResponse{},
			Subtotal: 0,
		}, nil
	}

	var responseItems []models.CartItemResponse
	var subtotal float64 = 0
	// Iterate through the cart items to build the response
	for _, item := range cart.Items {
		menuItem := item.MenuItem

		currentPrice := float64(menuItem.Price) / 100.0 // Convert from minor units to major units (e.g., pence to pounds)
		lockedPrice := float64(item.Price) / 100.0
		isAvailable := menuItem.Available
		// Acceptance Criteria: Build CartItemResponse
		itemResp := models.CartItemResponse{
			MenuItemID:   item.MenuItemID,
			Name:         menuItem.Name,
			Price:        currentPrice,
			LockedPrice:  lockedPrice,
			Quantity:     item.Quantity,
			IsAvailable:  isAvailable,
			PriceChanged: false,
		}

		//  Availability notice
		if !isAvailable {
			itemResp.Notice = "This item is currently unavailable."
		}

		// Price change notice
		if lockedPrice != currentPrice {
			itemResp.PriceChanged = true
			priceNotice := fmt.Sprintf("Price changed from £%.2f to £%.2f.", lockedPrice, currentPrice)
			if itemResp.Notice != "" {
				itemResp.Notice += " " + priceNotice
			} else {
				itemResp.Notice = priceNotice
			}
		}

		// Calculate line total and add to subtotal (only if available)
		if isAvailable {
			lineTotal := currentPrice * float64(item.Quantity)
			itemResp.LineTotal = lineTotal
			subtotal += lineTotal
		} else {
			itemResp.LineTotal = 0.0
		}
		// Append the item response to the list
		responseItems = append(responseItems, itemResp)
	}
	// Build the final CartViewResponse
	return &models.CartViewResponse{
		CartID:       cart.ID,
		RestaurantID: cart.RestaurantID,
		Items:        responseItems,
		Subtotal:     subtotal,
	}, nil
}