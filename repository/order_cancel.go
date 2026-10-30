package repository

import (
	"context"
	"errors"

	"alpha-dev/models"

	"gorm.io/gorm"
)

var ErrOrderNotCancellable = errors.New("order can no longer be cancelled")

// FindOrderByID loads one order with its restaurant, so the service can check who manages it.
func (r *PostgresOrderRepo) FindOrderByID(ctx context.Context, id uint) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Preload("Restaurant").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// CancelOrder sets the order to cancelled, but only while its status still allows it.
// The status check is part of the UPDATE itself, so if a manager moves the order to
// "preparing" at the same moment, the cancel can't slip through.
func (r *PostgresOrderRepo) CancelOrder(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ? AND status IN ?", id, models.CancellableStatuses).
		Update("status", models.OrderStatusCancelled)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrderNotCancellable
	}
	return nil
}

// CountedAsSales limits an orders query to orders that count towards sales totals.
// Cancelled orders never count. Use it in every sales total, for example:
//
//	db.Model(&models.Order{}).Scopes(repository.CountedAsSales).Select("COALESCE(SUM(total), 0)").Scan(&total)
func CountedAsSales(db *gorm.DB) *gorm.DB {
	return db.Where("orders.status <> ?", models.OrderStatusCancelled)
}
