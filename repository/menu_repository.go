package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type MenuRepository interface {
	FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error)
	FindCategoryByID(ctx context.Context, id uint) (models.Category, error)
	CreateMenuItem(ctx context.Context, item *models.MenuItem) error
	FindMenuItemByID(ctx context.Context, id uint) (models.MenuItem, error)
	// This allows us to retrieve a menu item before updating it.
	// We need this because the service must check which restuarant owns the item
	UpdateMenuItemAvailability(ctx context.Context, id uint, available bool) error
	// This allows us to change the menu item's availability in PostgreSQL
	// e.g Menu Item ID: 5 | Available: false
	// The database will store the item as unavailable.
}

type PostgresMenuRepo struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

func (repo *PostgresMenuRepo) FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error) {
	var restaurant models.Restaurant
	err := repo.db.WithContext(ctx).First(&restaurant, id).Error
	return restaurant, err
}

func (repo *PostgresMenuRepo) FindCategoryByID(ctx context.Context, id uint) (models.Category, error) {
	var category models.Category
	err := repo.db.WithContext(ctx).First(&category, id).Error
	return category, err
}

func (repo *PostgresMenuRepo) CreateMenuItem(ctx context.Context, item *models.MenuItem) error {
	return repo.db.WithContext(ctx).Create(item).Error
}

// Find MenuItemByID retrieves a menu item using its ID.
func (repo *PostgresMenuRepo) FindMenuItemByID(
	ctx context.Context,
	id uint,
) (models.MenuItem, error) {

	var item models.MenuItem

	err := repo.db.WithContext(ctx).
		First(&item, id).Error

	return item, err
}

// Update MenuItemAvailability updates whether a menu item is available
func (repo *PostgresMenuRepo) UpdateMenuItemAvailability(
	ctx context.Context,
	id uint,
	available bool,
) error {

	result := repo.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Where("id = ?", id).
		Update("available", available)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil

}
