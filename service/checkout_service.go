package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

// DefaultDeliveryFee is a flat fee in kobo (N2,000).
const DefaultDeliveryFee int64 = 200000

var ErrEmptyCart = errors.New("your cart is empty")

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

// Summary builds the order review from the customer's cart. Line prices are the
// ones saved when the item was added to the cart.
func (s *checkoutService) Summary(ctx context.Context, userID uint) (*models.CheckoutSummary, error) {
	if userID == 0 {
		return nil, ErrInvalidUser
	}

	//customer hasn't added anything

	cart, err := s.cartRepo.FindCartByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmptyCart
	}
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, ErrEmptyCart
	}
	//if the restaurant can't be found
	if cart.Restaurant == nil {
		return nil, ErrItemUnavailable
	}

	//building the response

	summary := &models.CheckoutSummary{
		Restaurant: models.CheckoutRestaurant{
			ID:   cart.Restaurant.ID,
			Name: cart.Restaurant.Name,
		},
		Items:       make([]models.CheckoutItem, 0, len(cart.Items)),
		DeliveryFee: s.deliveryFee,
	}

	for _, item := range cart.Items {
		//converting item.Quantity cause Go won't multiply different different number types
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