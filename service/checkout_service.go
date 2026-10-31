package service

import (
	"context"
	"errors"

	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

// DefaultDeliveryFee is a flat fee in kobo (N2,000).
// kobo is the Nigerian equivelant of cents, 100 kobo = 1 Naira
//units in kobo
const DefaultDeliveryFee int64 = 200000

var ErrEmptyCart = errors.New("your cart is empty")

//what the handler sees
type CheckoutService interface {
	Summary(ctx context.Context, userID uint) (*models.CheckoutSummary, error)
}

//holds what the service needs, a cart and a delivery fee
type checkoutService struct {
	cartRepo    repository.CartRepository
	deliveryFee int64
}

//is the constructor. It receives its dependencies from outside, which is called dependency injection
func NewCheckoutService(cartRepo repository.CartRepository, deliveryFee int64) CheckoutService {
	return &checkoutService{cartRepo: cartRepo, deliveryFee: deliveryFee}
}

// Summary builds the order review from the customer's cart. Line prices are the
// ones saved when the item was added to the cart.
func (s *checkoutService) Summary(ctx context.Context, userID uint) (*models.CheckoutSummary, error) {
	//Guard. A user ID of 0 means something went wrong upstream, such as missing auth data.
	//fails early instead of querying with bad ID
	if userID == 0 {
		return nil, ErrInvalidUser
	}

	//customer hasn't added anything

	cart, err := s.cartRepo.FindCartByUserID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEmptyCart
	}
	//Any other error is a real failure, such as the database being down, so we pass it up unchanged. The handler turns it into a 500.
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
		//make is a built-in Go function that creates a slice (a list) with a starting size and room to grow.
		//0 is the length: how many items are in the list right now. It starts empty
		Items:       make([]models.CheckoutItem, 0, len(cart.Items)),
		DeliveryFee: s.deliveryFee,
	}
	//calculcation

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