package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	PlaceOrder(
		ctx context.Context,
		userID uint,
		order *models.Order,
	) (*models.Order, error)

	FindCustomerOrders(
		ctx context.Context,
		customerID uint,
		limit int,
		offset int,
	) ([]models.Order, int64, error)
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

// Read one page of the customer's orders, newest first.
func (r *PostgresOrderRepo) FindCustomerOrders(
	ctx context.Context,
	customerID uint,
	limit int,
	offset int,
) ([]models.Order, int64, error) {
	orders := make([]models.Order, 0)
	var total int64

	query := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("customer_id = ?", customerID)

	// Count the customer's orders before applying pagination.
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Include the restaurant name and keep the order consistent.
	err := query.
		Preload("Restaurant").
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// Keep checkout and cart clearing inside one transaction.
func (r *PostgresOrderRepo) PlaceOrder(
	ctx context.Context,
	userID uint,
	order *models.Order,
) (*models.Order, error) {
	var result *models.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Retrieve the customer's cart.
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

		if len(cart.Items) == 0 || cart.RestaurantID == nil {
			return ErrEmptyCart
		}

		// Check that the restaurant is active.
		var restaurant models.Restaurant
		if err := tx.First(&restaurant, *cart.RestaurantID).Error; err != nil {
			return err
		}
		if restaurant.Status != models.RestaurantStatusActive {
			return ErrInactiveRestaurant
		}

		// Check availability and calculate the subtotal.
		var subtotal int64

		for _, cartItem := range cart.Items {
			menuItem := cartItem.MenuItem

			if !menuItem.Available ||
				menuItem.RestaurantID != *cart.RestaurantID {
				return ErrUnavailableItem
			}

			subtotal += menuItem.Price * int64(cartItem.Quantity)
		}

		// Copy the purchased items into the order.
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

		// Prepare the order using the existing checkout rules.
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

		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Clear the cart after saving the order.
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
