package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type OrderRepository interface {
	PlaceOrder(ctx context.Context, userID uint, order *models.Order) (*models.Order, error)
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

func (r *PostgresOrderRepo) PlaceOrder(ctx context.Context, userID uint, order *models.Order) (*models.Order, error) {
	var result *models.Order

	// Wrap all in transaction for atomicity
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Get cart with menu items
		var cart models.Cart
		if err := tx.Preload("Items.MenuItem").Where("user_id = ?", userID).First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEmptyCart
			}
			return err
		}
		if len(cart.Items) == 0 || cart.RestaurantID == nil {
			return ErrEmptyCart
		}

		// Step 2: Validate restaurant
		var restaurant models.Restaurant
		if err := tx.First(&restaurant, *cart.RestaurantID).Error; err != nil {
			return err
		}
		if restaurant.Status != models.RestaurantStatusActive {
			return ErrInactiveRestaurant
		}

		// Step 3: Validate items and calculate subtotal
		var subtotal int64
		for _, ci := range cart.Items {
			if !ci.MenuItem.Available || ci.MenuItem.RestaurantID != *cart.RestaurantID {
				return ErrUnavailableItem
			}
			subtotal += ci.MenuItem.Price * int64(ci.Quantity)
		}

		// Step 4: Build order item snapshots (so past orders never change)
		orderItems := make([]models.OrderItem, 0, len(cart.Items))
		for _, ci := range cart.Items {
			orderItems = append(orderItems, models.OrderItem{
				MenuItemID: ci.MenuItemID,
				Name:       ci.MenuItem.Name,
				UnitPrice:  ci.MenuItem.Price,
				Quantity:   ci.Quantity,
				LineTotal:  ci.MenuItem.Price * int64(ci.Quantity),
			})
		}

		// Step 5: ORD-01 CORE - Generate unique ORD-YYYYMMDD-NNN atomically
		// This is concurrency-safe: Postgres locks the row for order_date
		var seq struct {
			LastNumber int `gorm:"column:last_number"`
		}
		if err := tx.Raw(`
			INSERT INTO order_daily_sequences (order_date, last_number)
			VALUES (CURRENT_DATE, 1)
			ON CONFLICT (order_date) 
			DO UPDATE SET last_number = order_daily_sequences.last_number + 1
			RETURNING last_number
		`).Scan(&seq).Error; err != nil {
			return fmt.Errorf("failed to generate order number: %w", err)
		}

		// Format: ORD-YYYYMMDD-NNN -> e.g. ORD-20261028-001
		order.OrderNumber = fmt.Sprintf("ORD-%s-%03d", time.Now().Format("20060102"), seq.LastNumber)

		// Step 6: Fill remaining order fields
		order.CustomerID = userID
		order.RestaurantID = *cart.RestaurantID
		order.Status = models.OrderStatusPending
		order.Subtotal = subtotal
		order.Total = subtotal + order.DeliveryFee
		order.Items = orderItems

		// Step 7: Save order to DB (uniqueIndex will catch any duplicate)
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Step 8: Clear cart after successful order
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&cart).Update("restaurant_id", nil).Error; err != nil {
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
