package repository

import (
	"context"
	"errors"

	"alpha-dev/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartRepository interface {
	// ACTION: CART-02 - restaurantID is passed in to enforce one-restaurant rule at repo layer
	AddItem(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		restaurantID uint,
		quantity int,
		price int64,
	) (*models.CartItem, error)
	GetCart(ctx context.Context, userID uint) (*models.Cart, error)
	ClearCart(ctx context.Context, userID uint) error
	RemoveItem(ctx context.Context, userID, menuItemID uint) error
}

const MaxCartLineQuantity = 99

type PostgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &PostgresCartRepo{db: db}
}

var (
	// ACTION: CART-02 - Error returned when cart has different restaurant
	ErrDifferentRestaurant = errors.New("cart contains items from another restaurant")
	ErrCartLineFull        = errors.New("quantity cannot exceed 99")
)

func (r *PostgresCartRepo) AddItem(
	ctx context.Context,
	userID uint,
	menuItemID uint,
	restaurantID uint,
	quantity int,
	price int64,
) (*models.CartItem, error) {

	var result models.CartItem

	// ACTION: CART-02 - Transaction + row locking to prevent concurrent mixed-restaurant adds
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// Create the customer's cart if it does not exist.
		cart := models.Cart{
			UserID: userID,
		}

		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoNothing: true,
		}).Create(&cart).Error; err != nil {
			return err
		}

		// Lock the cart to protect quantity updates.
		if err := tx.Clauses(
			clause.Locking{Strength: "UPDATE"},
		).Where("user_id = ?", userID).
			First(&cart).Error; err != nil {
			return err
		}

		// ACTION: CART-02 - Ensure the cart belongs to one restaurant only - ACCEPTANCE: reject with clear explanation
		if cart.RestaurantID != nil &&
			*cart.RestaurantID != restaurantID {
			return ErrDifferentRestaurant
		}

		// ACTION: CART-02 - Associate the cart with the restaurant on first item
		if cart.RestaurantID == nil {
			cart.RestaurantID = &restaurantID

			if err := tx.Model(&cart).
				Update("restaurant_id", restaurantID).Error; err != nil {
				return err
			}
		}

		// Check whether the item is already in the cart.
		var existing models.CartItem

		err := tx.Where(
			"cart_id = ? AND menu_item_id = ?",
			cart.ID,
			menuItemID,
		).First(&existing).Error

		if err == nil {
			next := existing.Quantity + quantity
			if next > MaxCartLineQuantity {
				return ErrCartLineFull
			}
			if err := tx.Model(&models.CartItem{}).
				Where("id = ?", existing.ID).
				Update("quantity", next).Error; err != nil {
				return err
			}
			existing.Quantity = next
			result = existing
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if quantity > MaxCartLineQuantity {
			return ErrCartLineFull
		}

		item := models.CartItem{
			CartID:     cart.ID,
			MenuItemID: menuItemID,
			Quantity:   quantity,
			Price:      price,
		}

		if err := tx.Create(&item).Error; err != nil {
			return err
		}

		result = item
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *PostgresCartRepo) GetCart(ctx context.Context, userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.WithContext(ctx).Preload("Items.MenuItem").Where("user_id = ?", userID).First(&cart).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.Cart{UserID: userID, Items: []models.CartItem{}}, nil
	}
	return &cart, err
}

func (r *PostgresCartRepo) ClearCart(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Where("cart_id = ?", cart.ID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		// ACTION: CART-02 - Reset restaurant_id when cart cleared to allow new restaurant
		cart.RestaurantID = nil
		return tx.Model(&cart).Update("restaurant_id", nil).Error
	})
}

func (r *PostgresCartRepo) RemoveItem(ctx context.Context, userID, menuItemID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart models.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			return err
		}
		if err := tx.Where("cart_id = ? AND menu_item_id = ?", cart.ID, menuItemID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		var count int64
		tx.Model(&models.CartItem{}).Where("cart_id = ?", cart.ID).Count(&count)
		if count == 0 {
			// ACTION: CART-02 - Reset restaurant_id when last item removed
			return tx.Model(&cart).Update("restaurant_id", nil).Error
		}
		return nil
	})
}
