package repository

import (
	"alpha-dev/models"
	"context"
	"gorm.io/gorm"
)

type RestaurantRepository interface {
	FindActiveByID(ctx context.Context, id uint) (models.Restaurant, error)
}

type PostgresRestaurantRepo struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}

// FindActiveByID retrieves an active restaurant by its ID, including its menu items.
func (repo *PostgresRestaurantRepo) FindActiveByID(ctx context.Context, id uint) (models.Restaurant, error) {
	var restaurant models.Restaurant

	err := repo.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, models.RestaurantStatusActive).
		Preload("Manager").
		Preload("MenuItems").
		Preload("MenuItems.Category").
		First(&restaurant).Error

	return restaurant, err
}
