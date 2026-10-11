package repository

import (
	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"gorm.io/gorm"
)

type UserRepository interface {
	GetUserByID(userID uint) (*models.User, error)
	UpdateUser(user *models.User) error
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

// GetUserByID retrieves a user by their ID from the database.
func (r *PostgresUserRepo) GetUserByID(userID uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates the user's information in the database.
func (r *PostgresUserRepo) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}