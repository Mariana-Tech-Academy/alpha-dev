package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"alpha-dev/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	PlaceOrder(ctx context.Context, userID uint, order *models.Order) (*models.Order, error)
	FindCustomerOrders(ctx context.Context, customerID uint, limit, offset int) ([]models.Order, int64, error)
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

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("user_id = ?", userID).
			First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEmptyCart
			}
			return err
		}

		if err := tx.Preload("MenuItem").Where("cart_id = ?", cart.ID).Find(&cart.Items).Error; err != nil {
			return err
		}
		if len(cart.Items) == 0 || cart.RestaurantID == nil {
			return ErrEmptyCart
		}

		var restaurant models.Restaurant
		if err := tx.First(&restaurant, *cart.RestaurantID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrInactiveRestaurant
			}
			return err
		}
		if restaurant.Status != models.RestaurantStatusActive {
			return ErrInactiveRestaurant
		}

		orderItems := make([]models.OrderItem, 0, len(cart.Items))
		var subtotal int64
		for _, line := range cart.Items {
			if line.MenuItem.ID == 0 || !line.MenuItem.Available || line.MenuItem.RestaurantID != *cart.RestaurantID {
				return ErrUnavailableItem
			}
			lineTotal := line.Price * int64(line.Quantity)
			subtotal += lineTotal
			orderItems = append(orderItems, models.OrderItem{
				MenuItemID: line.MenuItemID,
				Name:       line.MenuItem.Name,
				UnitPrice:  line.Price,
				Quantity:   line.Quantity,
				LineTotal:  lineTotal,
			})
		}

		var seq struct {
			OrderDate  time.Time `gorm:"column:order_date"`
			LastNumber int       `gorm:"column:last_number"`
		}
		if err := tx.Raw(`
			INSERT INTO order_daily_sequences (order_date, last_number)
			VALUES (CURRENT_DATE, 1)
			ON CONFLICT (order_date)
			DO UPDATE SET last_number = order_daily_sequences.last_number + 1
			RETURNING order_date, last_number
		`).Scan(&seq).Error; err != nil {
			return err
		}
		if seq.LastNumber < 1 {
			return errors.New("order number sequence did not advance")
		}

		order.OrderNumber = fmt.Sprintf("ORD-%s-%03d", seq.OrderDate.UTC().Format("20060102"), seq.LastNumber)
		order.CustomerID = userID
		order.RestaurantID = *cart.RestaurantID
		order.Status = models.OrderStatusPending
		order.DeliveryFee = 0
		order.Subtotal = subtotal
		order.Total = subtotal
		order.Items = orderItems

		if err := tx.Create(order).Error; err != nil {
			return err
		}
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		if err := tx.Model(&cart).Updates(map[string]any{"restaurant_id": nil}).Error; err != nil {
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

func (r *PostgresOrderRepo) FindCustomerOrders(ctx context.Context, customerID uint, limit, offset int) ([]models.Order, int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("customer_id = ?", customerID).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	orders := make([]models.Order, 0)
	err = r.db.WithContext(ctx).
		Preload("Restaurant").
		Where("customer_id = ?", customerID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}
