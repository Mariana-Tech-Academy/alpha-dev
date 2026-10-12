// service/auth.go - user checks

package service

import (
	"context"
	"errors"

	"github.com/Mariana-Tech-Academy/alpha-dev/models"
	"github.com/Mariana-Tech-Academy/alpha-dev/repository"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInactiveUser = errors.New("account is not active")
)

type AuthService struct {
	users *repository.UserRepository
}

func NewAuthService(users *repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

func (s *AuthService) GetActiveUser(
	ctx context.Context,
	id uint,
) (models.User, error) {
	user, err := s.users.FindByID(ctx, id)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if string(user.Status) != "active" {
		return models.User{}, ErrInactiveUser
	}

	return user, nil
}
