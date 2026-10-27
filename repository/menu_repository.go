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
