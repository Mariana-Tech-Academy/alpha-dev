package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// List the database operations used by the menu service.
type MenuRepository interface {
	FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error)
	FindCategoryByID(ctx context.Context, id uint) (models.Category, error)
	CreateMenuItem(ctx context.Context, item *models.MenuItem) error
	FindMenuByRestaurant(ctx context.Context, restaurantID uint, categoryID uint) ([]models.MenuItem, error)
}

type PostgresMenuRepo struct {
	db *gorm.DB
}

// Connect the repository to the database.
func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &PostgresMenuRepo{db: db}
}

// Find a restaurant using its ID.
func (repo *PostgresMenuRepo) FindRestaurantByID(ctx context.Context, id uint) (models.Restaurant, error) {
	var restaurant models.Restaurant
	err := repo.db.WithContext(ctx).First(&restaurant, id).Error
	return restaurant, err
}

// Find a category using its ID.
func (repo *PostgresMenuRepo) FindCategoryByID(ctx context.Context, id uint) (models.Category, error) {
	var category models.Category
	err := repo.db.WithContext(ctx).First(&category, id).Error
	return category, err
}

// Save a new menu item.
func (repo *PostgresMenuRepo) CreateMenuItem(ctx context.Context, item *models.MenuItem) error {
	return repo.db.WithContext(ctx).Create(item).Error
}

// Read one restaurant's menu and include its category details.
func (repo *PostgresMenuRepo) FindMenuByRestaurant(ctx context.Context, restaurantID uint, categoryID uint) ([]models.MenuItem, error) {
	items := make([]models.MenuItem, 0)

	query := repo.db.WithContext(ctx).
		Preload("Category").
		Where("restaurant_id = ?", restaurantID)

	// Filter by category when one has been selected.
	if categoryID != 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	// Get the matching items in ID order.
	err := query.Order("id ASC").Find(&items).Error
	return items, err
}
