package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type RestaurantRepository interface {
	Create(ctx context.Context, restaurant *models.Restaurant) error
	FindByID(ctx context.Context, id uint) (models.Restaurant, error)
	FindActiveByID(ctx context.Context, id uint) (models.Restaurant, error)
	UpdateFields(ctx context.Context, id uint, fields map[string]any) error
	UpdateAvailability(
		ctx context.Context,
		id uint,
		managerID uint,
		currentStatus models.RestaurantStatus,
		nextStatus models.RestaurantStatus,
	) (bool, error)
	GetActiveRestaurants(ctx context.Context, page, limit int) ([]models.Restaurant, error)
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

func (repo *PostgresRestaurantRepo) FindActiveByID(ctx context.Context, id uint) (models.Restaurant, error) {
	var restaurant models.Restaurant
	err := repo.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Preload("MenuItems", "available = ?", true).
		Preload("MenuItems.Category").
		First(&restaurant, id).Error
	return restaurant, err
}

func (repo *PostgresRestaurantRepo) UpdateFields(ctx context.Context, id uint, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return repo.db.WithContext(ctx).Model(&models.Restaurant{}).Where("id = ?", id).Updates(fields).Error
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

func (repo *PostgresRestaurantRepo) GetActiveRestaurants(ctx context.Context, page, limit int) ([]models.Restaurant, error) {
	var restaurants []models.Restaurant
	offset := (page - 1) * limit
	err := repo.db.WithContext(ctx).
		Where("status = ?", models.RestaurantStatusActive).
		Order("id").
		Limit(limit).
		Offset(offset).
		Find(&restaurants).Error
	return restaurants, err
}
