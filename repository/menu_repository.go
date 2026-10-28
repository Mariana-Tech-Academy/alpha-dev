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
	UpdateMenuItemAvailability(ctx context.Context, id uint, available bool) error
	UpdateMenuItem(ctx context.Context, id uint, fields map[string]any) error
	DeleteMenuItem(ctx context.Context, id uint) error
	SearchMenuItems(ctx context.Context, q MenuSearchQuery) ([]models.MenuSearchResult, error)
}

// MenuSearchQuery is a normalized menu search. Query is already escaped for ILIKE.
type MenuSearchQuery struct {
	Query        string
	CategoryID   *uint
	RestaurantID *uint
	MinPrice     *int64
	MaxPrice     *int64
	Available    bool
	Limit        int
	Offset       int
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

func (repo *PostgresMenuRepo) FindMenuItemByID(ctx context.Context, id uint) (models.MenuItem, error) {
	var item models.MenuItem
	err := repo.db.WithContext(ctx).
		Preload("Restaurant").
		Preload("Category").
		First(&item, id).Error
	return item, err
}

func (repo *PostgresMenuRepo) UpdateMenuItemAvailability(ctx context.Context, id uint, available bool) error {
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

func (repo *PostgresMenuRepo) SearchMenuItems(ctx context.Context, q MenuSearchQuery) ([]models.MenuSearchResult, error) {
	items := make([]models.MenuSearchResult, 0)
	query := repo.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Select("menu_items.id, menu_items.name, menu_items.description, menu_items.price, menu_items.image_url, categories.name AS category, menu_items.restaurant_id, restaurants.name AS restaurant_name").
		Joins("JOIN restaurants ON restaurants.id = menu_items.restaurant_id AND restaurants.deleted_at IS NULL").
		Joins("LEFT JOIN categories ON categories.id = menu_items.category_id AND categories.deleted_at IS NULL").
		Where("restaurants.status = ?", models.RestaurantStatusActive).
		Where("menu_items.available = ?", q.Available)
	if q.Query != "" {
		query = query.Where("menu_items.name ILIKE ?", "%"+q.Query+"%")
	}
	if q.CategoryID != nil {
		query = query.Where("menu_items.category_id = ?", *q.CategoryID)
	}
	if q.RestaurantID != nil {
		query = query.Where("menu_items.restaurant_id = ?", *q.RestaurantID)
	}
	if q.MinPrice != nil {
		query = query.Where("menu_items.price >= ?", *q.MinPrice)
	}
	if q.MaxPrice != nil {
		query = query.Where("menu_items.price <= ?", *q.MaxPrice)
	}
	err := query.Order("menu_items.name ASC, menu_items.id ASC").
		Limit(q.Limit).
		Offset(q.Offset).
		Scan(&items).Error
	return items, err
}
