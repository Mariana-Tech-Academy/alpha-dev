
package repository

import (
    "context"

    "alpha-dev/models"

    "gorm.io/gorm"
)

type RestaurantRepository interface {
    Create(ctx context.Context, restaurant *models.Restaurant) error
}

type PostgresRestaurantRepo struct {
    db *gorm.DB
}

func NewRestaurantRepository(db *gorm.DB) RestaurantRepository {
    return &PostgresRestaurantRepo{db: db}
}

func (r *PostgresRestaurantRepo) Create(
    ctx context.Context,
    restaurant *models.Restaurant,
) error {
    return r.db.WithContext(ctx).Create(restaurant).Error
}
