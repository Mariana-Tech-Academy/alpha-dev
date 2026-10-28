package repository

import (
	"alpha-dev/models"
	"context"
	"gorm.io/gorm"
)

type RestaurantRepository interface {
	FindActiveByID(ctx context.Context, id uint) (models.Restaurant, error)
	SearchActiveByNameOrLocation(ctx context.Context, query string) ([]models.Restaurant, error)
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

// SearchActiveByNameOrLocation searches for active restaurants by name or location (city/state).
func (repo *PostgresRestaurantRepo) SearchActiveByNameOrLocation(ctx context.Context, pattern string) ([]models.Restaurant, error) {
	// Initialize an empty slice to hold the results
	var restaurants []models.Restaurant = []models.Restaurant{}
	// Use a case-insensitive search for name or location (city/state) and only return active restaurants.
	err := repo.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Where("LOWER(name) LIKE LOWER(?) OR LOWER(location) LIKE LOWER(?)", pattern, pattern).
		Find(&restaurants).Error
	// Return an empty slice if no records are found, instead of returning an error.
	if err != nil {
		return []models.Restaurant{}, err
	}

	return restaurants, nil
}