package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByID(ctx context.Context, id uint) (models.User, error)
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) FindByID(ctx context.Context, id uint) (models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, err
}
