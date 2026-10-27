package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.
type OrderRepository interface {
	PlaceOrder(
		ctx context.Context,
		userID uint,
		order *models.Order,
	) (*models.Order, error)
}

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

var (
	ErrEmptyCart          = errors.New("your cart is empty")
	ErrInactiveRestaurant = errors.New("restaurant is not active")
	ErrUnavailableItem    = errors.New("one or more menu items are unavailable")
)

func (r *PostgresOrderRepo) PlaceOrder(
	ctx context.Context,
	userID uint,
	order *models.Order,
) (*models.Order, error) {

	var result *models.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Step 1: Retrieve the customer's cart.
		var cart models.Cart

		if err := tx.
			Preload("Items.MenuItem").
			Where("user_id = ?", userID).
			First(&cart).Error; err != nil {

			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEmptyCart
			}

			return err
		}

		// Step 2: Check that the cart contains items.
		if len(cart.Items) == 0 {
			return ErrEmptyCart
		}

		// Step 3: Check that the restaurant is active.
		if cart.RestaurantID == nil {
			return ErrEmptyCart
		}

		var restaurant models.Restaurant

		if err := tx.First(
			&restaurant,
			*cart.RestaurantID,
		).Error; err != nil {
			return err
		}

		if restaurant.Status != models.RestaurantStatusActive {
			return ErrInactiveRestaurant
		}

		// Step 4: Check menu availability and calculate subtotal.
		var subtotal int64

		for _, cartItem := range cart.Items {
			menuItem := cartItem.MenuItem

			// Reject unavailable items or items from another restaurant.
			if !menuItem.Available ||
				menuItem.RestaurantID != *cart.RestaurantID {
				return ErrUnavailableItem
			}

			// Calculate the cost of each item.
			lineTotal := menuItem.Price * int64(cartItem.Quantity)

			// Add the item cost to the subtotal.
			subtotal += lineTotal
		}
		// Step 5A: Prepare the order items.
		orderItems := make([]models.OrderItem, 0, len(cart.Items))

		for _, cartItem := range cart.Items {
			menuItem := cartItem.MenuItem

			orderItems = append(orderItems, models.OrderItem{
				MenuItemID: cartItem.MenuItemID,
				Name:       menuItem.Name,
				UnitPrice:  menuItem.Price,
				Quantity:   cartItem.Quantity,
				LineTotal:  menuItem.Price * int64(cartItem.Quantity),
			})
		}
		// Step 5B: Prepare the order for checkout.
		// Generate a unique order reference
		randomBytes := make([]byte, 8)

		if _, err := rand.Read(randomBytes); err != nil {
			return err
		}

		order.OrderNumber = "ORD-" + hex.EncodeToString(randomBytes)
		order.CustomerID = userID
		order.RestaurantID = *cart.RestaurantID
		order.Status = models.OrderStatusPending
		order.Subtotal = subtotal
		order.Total = subtotal + order.DeliveryFee
		order.Items = orderItems

		// Step 5C: Save the order and its items to PostgreSQL
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Step 5D: Clear the customer's cart after checkout.
		if err := tx.Where("cart_id = ?", cart.ID).
			Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		if err := tx.Model(&cart).
			Update("restaurant_id", nil).Error; err != nil {
			return err
		}
		result = order

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
