package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

// Checkout must create the order, create its items and clear the cart inside one
// r.db.Transaction(func(tx *gorm.DB) error { ... }) call.


type OrderRepository interface {
	GetByID(ctx context.Context, id uint) (models.Order, error)
}

type PostgresOrderRepo struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &PostgresOrderRepo{db: db}
}

// GetByID retrieves one order and its associated items.
func (r *PostgresOrderRepo) GetByID(
	ctx context.Context,
	id uint,
) (models.Order, error) {
	var order models.Order

	err := r.db.WithContext(ctx).
		Preload("Items").
		First(&order, id).Error

	return order, err
}