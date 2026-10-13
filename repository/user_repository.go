package repository

import (
	"context"
	"errors"
	"strings"

	"alpha-dev/models"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

var ErrEmailTaken = errors.New("email is already taken")

type UserRepository interface {
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type PostgresUserRepo struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepo{db: db}
}

func (repo *PostgresUserRepo) EmailExists(
	ctx context.Context,
	email string,
) (bool, error) {
	var count int64

	// Include deleted accounts because email has a unique database index.
	err := repo.db.WithContext(ctx).
		Unscoped().
		Model(&models.User{}).
		Where("LOWER(email) = ?", email).
		Count(&count).Error

	return count > 0, err
}

func (repo *PostgresUserRepo) CreateUser(
	ctx context.Context,
	user *models.User,
) error {
	err := repo.db.WithContext(ctx).Create(user).Error
	if err == nil {
		return nil
	}

	// Catch an email duplicate even if another request saved it first.
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) &&
		postgresError.Code == "23505" &&
		strings.Contains(strings.ToLower(postgresError.ConstraintName), "email") {
		return ErrEmailTaken
	}

	return err
}
