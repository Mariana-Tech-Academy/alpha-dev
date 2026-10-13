package service

import (
	"context"
	"errors"

	"alpha-dev/auth"
	"alpha-dev/models"
	"alpha-dev/repository"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInactiveUser = errors.New("account is not active")
)

type UserService interface {
	GetActiveUser(ctx context.Context, id uint) (models.User, error)
}

type userService struct {
	userRepo repository.UserRepository
	tokens   *auth.JWTManager
}

func NewUserService(repo repository.UserRepository, tokens *auth.JWTManager) UserService {
	return &userService{userRepo: repo, tokens: tokens}
}

func (s *userService) GetActiveUser(ctx context.Context, id uint) (models.User, error) {
	if id == 0 {
		return models.User{}, ErrUserNotFound
	}

	user, err := s.userRepo.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, ErrUserNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	if user.Status != models.UserStatusActive {
		return models.User{}, ErrInactiveUser
	}

	return user, nil
}
