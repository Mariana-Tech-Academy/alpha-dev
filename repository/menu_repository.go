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
