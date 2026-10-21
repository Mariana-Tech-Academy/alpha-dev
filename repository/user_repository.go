package repository

import (
	"context"

	"alpha-dev/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (models.User, error)
	FindByEmail(ctx context.Context, email string) (models.User, error)
	UpdatePassword(ctx context.Context, id uint, hash string) error
	Update(ctx context.Context, user *models.User) error
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

// EmailExists expects a lowercased email and includes soft-deleted users,
// since the unique index on email still covers them.
func (repo *PostgresUserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	err := repo.db.WithContext(ctx).
		Unscoped().
		Model(&models.User{}).
		Where("email = ?", email).
		Count(&count).Error
	return count > 0, err
}

// CreateUser returns gorm.ErrDuplicatedKey if the email is already taken.
func (repo *PostgresUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	return repo.db.WithContext(ctx).Create(user).Error
}

func (repo *PostgresUserRepo) FindByID(ctx context.Context, id uint) (models.User, error) {
	var user models.User
	err := repo.db.WithContext(ctx).First(&user, id).Error
	return user, err
}

// FindByEmail expects a lowercased email.
func (repo *PostgresUserRepo) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := repo.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (repo *PostgresUserRepo) UpdatePassword(ctx context.Context, id uint, hash string) error {
	return repo.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", id).
		Update("password", hash).Error
}

func (repo *PostgresUserRepo) Update(ctx context.Context, user *models.User) error {
	return repo.db.WithContext(ctx).Save(user).Error
}
