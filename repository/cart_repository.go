package repository

import (
	"context"
	"errors"

	"alpha-dev/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CartRepository interface {
	AddItem(
		ctx context.Context,
		userID uint,
		menuItemID uint,
		restaurantID uint,
		quantity int,
	) (*models.CartItem, error)
}

type PostgresCartRepo struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) CartRepository {
	return &PostgresCartRepo{db: db}
}

var ErrDifferentRestaurant = errors.New(
	"cart contains items from another restaurant",
)

func (r *PostgresCartRepo) AddItem(
	ctx context.Context,
	userID uint,
	menuItemID uint,
	restaurantID uint,
	quantity int,
) (*models.CartItem, error) {

	var result models.CartItem

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

		// Ensure the cart belongs to one restaurant.
		if cart.RestaurantID != nil &&
			*cart.RestaurantID != restaurantID {
			return ErrDifferentRestaurant
		}

		// Associate the cart with the restaurant.
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
			// Increase quantity without creating a duplicate.
			existing.Quantity += quantity

			if err := tx.Save(&existing).Error; err != nil {
				return err
			}

			result = existing
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		// Add a new item to the cart.
		item := models.CartItem{
			CartID:     cart.ID,
			MenuItemID: menuItemID,
			Quantity:   quantity,
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
