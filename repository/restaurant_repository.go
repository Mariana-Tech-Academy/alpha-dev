package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *models.Restaurant) error
	FindByID(ctx context.Context, id uint) (models.Restaurant, error)
	Update(ctx context.Context, restaurant *models.Restaurant) error
	UpdateAvailability(
		ctx context.Context,
		id uint,
		managerID uint,
		currentStatus models.RestaurantStatus,
		nextStatus models.RestaurantStatus,
	) (bool, error)
}

type PostgresRestaurantRepo struct {
	db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
	return &PostgresRestaurantRepo{db: db}
}

func (repo *PostgresRestaurantRepo) Create(ctx context.Context, restaurant *models.Restaurant) error {
	return repo.db.WithContext(ctx).Create(restaurant).Error
}

func (repo *PostgresRestaurantRepo) FindByID(ctx context.Context, id uint) (models.Restaurant, error) {
	var restaurant models.Restaurant
	err := repo.db.WithContext(ctx).First(&restaurant, id).Error
	return restaurant, err
}

func (repo *PostgresRestaurantRepo) Update(ctx context.Context, restaurant *models.Restaurant) error {
	return repo.db.WithContext(ctx).Save(restaurant).Error
}

// UpdateAvailability changes status only if the owner and previous status still match.
func (repo *PostgresRestaurantRepo) UpdateAvailability(
	ctx context.Context,
	id uint,
	managerID uint,
	currentStatus models.RestaurantStatus,
	nextStatus models.RestaurantStatus,
) (bool, error) {
	result := repo.db.WithContext(ctx).
		Model(&models.Restaurant{}).
		Where("id = ? AND manager_id = ? AND status = ?", id, managerID, currentStatus).
		Where("status IN ?", []models.RestaurantStatus{
			models.RestaurantStatusActive,
			models.RestaurantStatusInactive,
		}).
		Update("status", nextStatus)

	return result.RowsAffected == 1, result.Error
}
