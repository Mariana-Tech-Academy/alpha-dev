package service

import (
	"errors"
	"github.com/Mariana-Tech-Academy/alpha-dev/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	ChangePassword(userID uint, currentPassword, newPassword string) error
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{userRepo: repo}
}

func (s *userService) ChangePassword(userID uint, currentPassword, newPassword string) error {
	//Get user by userID
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Compare old password with stored PasswordHash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return errors.New("incorrect current password")
	}

	// Hash new password securely
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update user's password hash in the database
	user.PasswordHash = string(hashedBytes)
	return s.userRepo.UpdateUser(user)
}