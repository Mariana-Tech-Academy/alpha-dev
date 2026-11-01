package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

type CheckoutService interface {
	Summary(ctx context.Context, userID uint) (*models.CheckoutSummary, error)
}

type checkoutService struct {
	cartRepo    repository.CartRepository
	deliveryFee int64
}

func NewCheckoutService(cartRepo repository.CartRepository, deliveryFee int64) CheckoutService {
	return &checkoutService{cartRepo: cartRepo, deliveryFee: deliveryFee}
}

func (s *checkoutService) Summary(ctx context.Context, userID uint) (*models.CheckoutSummary, error) {
	if userID == 0 {
		return nil, ErrInvalidUser
	}

	cart, err := s.cartRepo.FindCartByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmptyCart
	}
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 || cart.RestaurantID == nil || cart.Restaurant == nil {
		return nil, ErrEmptyCart
	}
	if cart.Restaurant.Status != models.RestaurantStatusActive {
		return nil, ErrCheckoutRestaurantInactive
	}

	summary := &models.CheckoutSummary{
		Restaurant: models.CheckoutRestaurant{
			ID:   cart.Restaurant.ID,
			Name: cart.Restaurant.Name,
		},
		Items:       make([]models.CheckoutItem, 0, len(cart.Items)),
		DeliveryFee: s.deliveryFee,
	}
	for _, item := range cart.Items {
		if item.MenuItem.ID == 0 || !item.MenuItem.Available || item.MenuItem.RestaurantID != *cart.RestaurantID {
			return nil, ErrCheckoutItemUnavailable
		}
		lineTotal := item.Price * int64(item.Quantity)
		summary.Items = append(summary.Items, models.CheckoutItem{
			MenuItemID: item.MenuItemID,
			Name:       item.MenuItem.Name,
			Quantity:   item.Quantity,
			UnitPrice:  item.Price,
			LineTotal:  lineTotal,
		})
		summary.Subtotal += lineTotal
	}
	summary.Total = summary.Subtotal + summary.DeliveryFee
	return summary, nil
}
