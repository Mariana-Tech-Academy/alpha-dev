package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type CategoryRepository interface {
	Create(ctx context.Context, cat *models.Category) error
	FindAll(ctx context.Context, page, limit int) ([]models.Category, int64, error)
	FindByID(ctx context.Context, id uint) (models.Category, error)
	FindPublicMenuItems(ctx context.Context, categoryID uint) ([]models.CategoryMenuItem, error)
}

type PostgresCategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &PostgresCategoryRepo{db: db}
}

func (r *PostgresCategoryRepo) Create(ctx context.Context, cat *models.Category) error {
	return r.db.WithContext(ctx).Create(cat).Error
}

func (r *PostgresCategoryRepo) FindAll(ctx context.Context, page, limit int) ([]models.Category, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&models.Category{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	cats := make([]models.Category, 0)
	err := r.db.WithContext(ctx).
		Order("name ASC, id ASC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&cats).Error
	return cats, total, err
}

func (r *PostgresCategoryRepo) FindByID(ctx context.Context, id uint) (models.Category, error) {
	var cat models.Category
	err := r.db.WithContext(ctx).First(&cat, id).Error
	return cat, err
}

func (r *PostgresCategoryRepo) FindPublicMenuItems(ctx context.Context, categoryID uint) ([]models.CategoryMenuItem, error) {
	items := make([]models.CategoryMenuItem, 0)
	err := r.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Select("menu_items.id, menu_items.name, menu_items.description, menu_items.price, menu_items.image_url, menu_items.restaurant_id, restaurants.name AS restaurant_name").
		Joins("JOIN restaurants ON restaurants.id = menu_items.restaurant_id AND restaurants.deleted_at IS NULL").
		Where("menu_items.category_id = ?", categoryID).
		Where("menu_items.available = ?", true).
		Where("restaurants.status = ?", models.RestaurantStatusActive).
		Order("menu_items.name ASC, menu_items.id ASC").
		Scan(&items).Error
	return items, err
}
