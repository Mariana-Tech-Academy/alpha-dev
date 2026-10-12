// repository/user.go — database lookup
package repository

import (
	"context"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) FindByID(
	ctx context.Context,
	id uint,
) (models.User, error) {
	var user models.User

	err := repo.db.WithContext(ctx).
		Select(
			"id", "first_name", "last_name", "email",
			"phone_number", "role", "status",
		).
		First(&user, id).Error

	return user, err
}
