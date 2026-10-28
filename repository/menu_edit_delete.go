package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

func (repo *PostgresMenuRepo) UpdateMenuItem(ctx context.Context, id uint, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	result := repo.db.WithContext(ctx).
		Model(&models.MenuItem{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteMenuItem is a soft delete: GORM only fills in deleted_at.
// The item disappears from normal queries, and the row stays so old orders can still point to it.
func (repo *PostgresMenuRepo) DeleteMenuItem(ctx context.Context, id uint) error {
	result := repo.db.WithContext(ctx).Delete(&models.MenuItem{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
